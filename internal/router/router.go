package router

import (
	"net/http"
	"spiropoulos94/youtube-downloader/internal/handler"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Router struct {
	router   *chi.Mux
	handlers *handler.Handlers
}

func BuildRouter(handlers *handler.Handlers) *Router {
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
