package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/paperbanana/paperbanana/internal/api"
	criticagent "github.com/paperbanana/paperbanana/internal/application/agents/critic"
	planneragent "github.com/paperbanana/paperbanana/internal/application/agents/planner"
	retrieveragent "github.com/paperbanana/paperbanana/internal/application/agents/retriever"
	stylistagent "github.com/paperbanana/paperbanana/internal/application/agents/stylist"
	visualizeragent "github.com/paperbanana/paperbanana/internal/application/agents/visualizer"
	configservice "github.com/paperbanana/paperbanana/internal/application/config"
	"github.com/paperbanana/paperbanana/internal/application/orchestrator"
	"github.com/paperbanana/paperbanana/internal/application/persistence"
	"github.com/paperbanana/paperbanana/internal/config"
	domainagent "github.com/paperbanana/paperbanana/internal/domain/agent"
	domainllm "github.com/paperbanana/paperbanana/internal/domain/llm"
	rediscache "github.com/paperbanana/paperbanana/internal/infrastructure/cache/redis"
	"github.com/paperbanana/paperbanana/internal/infrastructure/crypto/aesgcm"
	llminfra "github.com/paperbanana/paperbanana/internal/infrastructure/llm"
	"github.com/paperbanana/paperbanana/internal/infrastructure/nodes/httpnode"
	"github.com/paperbanana/paperbanana/internal/infrastructure/persistence/sqlite"
	"github.com/paperbanana/paperbanana/internal/infrastructure/resilience"
	goredis "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("failed to create logger: %v", err)
	}
	defer func() {
		_ = logger.Sync()
	}()

	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("failed to load config", zap.Error(err))
	}

	providerConfig, ok := cfg.LLM.Providers[cfg.LLM.Default]
	if !ok {
		logger.Fatal("default provider not configured", zap.String("provider", cfg.LLM.Default))
	}

	resilientClient := resilience.NewResilientClient("llm-"+cfg.LLM.Default, providerConfig.Timeout)
	options := llminfra.ClientOptions{HTTPClient: resilientClient.HTTPClient()}
	if cfg.Cache.Redis.Enabled {
		redisClient := goredis.NewClient(&goredis.Options{
			Addr:     cfg.Cache.Redis.Addr,
			Password: cfg.Cache.Redis.Password,
			DB:       cfg.Cache.Redis.DB,
		})
		options.Cache = rediscache.NewCache(rediscache.NewStore(redisClient))
	}

	// Bootstrap SQLite persistence
	bootstrapResult, err := sqlite.Bootstrap(context.Background(), sqlite.BootstrapConfig{
		DatabasePath:      cfg.Persistence.DatabasePath,
		EnableForeignKeys: cfg.Persistence.EnableForeignKeys,
		BusyTimeoutMs:     cfg.Persistence.BusyTimeoutMs,
		EnableWAL:         cfg.Persistence.EnableWAL,
	})
	if err != nil {
		logger.Fatal("failed to bootstrap persistence", zap.Error(err))
	}
	db := bootstrapResult.DB

	// Build repositories and services
	txManager := sqlite.NewTxManager(db)
	workspaceService := persistence.NewWorkspaceService(txManager)
	historyService := persistence.NewHistoryService(txManager)

	// Build asset store and service
	assetRepo := sqlite.NewAssetRepository(db)
	assetStore := sqlite.NewLocalAssetStore(cfg.Assets.Root)
	assetService := persistence.NewAssetService(assetRepo, assetStore)

	// Build the persistence-backed snapshot store
	sessionRepo := sqlite.NewSessionRepository(db)
	snapshotStore := sqlite.NewPersistentSnapshotStore(sessionRepo)

	// Build encryption service for API keys
	encryptionService, err := aesgcm.NewService()
	if err != nil {
		logger.Fatal("failed to create encryption service", zap.Error(err))
	}

	// Build provider and API key repositories
	providerRepo := sqlite.NewProviderRepository(db)
	apiKeyRepo := sqlite.NewAPIKeyRepository(db, encryptionService)

	// Initialize system providers (idempotent - only creates if not exists)
	if err := providerRepo.InitializeSystemProviders(); err != nil {
		logger.Warn("failed to initialize system providers", zap.Error(err))
	}

	// Build config service with watcher for hot reload
	configWatcher := configservice.NewWatcher()
	configSvc := configservice.NewServiceWithWatcher(providerRepo, apiKeyRepo, configWatcher)

	// Load node catalog for visualizer
	nodeCatalog, err := loadNodeCatalog(logger)
	if err != nil {
		logger.Fatal("failed to load node catalog", zap.Error(err))
	}

	queryClient := llminfra.NewRuntimeClient(
		llminfra.RuntimePurposeQuery,
		cfg.LLM.Default,
		providerConfig,
		options,
		providerRepo,
		apiKeyRepo,
	)
	genClient := llminfra.NewRuntimeClient(
		llminfra.RuntimePurposeGen,
		cfg.LLM.Default,
		providerConfig,
		options,
		providerRepo,
		apiKeyRepo,
	)

	runner, err := buildRunner(providerConfig, queryClient, genClient, snapshotStore, nodeCatalog)
	if err != nil {
		logger.Fatal("failed to build runner", zap.Error(err))
	}

	// Create agent factory for batch processing
	agentFactory := &agentFactory{
		queryClient:    queryClient,
		genClient:      genClient,
		providerConfig: providerConfig,
		nodeCatalog:    nodeCatalog,
		snapshotStore:  snapshotStore,
	}

	// Wire up the full router with persistence and config endpoints
	router := api.SetupRouterWithConfigAndBatch(runner, api.PersistenceServices{
		WorkspaceService: workspaceService,
		HistoryService:   historyService,
		AssetService:     assetService,
	}, &api.ConfigServices{
		ConfigService: configSvc,
	}, &api.BatchServices{
		BatchRunner:  orchestrator.NewBatchRunner(agentFactory),
		AgentFactory: agentFactory,
	}, nil, logger)

	address := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)

	logger.Info("starting server",
		zap.String("address", address),
		zap.String("provider", cfg.LLM.Default),
		zap.String("model", providerConfig.Model),
		zap.String("database", cfg.Persistence.DatabasePath),
	)

	if err := router.Run(address); err != nil {
		logger.Fatal("server stopped", zap.Error(err))
	}
}

