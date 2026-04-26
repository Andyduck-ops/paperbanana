package api

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/paperbanana/paperbanana/internal/api/handlers"
	"github.com/paperbanana/paperbanana/internal/api/middleware"
	configservice "github.com/paperbanana/paperbanana/internal/application/config"
	"github.com/paperbanana/paperbanana/internal/application/orchestrator"
	domainagent "github.com/paperbanana/paperbanana/internal/domain/agent"
	domainllm "github.com/paperbanana/paperbanana/internal/domain/llm"
	"github.com/paperbanana/paperbanana/internal/domain/workspace"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"gorm.io/gorm"
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
	RegisterRetainedAssets(ctx context.Context, projectID, visualizationID string, versionID *string, artifacts []domainagent.Artifact) ([]*workspace.Asset, error)
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

// SessionRegistry holds the session registry for cancellation support.
type SessionRegistry interface {
	Register(sessionID string, ctx context.Context) (context.Context, context.CancelFunc)
	Cancel(sessionID string) bool
	Exists(sessionID string) bool
}

// RefineServices holds the services needed for refine endpoints.
type RefineServices struct {
	Generator    domainllm.LLMClient
	SessionSaver handlers.RefineSessionSaver
}

// HealthChecker provides dependency health check functionality.
type HealthChecker struct {
	db *gorm.DB
}

// NewHealthChecker creates a new HealthChecker.
func NewHealthChecker(db *gorm.DB) *HealthChecker {
	return &HealthChecker{db: db}
}

// CheckDatabase verifies database connectivity.
func (h *HealthChecker) CheckDatabase(ctx context.Context) error {
	if h.db == nil {
		return nil // No database configured, skip check
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	sqlDB, err := h.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}

// HealthStatus represents the health status of a dependency.
type HealthStatus struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// HealthResponse represents the health check response.
type HealthResponse struct {
	Status   string        `json:"status"`
	Database *HealthStatus `json:"database,omitempty"`
}

// SetupRouter creates the main router with generate endpoints.
// For Phase 1-2 compatibility, this only registers the generate endpoints.
func SetupRouter(runner *orchestrator.Runner, logger *zap.Logger) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.Logger(logger))
	// Apply secure default CORS middleware
	router.Use(middleware.CORS(middleware.DefaultCORSConfig()))

	// Health endpoints (no dependencies to check in basic setup)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	router.GET("/ready", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
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
	return SetupRouterWithPersistenceAndDB(runner, services, nil, logger)
}

// SetupRouterWithPersistenceAndDB creates the full router with database health check support.
func SetupRouterWithPersistenceAndDB(runner *orchestrator.Runner, services PersistenceServices, db *gorm.DB, logger *zap.Logger) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.Logger(logger))
	// Apply secure default CORS middleware
	router.Use(middleware.CORS(middleware.DefaultCORSConfig()))

	healthChecker := NewHealthChecker(db)

	// Health endpoints with dependency checking
	router.GET("/health", func(c *gin.Context) {
		ctx := c.Request.Context()
		response := HealthResponse{Status: "ok"}

		if err := healthChecker.CheckDatabase(ctx); err != nil {
			response.Status = "degraded"
			response.Database = &HealthStatus{
				Status:  "unhealthy",
				Message: err.Error(),
			}
			c.JSON(http.StatusServiceUnavailable, response)
			return
		}

		if db != nil {
			response.Database = &HealthStatus{Status: "healthy"}
		}
		c.JSON(http.StatusOK, response)
	})

	router.GET("/ready", func(c *gin.Context) {
		ctx := c.Request.Context()
		response := HealthResponse{Status: "ready"}

		if err := healthChecker.CheckDatabase(ctx); err != nil {
			response.Status = "not_ready"
			response.Database = &HealthStatus{
				Status:  "unhealthy",
				Message: err.Error(),
			}
			c.JSON(http.StatusServiceUnavailable, response)
			return
		}

		if db != nil {
			response.Database = &HealthStatus{Status: "healthy"}
		}
		c.JSON(http.StatusOK, response)
	})

	// Create asset adapter for both generate handler and asset handler
	assetAdapter := handlers.NewAssetServiceAdapter(services.AssetService)

	// Generate endpoints with asset persistence support
	generateHandler := handlers.NewHandlerWithAssetService(handlers.NewRunnerAdapter(runner), nil, assetAdapter, logger)

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
	v1.GET("/sessions/recent", historyHandler.ListRecentSessions)
	v1.GET("/session/latest", historyHandler.GetLatestSession)
	v1.GET("/session/:session_id", historyHandler.GetSession)

	// Asset endpoints - reuse assetAdapter created for generate handler
	assetHandler := handlers.NewAssetHandler(assetAdapter, logger)
	v1.GET("/assets", assetHandler.ListAssets)
	v1.GET("/assets/:project_id/:asset_id", assetHandler.GetAsset)
	v1.GET("/assets/:project_id/:asset_id/download", assetHandler.DownloadAsset)
	v1.GET("/assets/version/:project_id/:version_id", assetHandler.ListAssetsByVersion)

	return router
}

// SetupRouterWithPersistenceWithRegistry creates the full router with session registry for cancellation support.
func SetupRouterWithPersistenceWithRegistry(runner *orchestrator.Runner, services PersistenceServices, registry *handlers.SessionRegistry, logger *zap.Logger) *gin.Engine {
	return SetupRouterWithPersistenceWithRegistryAndDB(runner, services, registry, nil, logger)
}

