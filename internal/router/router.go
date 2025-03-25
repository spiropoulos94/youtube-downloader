package router

import (
	"net/http"
	"spiropoulos94/youtube-downloader/internal/handler"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Router struct {
	router   *chi.Mux
	handlers *handler.YouTubeHandler
}

func NewRouter(handlers *handler.YouTubeHandler) *Router {
	return &Router{
		router:   chi.NewRouter(),
		handlers: handlers,
	}
}

func (r *Router) SetupRoutes() {
	// Middleware
	r.router.Use(middleware.Logger)
	r.router.Use(middleware.Recoverer)

	// Routes
	r.router.Route("/api", func(router chi.Router) {
		// Download endpoint
		router.Post("/download", r.handlers.DownloadVideo)

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
