package worker

import (
	"spiropoulos94/youtube-downloader/internal/service"
	"spiropoulos94/youtube-downloader/internal/tasks"

	"github.com/hibiken/asynq"
)

// Manager handles the Asynq worker server and task processing
type Manager struct {
	client         *asynq.Client
	server         *asynq.Server
	youtubeService *service.YouTubeService
}

// NewManager creates a new worker manager
func NewManager(redisAddr string, youtubeService *service.YouTubeService) *Manager {
	redisOpt := asynq.RedisClientOpt{
		Addr: redisAddr,
	}

	return &Manager{
		client: asynq.NewClient(redisOpt),
		server: asynq.NewServer(redisOpt, asynq.Config{
			Concurrency: 10,
		}),
		youtubeService: youtubeService,
	}
}

// Start initializes and starts the worker server
func (m *Manager) Start() error {
	// Initialize processors
	downloadProcessor := tasks.NewVideoDownloadProcessor(m.youtubeService)

	// Initialize mux
	mux := asynq.NewServeMux()

	// Register processors
	mux.HandleFunc(tasks.TypeVideoDownload, downloadProcessor.ProcessTask)

	return m.server.Run(mux)
}

// Stop gracefully stops the worker server
func (m *Manager) Stop() {
	m.server.Stop()
	m.client.Close()
}

// GetClient returns the Asynq client for task enqueuing
func (m *Manager) GetClient() *asynq.Client {
	return m.client
}
