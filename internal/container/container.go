package container

import (
	"net/http"
	"spiropoulos94/youtube-downloader/internal/config"
	"spiropoulos94/youtube-downloader/internal/handler"
	"spiropoulos94/youtube-downloader/internal/service"
)

type Container struct {
	config   *config.Config
	services *Services
	handlers *Handlers
	server   *http.Server
}

type Services struct {
	YouTube *service.YouTubeService
}

type Handlers struct {
	YouTube *handler.YouTubeHandler
}

// InitContainer initializes the container with configuration
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

	// Initialize HTTP server
	c.server = &http.Server{
		Addr: ":" + c.config.Port,
	}

	return nil
}

func (c *Container) GetYouTubeHandler() *handler.YouTubeHandler {
	return c.handlers.YouTube
}

func (c *Container) GetPort() string {
	return c.config.Port
}

func (c *Container) StartServer() error {
	// Set up routes
	http.HandleFunc("/download", c.handlers.YouTube.DownloadVideo)
	return c.server.ListenAndServe()
}
