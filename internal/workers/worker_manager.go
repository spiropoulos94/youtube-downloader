package workers

import (
	"log"
	"spiropoulos94/youtube-downloader/internal/config"
	"spiropoulos94/youtube-downloader/internal/services"
	"spiropoulos94/youtube-downloader/internal/tasks"
	"time"

	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

// Manager handles the Asynq worker server and task processing
type Manager struct {
	config         *config.Config
	client         *asynq.Client
	server         *asynq.Server
	inspector      *asynq.Inspector
	youtubeService services.YouTubeServiceInterface
	redis          *redis.Client
}

// NewManager creates a new worker manager
func NewManager(config *config.Config, youtubeService services.YouTubeServiceInterface) *Manager {
	redis := redis.NewClient(&redis.Options{
		Addr: config.RedisAddr,
	})

	redisOpt := asynq.RedisClientOpt{
		Addr: redis.Options().Addr,
	}

	// Create an Asynq server with configuration options
	serverOpts := asynq.Config{
		Concurrency:         10,
		HealthCheckInterval: 5 * time.Second,
		// Can add other asynq server configurations here
	}

	client := asynq.NewClient(redisOpt)
	server := asynq.NewServer(redisOpt, serverOpts)
	inspector := asynq.NewInspector(redisOpt)

	return &Manager{
		config:         config,
		client:         client,
		server:         server,
		inspector:      inspector,
		youtubeService: youtubeService,
		redis:          redis,
	}
}

// Start initializes and starts the worker server
func (m *Manager) Start() error {
	log.Println("Starting worker server...")

	// Initialize processors
	downloadProcessor := tasks.NewVideoDownloadProcessor(m.youtubeService)

	// Initialize mux
	mux := asynq.NewServeMux()

	// Register processors
	mux.HandleFunc(tasks.TypeVideoDownload, downloadProcessor.ProcessTask)

	log.Println("Worker server initialized, starting...")
	return m.server.Run(mux)
}

// Stop gracefully stops the worker server
func (m *Manager) Stop() {
	log.Println("Stopping worker server...")
	m.server.Stop()
	m.client.Close()
	m.inspector.Close()
	log.Println("Worker server stopped")
}

// GetClient returns the Asynq client for task enqueuing
func (m *Manager) GetClient() *asynq.Client {
	return m.client
}

// GetInspector returns the Asynq inspector for task inspection
func (m *Manager) GetInspector() *asynq.Inspector {
	return m.inspector
}

// GetRedisOpt returns the Redis connection options
func (m *Manager) GetRedisOpt() asynq.RedisClientOpt {
	return asynq.RedisClientOpt{
		Addr: m.redis.Options().Addr,
	}
}
