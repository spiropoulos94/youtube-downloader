package handlers

import (
	"net/http"
	"spiropoulos94/youtube-downloader/internal/services"
)

// FrontendHandler implements FrontendHandlerInterface
type FrontendHandler struct {
	frontendService services.FrontendServiceInterface
}

// NewFrontendHandler creates a new instance of FrontendHandler
func NewFrontendHandler(
	frontendService services.FrontendServiceInterface,
) FrontendHandlerInterface {
	return &FrontendHandler{
		frontendService: frontendService,
	}
}

// ServeFrontend serves the frontend application
func (h *FrontendHandler) ServeFrontend(w http.ResponseWriter, r *http.Request) {
	h.frontendService.ServeStaticFiles(w, r)
}
