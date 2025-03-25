package container

import (
	"flag"
	"spiropoulos94/youtube-downloader/internal/handler"
	"spiropoulos94/youtube-downloader/internal/service"
)

type Container struct {
	config   *Config
	services *Services
	handlers *Handlers
}

type Config struct {
	Port      string
	OutputDir string
}

type Services struct {
	YouTube *service.YouTubeService
}

type Handlers struct {
	YouTube *handler.YouTubeHandler
}

// InitContainer initializes the container with command line flags
func InitContainer() (*Container, error) {
	// Define command line flags
	port := flag.String("port", "8080", "Server port")
	outputDir := flag.String("output", "downloads", "Output directory for downloaded videos")
	flag.Parse()

	config := &Config{
		Port:      *port,
		OutputDir: *outputDir,
	}

	container := NewContainer(config)
	if err := container.Build(); err != nil {
		return nil, err
	}

	return container, nil
}

func NewContainer(config *Config) *Container {
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

	return nil
}

func (c *Container) GetYouTubeHandler() *handler.YouTubeHandler {
	return c.handlers.YouTube
}

func (c *Container) GetPort() string {
	return c.config.Port
}
