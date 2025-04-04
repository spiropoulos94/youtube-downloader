package container

import (
	"context"
	"fmt"
	"net/http"
	"spiropoulos94/youtube-downloader/internal/config"
	"spiropoulos94/youtube-downloader/internal/handlers"
	"spiropoulos94/youtube-downloader/internal/router"
	"spiropoulos94/youtube-downloader/internal/services"
	"spiropoulos94/youtube-downloader/internal/validators"
	"spiropoulos94/youtube-downloader/internal/workers"

	"github.com/redis/go-redis/v9"
)

type Container struct {
	config        *config.Config
	services      *services.Services
	handlers      *handlers.Handlers
	router        *router.Router
	server        *http.Server
	workerManager *workers.Manager
	redis         *redis.Client
}

// InitContainer Initializes the container with configuration and Builds it
func InitContainer() (*Container, error) {
	cfg := config.Load()
	container := NewContainer(cfg)

	// Check Redis connectivity
	if err := container.redis.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis at %s: %v", cfg.RedisAddr, err)
	}

	if err := container.Build(); err != nil {
		return nil, err
	}
	return container, nil
}

// NewContainer creates a new container with the given configuration and dependencies
func NewContainer(config *config.Config) *Container {
	// Create Redis client for services that might still need direct access
	redis := redis.NewClient(&redis.Options{
		Addr: config.RedisAddr,
	})

	// Create core services
	youtubeService := services.NewYouTubeService(config, redis)
	cleanupService := services.NewCleanupService(config, redis)
	frontendService := services.NewFrontendService()

	// Create worker manager with dependencies
	workerManager := workers.NewManager(config, youtubeService)

	// Create validators
	urlValidator := validators.NewYouTubeURLValidator()

	// Create handlers with direct dependencies
	youtubeHandler := handlers.NewYouTubeHandler(
		config,
		youtubeService,
		workerManager.GetClient(),
		workerManager.GetInspector(),
		urlValidator,
	)
	frontendHandler := handlers.NewFrontendHandler(frontendService)

	return &Container{
		config:        config,
		services:      &services.Services{YouTube: youtubeService, Cleanup: cleanupService, Frontend: frontendService},
		handlers:      &handlers.Handlers{YouTube: youtubeHandler, Frontend: frontendHandler},
		workerManager: workerManager,
		redis:         redis,
	}
}

// Build builds the container for the HTTP server and workers
func (c *Container) Build() error {
	// Initialize router
	c.router = router.BuildRouter(c.handlers, c.workerManager)

	// Initialize server
	c.server = &http.Server{
		Addr:    ":" + c.config.Port,
		Handler: c.router,
	}

	// Start workers in a goroutine
	go func() {
		if err := c.workerManager.Start(); err != nil {
			panic(err)
		}
	}()

	// Start cleanup service to remove files that haven't been requested in the last hour
	c.services.Cleanup.Start()

	return nil
}

func (c *Container) StartServer() error {
	return c.server.ListenAndServe()
}

func (c *Container) Close() error {
	c.workerManager.Stop()
	c.services.Cleanup.Stop()
	c.redis.Close()
	return c.server.Close()
}

func (c *Container) GetPort() string {
	return c.config.Port
}

func (c *Container) GetHandlers() *handlers.Handlers {
	return c.handlers
}

func (c *Container) GetWorkerManager() *workers.Manager {
	return c.workerManager
}
