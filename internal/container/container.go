package container

import (
	"net/http"
	"spiropoulos94/youtube-downloader/internal/config"
	"spiropoulos94/youtube-downloader/internal/handlers"
	"spiropoulos94/youtube-downloader/internal/router"
	"spiropoulos94/youtube-downloader/internal/services"
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
	redisClient   *redis.Client
}

// InitContainer Initializes the container with configuration and Builds it
func InitContainer() (*Container, error) {
	cfg := config.Load()
	container := NewContainer(cfg)
	if err := container.Build(); err != nil {
		return nil, err
	}
	return container, nil
}

// NewContainer creates a new container with the given configuration and dependencies
func NewContainer(config *config.Config) *Container {
	// Create Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr: config.RedisAddr,
	})

	// Create services
	youtubeService := services.NewYouTubeService(config.OutputDir)

	// Create worker manager
	workerManager := workers.NewManager(config.RedisAddr, youtubeService, redisClient)

	// Create handlers
	handlers := &handlers.Handlers{
		YouTube: handlers.NewYouTubeHandler(youtubeService, workerManager.GetClient(), workerManager.GetInspector(), redisClient),
	}

	return &Container{
		config:        config,
		services:      &services.Services{YouTube: youtubeService},
		handlers:      handlers,
		workerManager: workerManager,
		redisClient:   redisClient,
	}
}

// Build builds the container for the HTTP server and workers
func (c *Container) Build() error {
	// Initialize router
	c.router = router.BuildRouter(c.handlers)

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

	return nil
}

func (c *Container) StartServer() error {
	return c.server.ListenAndServe()
}

func (c *Container) Close() error {
	c.workerManager.Stop()
	c.redisClient.Close()
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
