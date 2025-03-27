package router

import (
	"net/http"
	"spiropoulos94/youtube-downloader/internal/handlers"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Router struct {
	router   *chi.Mux
	handlers *handlers.Handlers
}

func BuildRouter(handlers *handlers.Handlers) *Router {
	r := &Router{
		router:   chi.NewRouter(),
		handlers: handlers,
	}
	r.setupRoutes()
	return r
}

func (r *Router) setupRoutes() {
	// Middleware
	r.router.Use(middleware.Logger)
	r.router.Use(middleware.Recoverer)

	// Routes
	r.router.Route("/api", func(router chi.Router) {
		// Download endpoint
		router.Post("/download", r.handlers.YouTube.DownloadVideo)

		// Task status endpoint
		router.Get("/tasks/{taskID}", r.handlers.YouTube.GetTaskStatus)

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
