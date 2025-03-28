package router

import (
	"net/http"
	"spiropoulos94/youtube-downloader/internal/handlers"
	"spiropoulos94/youtube-downloader/internal/workers"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/hibiken/asynqmon"
)

type Router struct {
	router        *chi.Mux
	handlers      *handlers.Handlers
	workerManager *workers.Manager
}

func BuildRouter(handlers *handlers.Handlers, workerManager *workers.Manager) *Router {
	r := &Router{
		router:        chi.NewRouter(),
		handlers:      handlers,
		workerManager: workerManager,
	}
	r.setupRoutes()
	return r
}

func (r *Router) setupRoutes() {
	// Middleware
	r.router.Use(middleware.Logger)
	r.router.Use(middleware.Recoverer)

	// Asynqmon dashboard
	asynqmonHandler := asynqmon.New(asynqmon.Options{
		RedisConnOpt: r.workerManager.GetRedisOpt(),
		RootPath:     "/monitoring", // RootPath specifies the root for asynqmon app
	})
	r.router.Get("/monitoring/*", asynqmonHandler.ServeHTTP)

	// Routes
	r.router.Route("/api", func(router chi.Router) {
		// Download endpoint
		router.Post("/download", r.handlers.YouTube.DownloadVideo)

		// Task status endpoint
		router.Get("/tasks/{task_id}", r.handlers.YouTube.GetTaskStatus)

		// Video download endpoint
		router.Get("/videos/{task_id}", r.handlers.YouTube.ServeVideo)

		// Health check endpoint
		router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		})
	})
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.router.ServeHTTP(w, req)
}
