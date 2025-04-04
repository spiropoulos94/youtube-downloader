package router

import (
	"net/http"
	"net/http/httptest"
	"spiropoulos94/youtube-downloader/internal/handlers"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock implementations
type MockYouTubeHandler struct {
	mock.Mock
}

func (m *MockYouTubeHandler) DownloadVideo(w http.ResponseWriter, r *http.Request) {
	m.Called(w, r)
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(`{"success":true}`))
}

func (m *MockYouTubeHandler) GetTaskStatus(w http.ResponseWriter, r *http.Request) {
	m.Called(w, r)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"success":true,"status":"completed"}`))
}

func (m *MockYouTubeHandler) ServeVideo(w http.ResponseWriter, r *http.Request) {
	m.Called(w, r)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("mock video data"))
}

type MockFrontendHandler struct {
	mock.Mock
}

func (m *MockFrontendHandler) ServeFrontend(w http.ResponseWriter, r *http.Request) {
	m.Called(w, r)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("mock frontend"))
}

// TestRouterServeHTTP tests the ServeHTTP method of Router
func TestRouterServeHTTP(t *testing.T) {
	// Create a simple Router implementation with a chi router
	router := &Router{
		router: chi.NewRouter(),
	}

	// Set up a simple health check route
	router.router.Get("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Test the health endpoint
	req := httptest.NewRequest("GET", "/api/health", nil)
	w := httptest.NewRecorder()

	// Call ServeHTTP method
	router.ServeHTTP(w, req)

	// Verify the response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "OK", w.Body.String())
}

// TestRouterAPIEndpoints tests the API endpoints
func TestRouterAPIEndpoints(t *testing.T) {
	// Create mock handlers
	mockYouTubeHandler := new(MockYouTubeHandler)
	mockFrontendHandler := new(MockFrontendHandler)

	// Create handlers struct with mocks
	handlers := &handlers.Handlers{
		YouTube:  mockYouTubeHandler,
		Frontend: mockFrontendHandler,
	}

	// Create a router manually with just the mock handlers
	router := &Router{
		router:   chi.NewRouter(),
		handlers: handlers,
	}

	// Setup API routes only (skip asynqmon mounting)
	router.router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		})
	})

	// Setup API routes
	router.router.Route("/api", func(r chi.Router) {
		r.Post("/download", mockYouTubeHandler.DownloadVideo)
		r.Get("/tasks/{task_id}", mockYouTubeHandler.GetTaskStatus)
		r.Get("/videos/{task_id}", mockYouTubeHandler.ServeVideo)
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		})
	})

	// Setup frontend route
	router.router.Get("/*", mockFrontendHandler.ServeFrontend)

	// Test API health endpoint
	req := httptest.NewRequest("GET", "/api/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "OK", w.Body.String())

	// Test download endpoint
	mockYouTubeHandler.On("DownloadVideo", mock.Anything, mock.Anything).Return()
	req = httptest.NewRequest("POST", "/api/download", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusAccepted, w.Code)
	mockYouTubeHandler.AssertCalled(t, "DownloadVideo", w, mock.Anything)

	// Test task status endpoint
	mockYouTubeHandler.On("GetTaskStatus", mock.Anything, mock.Anything).Return()
	req = httptest.NewRequest("GET", "/api/tasks/123", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	mockYouTubeHandler.AssertCalled(t, "GetTaskStatus", w, mock.Anything)

	// Test video endpoint
	mockYouTubeHandler.On("ServeVideo", mock.Anything, mock.Anything).Return()
	req = httptest.NewRequest("GET", "/api/videos/123", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	mockYouTubeHandler.AssertCalled(t, "ServeVideo", w, mock.Anything)

	// Test frontend endpoint
	mockFrontendHandler.On("ServeFrontend", mock.Anything, mock.Anything).Return()
	req = httptest.NewRequest("GET", "/", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	mockFrontendHandler.AssertCalled(t, "ServeFrontend", w, mock.Anything)

	// Verify all expectations were met
	mockYouTubeHandler.AssertExpectations(t)
	mockFrontendHandler.AssertExpectations(t)
}