func buildRunner(
	providerConfig config.ProviderConfig,
	queryClient domainllm.LLMClient,
	genClient domainllm.LLMClient,
	snapshotStore orchestrator.SnapshotStore,
	nodeCatalog *config.NodeCatalog,
) (*orchestrator.Runner, error) {
	visualizerAgent := visualizeragent.NewAgent(genClient, visualizeragent.Config{
		Model:       providerConfig.Model,
		NodeCatalog: nodeCatalog,
		NodeAdapter: httpnode.NewAdapter(resilience.NewResilientClient("visualizer-node", providerConfig.Timeout)),
	})
	criticAgent := criticagent.NewAgent(queryClient, criticagent.Config{
		Model:         providerConfig.Model,
		RevisionAgent: visualizerAgent,
	})
	stylistAgent := stylistagent.NewAgent(queryClient, stylistagent.Config{Model: providerConfig.Model})

	return orchestrator.NewCanonicalRunner(
		retrieveragent.NewAgent(queryClient, retrieveragent.Config{
			Mode:  retrieveragent.RetrievalModeAuto,
			Store: retrieveragent.FileStore{Root: "data/PaperBananaBench"},
			Model: providerConfig.Model,
		}),
		planneragent.NewAgent(queryClient, planneragent.Config{Model: providerConfig.Model}),
		stylistAgent,
		visualizerAgent,
		criticAgent,
		orchestrator.WithSnapshotStore(snapshotStore),
	), nil
}

func buildStartupLLMClient(logger *zap.Logger, providerName string, providerConfig config.ProviderConfig, options llminfra.ClientOptions) (domainllm.LLMClient, error) {
	if strings.TrimSpace(providerConfig.APIKey) == "" {
		reason := fmt.Sprintf(
			"default provider %s has no API key configured; open Settings to add a key before generating",
			providerName,
		)
		logger.Warn("starting with unconfigured default provider", zap.String("provider", providerName))
		return llminfra.NewUnavailableClient(providerName, reason), nil
	}

	return llminfra.NewLLMClientWithOptions(providerName, providerConfig, options)
}

// agentFactory implements orchestrator.AgentFactory for batch processing.
type agentFactory struct {
	queryClient    domainllm.LLMClient
	genClient      domainllm.LLMClient
	providerConfig config.ProviderConfig
	nodeCatalog    *config.NodeCatalog
	snapshotStore  orchestrator.SnapshotStore
}

func (f *agentFactory) CreateRetriever() domainagent.BaseAgent {
	return retrieveragent.NewAgent(f.queryClient, retrieveragent.Config{
		Mode:  retrieveragent.RetrievalModeAuto,
		Store: retrieveragent.FileStore{Root: "data/PaperBananaBench"},
		Model: f.providerConfig.Model,
	})
}

func (f *agentFactory) CreatePlanner() domainagent.BaseAgent {
	return planneragent.NewAgent(f.queryClient, planneragent.Config{Model: f.providerConfig.Model})
}

func (f *agentFactory) CreateStylist() domainagent.BaseAgent {
	return stylistagent.NewAgent(f.queryClient, stylistagent.Config{Model: f.providerConfig.Model})
}

func (f *agentFactory) CreateVisualizer() domainagent.BaseAgent {
	return visualizeragent.NewAgent(f.genClient, visualizeragent.Config{
		Model:       f.providerConfig.Model,
		NodeCatalog: f.nodeCatalog,
		NodeAdapter: httpnode.NewAdapter(resilience.NewResilientClient("visualizer-node", f.providerConfig.Timeout)),
	})
}

func (f *agentFactory) CreateCritic() domainagent.BaseAgent {
	return criticagent.NewAgent(f.queryClient, criticagent.Config{
		Model:         f.providerConfig.Model,
		RevisionAgent: f.CreateVisualizer(),
	})
}

func loadNodeCatalog(logger *zap.Logger) (*config.NodeCatalog, error) {
	explicitPath := os.Getenv("PAPERBANANA_NODE_CONFIG_FILE")
	for _, path := range []string{
		explicitPath,
		"configs/custom_nodes.yaml",
	} {
		if path == "" {
			continue
		}
		if _, err := os.Stat(path); err != nil {
			if explicitPath == path {
				return nil, fmt.Errorf("stat custom node config %s: %w", path, err)
			}
			continue
		}

		catalog, err := config.LoadNodeConfig(path)
		if err != nil {
			return nil, fmt.Errorf("load custom node config %s: %w", path, err)
		}
		logger.Info("loaded custom node config", zap.String("path", path), zap.Int("nodes", len(catalog.CustomNodes)))
		return catalog, nil
	}

	return nil, nil
}
