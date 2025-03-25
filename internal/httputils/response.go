package httputils

import (
	"encoding/json"
	"net/http"
)

// Response represents a standard API response
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// SendJSON writes a successful JSON response
func SendJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(Response{
		Success: status >= 200 && status < 300,
		Data:    data,
	})
}

// HTTPError represents an HTTP error
type HTTPError struct {
	Message    string
	StatusCode int
}

func (e *HTTPError) Error() string {
	return e.Message
}

// SendError writes a JSON error response
func SendError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	if httpErr, ok := err.(*HTTPError); ok {
		status = httpErr.StatusCode
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(Response{
		Success: false,
		Error:   err.Error(),
	})
}

// ParseJSON reads a JSON request body
func ParseJSON(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}

// NewError creates a new HTTPError with the given status code and message
func NewError(status int, message string) *HTTPError {
	return &HTTPError{
		StatusCode: status,
		Message:    message,
	}
}

// Common HTTP errors
var (
	ErrBadRequest         = NewError(http.StatusBadRequest, "Bad request")
	ErrUnauthorized       = NewError(http.StatusUnauthorized, "Unauthorized")
	ErrForbidden          = NewError(http.StatusForbidden, "Forbidden")
	ErrNotFound           = NewError(http.StatusNotFound, "Not found")
	ErrMethodNotAllowed   = NewError(http.StatusMethodNotAllowed, "Method not allowed")
	ErrInternalServer     = NewError(http.StatusInternalServerError, "Internal server error")
	ErrServiceUnavailable = NewError(http.StatusServiceUnavailable, "Service unavailable")
)
