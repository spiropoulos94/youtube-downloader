package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

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
