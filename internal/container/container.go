package container

import (
	"net/http"
	"spiropoulos94/youtube-downloader/internal/config"
	"spiropoulos94/youtube-downloader/internal/handler"
	"spiropoulos94/youtube-downloader/internal/router"
	"spiropoulos94/youtube-downloader/internal/service"
)

type Container struct {
	config   *config.Config
	services *Services
	handlers *Handlers
	router   *router.Router
	server   *http.Server
}

type Services struct {
	YouTube *service.YouTubeService
}

type Handlers struct {
	YouTube *handler.YouTubeHandler
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
	c.services = &Services{
		YouTube: service.NewYouTubeService(c.config.OutputDir),
	}

	// Initialize handlers
	c.handlers = &Handlers{
		YouTube: handler.NewYouTubeHandler(c.services.YouTube),
	}

	// Initialize router
	c.router = router.NewRouter(c.handlers.YouTube)
	c.router.SetupRoutes()

	// Initialize HTTP server
	c.server = &http.Server{
		Addr:    ":" + c.config.Port,
		Handler: c.router,
	}

	return nil
}

func (c *Container) GetPort() string {
	return c.config.Port
}

func (c *Container) StartServer() error {
	return c.server.ListenAndServe()
}
