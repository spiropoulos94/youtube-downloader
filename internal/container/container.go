package container

import (
	"net/http"
	"spiropoulos94/youtube-downloader/internal/config"
	"spiropoulos94/youtube-downloader/internal/handlers"
	"spiropoulos94/youtube-downloader/internal/router"
	"spiropoulos94/youtube-downloader/internal/services"
	"spiropoulos94/youtube-downloader/internal/workers"
)

type Container struct {
	config   *config.Config
	services *services.Services
	handlers *handlers.Handlers
	router   *router.Router
	server   *http.Server
	worker   *workers.Manager
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

func NewContainer(config *config.Config) *Container {
	return &Container{
		config: config,
	}
}

func (c *Container) Build() error {
	// Initialize services
	c.services = &services.Services{
		YouTube: services.NewYouTubeService(c.config.OutputDir),
	}

	// Initialize worker
	c.worker = workers.NewManager(c.config.RedisAddr, c.services.YouTube)

	// Initialize handlers
	c.handlers = &handlers.Handlers{
		YouTube: handlers.NewYouTubeHandler(c.services.YouTube, c.worker.GetClient(), c.worker.GetInspector()),
	}

	// Build router
	c.router = router.BuildRouter(c.handlers)

	// Initialize HTTP server
	c.server = &http.Server{
		Addr:    ":" + c.config.Port,
		Handler: c.router,
	}

	return nil
}

func (c *Container) StartServer() error {
	// Start workers in a goroutine
	go func() {
		if err := c.worker.Start(); err != nil {
			panic(err)
		}
	}()

	// Start HTTP server
	return c.server.ListenAndServe()
}

func (c *Container) Close() error {
	c.worker.Stop()
	return nil
}

func (c *Container) GetPort() string {
	return c.config.Port
}
