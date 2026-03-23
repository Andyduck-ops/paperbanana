package api

import (
	"context"

	"github.com/gin-gonic/gin"
	configservice "github.com/paperbanana/paperbanana/internal/application/config"
	"github.com/paperbanana/paperbanana/internal/api/handlers"
	"github.com/paperbanana/paperbanana/internal/api/middleware"
	"github.com/paperbanana/paperbanana/internal/application/orchestrator"
	domainllm "github.com/paperbanana/paperbanana/internal/domain/llm"
	"github.com/paperbanana/paperbanana/internal/domain/workspace"
	"go.uber.org/zap"
)

// PersistenceServices holds the services needed for persistence endpoints.
type PersistenceServices struct {
	WorkspaceService handlers.WorkspaceService
	HistoryService   handlers.HistoryService
	AssetService     AssetPersistenceService
}

// AssetPersistenceService is the interface for asset operations.
// Matches the persistence.AssetService method signatures.
type AssetPersistenceService interface {
	ListAssets(ctx context.Context, projectID, visualizationID string) ([]*workspace.Asset, error)
	GetAsset(ctx context.Context, projectID, assetID string) (*workspace.Asset, []byte, error)
	ListAssetsByVersion(ctx context.Context, projectID, versionID string) ([]*workspace.Asset, error)
}

// ConfigServices holds the services needed for config endpoints.
type ConfigServices struct {
	ConfigService *configservice.Service
}

// BatchServices holds the services needed for batch endpoints.
type BatchServices struct {
	BatchRunner  *orchestrator.BatchRunner
	AgentFactory orchestrator.AgentFactory
}

// RefineServices holds the services needed for refine endpoints.
type RefineServices struct {
	ClientManager ClientManager
}

// ClientManager is the interface for obtaining LLM clients.
type ClientManager interface {
	GetClient(ctx context.Context, providerID string) (domainllm.LLMClient, error)
}

// SetupRouter creates the main router with generate endpoints.
// For Phase 1-2 compatibility, this only registers the generate endpoints.
func SetupRouter(runner *orchestrator.Runner, logger *zap.Logger) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.Logger(logger))

	// Health endpoints
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	router.GET("/ready", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ready"})
	})

	handler := handlers.NewHandler(handlers.NewRunnerAdapter(runner), logger)

	v1 := router.Group("/api/v1")
	v1.POST("/generate", handler.Generate)
	v1.POST("/generate/stream", handler.StreamGenerate)

	return router
}

// SetupRouterWithPersistence creates the full router with all Phase 3 endpoints.
// This includes workspace, history, and asset routes alongside generate endpoints.
func SetupRouterWithPersistence(runner *orchestrator.Runner, services PersistenceServices, logger *zap.Logger) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.Logger(logger))

	// Health endpoints
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	router.GET("/ready", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ready"})
	})

	// Generate endpoints
	generateHandler := handlers.NewHandler(handlers.NewRunnerAdapter(runner), logger)

	v1 := router.Group("/api/v1")
	v1.POST("/generate", generateHandler.Generate)
	v1.POST("/generate/stream", generateHandler.StreamGenerate)

	// Workspace endpoints (projects, folders, visualizations, move/reparent)
	workspaceHandler := handlers.NewWorkspaceHandler(services.WorkspaceService, logger)
	v1.POST("/projects", workspaceHandler.CreateProject)
	v1.GET("/projects", workspaceHandler.ListProjects)
	v1.GET("/projects/:project_id", workspaceHandler.GetProject)
	v1.POST("/folders", workspaceHandler.CreateFolder)
	v1.POST("/visualizations", workspaceHandler.CreateVisualization)
	v1.GET("/folders/contents", workspaceHandler.ListFolderContents)
	v1.POST("/workspace/move", workspaceHandler.MoveItem)
	v1.POST("/workspace/trash", workspaceHandler.TrashItem)
	v1.POST("/workspace/restore", workspaceHandler.RestoreItem)

	// History/session endpoints
	historyHandler := handlers.NewHistoryHandler(services.HistoryService, logger)
	v1.GET("/history", historyHandler.ListHistory)
	v1.GET("/history/:project_id/:version_id", historyHandler.GetVersion)
	v1.GET("/session/latest", historyHandler.GetLatestSession)
	v1.GET("/session/:session_id", historyHandler.GetSession)

	// Asset endpoints - adapt persistence.AssetService to handlers.AssetService
	assetAdapter := handlers.NewAssetServiceAdapter(services.AssetService)
	assetHandler := handlers.NewAssetHandler(assetAdapter, logger)
	v1.GET("/assets", assetHandler.ListAssets)
	v1.GET("/assets/:project_id/:asset_id", assetHandler.GetAsset)
	v1.GET("/assets/:project_id/:asset_id/download", assetHandler.DownloadAsset)
	v1.GET("/assets/version/:project_id/:version_id", assetHandler.ListAssetsByVersion)

	return router
}