// SetupRouterWithPersistenceWithRegistryAndDB creates the full router with session registry and database health check.
func SetupRouterWithPersistenceWithRegistryAndDB(runner *orchestrator.Runner, services PersistenceServices, registry *handlers.SessionRegistry, db *gorm.DB, logger *zap.Logger) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.Logger(logger))
	router.Use(middleware.Metrics())
	// Apply secure default CORS middleware
	router.Use(middleware.CORS(middleware.DefaultCORSConfig()))

	healthChecker := NewHealthChecker(db)

	// Health endpoints with dependency checking
	router.GET("/health", func(c *gin.Context) {
		ctx := c.Request.Context()
		response := HealthResponse{Status: "ok"}

		if err := healthChecker.CheckDatabase(ctx); err != nil {
			response.Status = "degraded"
			response.Database = &HealthStatus{
				Status:  "unhealthy",
				Message: err.Error(),
			}
			c.JSON(http.StatusServiceUnavailable, response)
			return
		}

		if db != nil {
			response.Database = &HealthStatus{Status: "healthy"}
		}
		c.JSON(http.StatusOK, response)
	})

	router.GET("/ready", func(c *gin.Context) {
		ctx := c.Request.Context()
		response := HealthResponse{Status: "ready"}

		if err := healthChecker.CheckDatabase(ctx); err != nil {
			response.Status = "not_ready"
			response.Database = &HealthStatus{
				Status:  "unhealthy",
				Message: err.Error(),
			}
			c.JSON(http.StatusServiceUnavailable, response)
			return
		}

		if db != nil {
			response.Database = &HealthStatus{Status: "healthy"}
		}
		c.JSON(http.StatusOK, response)
	})

	// Create asset adapter for both generate handler and asset handler
	assetAdapter := handlers.NewAssetServiceAdapter(services.AssetService)

	// Generate endpoints with session registry for cancellation and asset persistence
	generateHandler := handlers.NewHandlerWithAssetService(handlers.NewRunnerAdapter(runner), registry, assetAdapter, logger)

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
	v1.GET("/sessions/recent", historyHandler.ListRecentSessions)
	v1.GET("/session/latest", historyHandler.GetLatestSession)
	v1.GET("/session/:session_id", historyHandler.GetSession)

	// Asset endpoints - reuse assetAdapter from above
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
	return SetupRouterWithConfigAndBatchAndDB(runner, services, configSvc, batchSvc, refineSvc, nil, logger, middleware.AuthConfig{}, middleware.RateLimitConfig{})
}

// SetupRouterWithConfigAndBatchAndDB creates the full router with all endpoints and database health check.
func SetupRouterWithConfigAndBatchAndDB(runner *orchestrator.Runner, services PersistenceServices, configSvc *ConfigServices, batchSvc *BatchServices, refineSvc *RefineServices, db *gorm.DB, logger *zap.Logger, authCfg middleware.AuthConfig, rateLimitCfg middleware.RateLimitConfig) *gin.Engine {
	// Create session registry for cancellation support
	sessionRegistry := handlers.NewSessionRegistry()

	router := SetupRouterWithPersistenceWithRegistryAndDB(runner, services, sessionRegistry, db, logger)

	// Apply authentication middleware (disabled by default — set auth_enabled=true to activate)
	if authCfg.Enabled {
		router.Use(middleware.Auth(authCfg))
		logger.Info("API key authentication enabled")
	}

	// Apply rate limiting middleware (enabled by default — 60 req/min, 10 burst)
	if rateLimitCfg.RequestsPerMinute > 0 {
		router.Use(middleware.RateLimitByIP(rateLimitCfg))
	}

	// Add Prometheus metrics endpoint
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	v1 := router.Group("/api/v1")

	// Cancel endpoint
	cancelHandler := handlers.NewCancelHandler(sessionRegistry, logger)
	v1.POST("/sessions/:session_id/cancel", cancelHandler.Cancel)

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

	// Channel endpoints (alias for providers, aligned with PRD terminology)
	channelHandler := handlers.NewChannelHandler(configSvc.ConfigService)
	v1.GET("/channels/presets", channelHandler.ListPresets)
	v1.GET("/channels", channelHandler.ListChannels)
	v1.GET("/channels/:id", channelHandler.GetChannel)
	v1.POST("/channels", channelHandler.CreateChannel)
	v1.PUT("/channels/:id", channelHandler.UpdateChannel)
	v1.DELETE("/channels/:id", channelHandler.DeleteChannel)
	v1.GET("/channels/:id/models", channelHandler.FetchChannelModels)
	v1.POST("/channels/test", channelHandler.TestChannel)
	v1.GET("/channels/:id/keys", channelHandler.ListChannelAPIKeys)
	v1.POST("/channels/:id/keys", channelHandler.AddChannelAPIKey)
	v1.DELETE("/channels/:id/keys/:keyId", channelHandler.DeleteChannelAPIKey)

	// Role assignment endpoint
	v1.PUT("/config/roles", channelHandler.SetRoleAssignment)
	v1.DELETE("/config/roles/:role", channelHandler.ClearRoleAssignment)

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
	if refineSvc != nil && refineSvc.Generator != nil {
		refineHandler := handlers.NewRefineHandler(refineSvc.Generator, refineSvc.SessionSaver, logger)
		v1.POST("/refine", refineHandler.Refine)
	}

	return router
}
