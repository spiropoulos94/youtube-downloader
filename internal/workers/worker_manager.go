package workers

import (
	"log"
	"spiropoulos94/youtube-downloader/internal/services"
	"spiropoulos94/youtube-downloader/internal/tasks"

	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

// Manager handles the Asynq worker server and task processing
type Manager struct {
	client         *asynq.Client
	server         *asynq.Server
	inspector      *asynq.Inspector
	youtubeService *services.YouTubeService
	redisClient    *redis.Client
}

// NewManager creates a new worker manager
func NewManager(redisAddr string, youtubeService *services.YouTubeService, redisClient *redis.Client) *Manager {
	redisOpt := asynq.RedisClientOpt{
		Addr: redisAddr,
	}

	return &Manager{
		client: asynq.NewClient(redisOpt),
		server: asynq.NewServer(redisOpt, asynq.Config{
			Concurrency: 10,
		}),
		inspector:      asynq.NewInspector(redisOpt),
		youtubeService: youtubeService,
		redisClient:    redisClient,
	}
}

// Start initializes and starts the worker server
func (m *Manager) Start() error {
	log.Println("Starting worker server...")

	// Initialize processors
	downloadProcessor := tasks.NewVideoDownloadProcessor(m.youtubeService, m.redisClient)

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