// SetupRouterWithConfig creates the full router with all endpoints including config management.
func SetupRouterWithConfig(runner *orchestrator.Runner, services PersistenceServices, configSvc *ConfigServices, logger *zap.Logger) *gin.Engine {
	return SetupRouterWithConfigAndBatch(runner, services, configSvc, nil, nil, logger)
}

// SetupRouterWithConfigAndBatch creates the full router with all endpoints including config and batch management.
func SetupRouterWithConfigAndBatch(runner *orchestrator.Runner, services PersistenceServices, configSvc *ConfigServices, batchSvc *BatchServices, refineSvc *RefineServices, logger *zap.Logger) *gin.Engine {
	router := SetupRouterWithPersistence(runner, services, logger)

	v1 := router.Group("/api/v1")

	// Provider endpoints
	providerHandler := handlers.NewProviderHandler(configSvc.ConfigService)
	v1.GET("/providers/presets", providerHandler.ListPresets)
	v1.GET("/providers", providerHandler.ListProviders)
	v1.GET("/providers/:id", providerHandler.GetProvider)
	v1.POST("/providers", providerHandler.CreateProvider)
	v1.PUT("/providers/:id", providerHandler.UpdateProvider)
	v1.DELETE("/providers/:id", providerHandler.DeleteProvider)
	v1.POST("/providers/:id/default", providerHandler.SetDefaultProvider)

	// API Key endpoints
	v1.GET("/providers/:id/keys", providerHandler.ListAPIKeys)
	v1.POST("/providers/:id/keys", providerHandler.AddAPIKey)
	v1.DELETE("/providers/:id/keys/:keyId", providerHandler.DeleteAPIKey)
	v1.PATCH("/providers/:id/keys/:keyId", providerHandler.ToggleAPIKey)

	// Model and validation endpoints
	v1.GET("/providers/:id/models", providerHandler.ListModels)
	v1.POST("/providers/:id/test", providerHandler.TestExistingProvider)
	v1.POST("/providers/test", providerHandler.TestProvider)

	// Provider reset endpoint
	v1.POST("/providers/reset", providerHandler.ResetSystemProviders)

	// Config SSE endpoint
	if configSvc.ConfigService.GetWatcher() != nil {
		configSSEHandler := handlers.NewConfigSSEHandler(configSvc.ConfigService.GetWatcher())
		v1.GET("/config/stream", configSSEHandler.StreamConfigChanges)
	}

	// Batch generation endpoint
	if batchSvc != nil && batchSvc.BatchRunner != nil {
		batchHandler := handlers.NewBatchHandler(batchSvc.BatchRunner, batchSvc.AgentFactory, logger)
		v1.POST("/generate/batch", batchHandler.StreamBatchGenerate)
		v1.POST("/batch/download", batchHandler.DownloadBatchZip)
	}

	// Refine endpoint for image enhancement
	if refineSvc != nil && refineSvc.ClientManager != nil {
		refineHandler := handlers.NewRefineHandler(refineSvc.ClientManager, logger)
		v1.POST("/refine", refineHandler.Refine)
	}

	return router
}
