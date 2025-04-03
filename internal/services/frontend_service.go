package services

import (
	"net/http"
	"os"
	"path/filepath"
)

// FrontendService implements FrontendServiceInterface
type FrontendService struct {
	frontendDir string
	fileServer  http.Handler
}

// NewFrontendService creates a new FrontendService
func NewFrontendService() FrontendServiceInterface {
	workDir, _ := os.Getwd()
	buildPath := filepath.Join(workDir, "frontend/build")

	return &FrontendService{
		frontendDir: buildPath,
		fileServer:  http.FileServer(http.Dir(buildPath)),
	}
}

// ServeStaticFiles serves the static files from the React app build directory
func (s *FrontendService) ServeStaticFiles(w http.ResponseWriter, r *http.Request) {
	// Check if the file exists
	path := filepath.Join(s.frontendDir, r.URL.Path)
	_, err := os.Stat(path)

	if os.IsNotExist(err) {
		// File doesn't exist, serve index.html for client-side routing
		http.ServeFile(w, r, filepath.Join(s.frontendDir, "index.html"))
		return
	}

	// File exists, serve it
	s.fileServer.ServeHTTP(w, r)
}
