package handlers

import (
	"net/http"
	"spiropoulos94/youtube-downloader/internal/services"
)

// FrontendHandler handles requests for the React frontend
type FrontendHandler struct {
	frontendService *services.FrontendService
}

// NewFrontendHandler creates a new FrontendHandler
func NewFrontendHandler(frontendService *services.FrontendService) *FrontendHandler {
	return &FrontendHandler{
		frontendService: frontendService,
	}
}

// ServeHTTP is the handler function that serves the React frontend
func (h *FrontendHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.frontendService.ServeStaticFiles(w, r)
}
