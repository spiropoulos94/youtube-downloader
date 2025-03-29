package httputils

import "net/http"

// CleanupResponseWriter wraps http.ResponseWriter to handle cleanup after closing the response
type CleanupResponseWriter struct {
	http.ResponseWriter
	cleanup func()
}

func (w *CleanupResponseWriter) Write(b []byte) (int, error) {
	return w.ResponseWriter.Write(b)
}

func (w *CleanupResponseWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *CleanupResponseWriter) Flush() {
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// NewCleanupResponseWriter creates a new CleanupResponseWriter
func NewCleanupResponseWriter(w http.ResponseWriter, cleanup func()) *CleanupResponseWriter {
	return &CleanupResponseWriter{
		ResponseWriter: w,
		cleanup:        cleanup,
	}
}
