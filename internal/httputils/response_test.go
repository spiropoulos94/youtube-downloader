package httputils

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSendJSON(t *testing.T) {
	tests := []struct {
		name       string
		status     int
		data       interface{}
		wantStatus int
		wantBody   string
	}{
		{
			name:       "Success response",
			status:     http.StatusOK,
			data:       map[string]string{"message": "success"},
			wantStatus: http.StatusOK,
			wantBody:   `{"success":true,"data":{"message":"success"}}`,
		},
		{
			name:       "Created response",
			status:     http.StatusCreated,
			data:       map[string]interface{}{"id": 1, "name": "test"},
			wantStatus: http.StatusCreated,
			wantBody:   `{"success":true,"data":{"id":1,"name":"test"}}`,
		},
		{
			name:       "Not success status",
			status:     http.StatusBadRequest,
			data:       map[string]string{"message": "invalid input"},
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"success":false,"data":{"message":"invalid input"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			SendJSON(w, tt.status, tt.data)

			resp := w.Result()
			body, _ := io.ReadAll(resp.Body)
			defer resp.Body.Close()

			// Normalize JSON output for comparison
			var normalizedActual bytes.Buffer
			err := json.Compact(&normalizedActual, body)
			if err != nil {
				t.Fatalf("Error compacting actual JSON: %v", err)
			}

			var normalizedExpected bytes.Buffer
			err = json.Compact(&normalizedExpected, []byte(tt.wantBody))
			if err != nil {
				t.Fatalf("Error compacting expected JSON: %v", err)
			}

			// Check response status and body
			if resp.StatusCode != tt.wantStatus {
				t.Errorf("Status code = %v, want %v", resp.StatusCode, tt.wantStatus)
			}

			if normalizedActual.String() != normalizedExpected.String() {
				t.Errorf("Response body = %v, want %v", normalizedActual.String(), normalizedExpected.String())
			}

			// Check content type
			contentType := resp.Header.Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("Content-Type = %v, want application/json", contentType)
			}
		})
	}
}

func TestSendError(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantBody   string
	}{
		{
			name:       "Regular error",
			err:        errors.New("something went wrong"),
			wantStatus: http.StatusInternalServerError,
			wantBody:   `{"success":false,"error":"something went wrong"}`,
		},
		{
			name:       "HTTP error",
			err:        NewError(http.StatusBadRequest, "Invalid input"),
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"success":false,"error":"Invalid input"}`,
		},
		{
			name:       "Predefined error",
			err:        ErrNotFound,
			wantStatus: http.StatusNotFound,
			wantBody:   `{"success":false,"error":"Not found"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			SendError(w, tt.err)

			resp := w.Result()
			body, _ := io.ReadAll(resp.Body)
			defer resp.Body.Close()

			// Normalize JSON output for comparison
			var normalizedActual bytes.Buffer
			err := json.Compact(&normalizedActual, body)
			if err != nil {
				t.Fatalf("Error compacting actual JSON: %v", err)
			}

			var normalizedExpected bytes.Buffer
			err = json.Compact(&normalizedExpected, []byte(tt.wantBody))
			if err != nil {
				t.Fatalf("Error compacting expected JSON: %v", err)
			}

			// Check response status and body
			if resp.StatusCode != tt.wantStatus {
				t.Errorf("Status code = %v, want %v", resp.StatusCode, tt.wantStatus)
			}

			if normalizedActual.String() != normalizedExpected.String() {
				t.Errorf("Response body = %v, want %v", normalizedActual.String(), normalizedExpected.String())
			}

			// Check content type
			contentType := resp.Header.Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("Content-Type = %v, want application/json", contentType)
			}
		})
	}
}

func TestParseJSON(t *testing.T) {
	tests := []struct {
		name      string
		body      string
		target    interface{}
		wantError bool
	}{
		{
			name: "Valid JSON",
			body: `{"name":"test","value":123}`,
			target: &struct {
				Name  string
				Value int
			}{},
			wantError: false,
		},
		{
			name: "Invalid JSON",
			body: `{"name":"test","value":123`,
			target: &struct {
				Name  string
				Value int
			}{},
			wantError: true,
		},
		{
			name: "Empty body",
			body: ``,
			target: &struct {
				Name  string
				Value int
			}{},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBufferString(tt.body))
			err := ParseJSON(req, tt.target)

			if (err != nil) != tt.wantError {
				t.Errorf("ParseJSON() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestHTTPErrorInterface(t *testing.T) {
	err := NewError(http.StatusBadRequest, "Test error")

	if err.Error() != "Test error" {
		t.Errorf("Error() = %v, want %v", err.Error(), "Test error")
	}

	if err.StatusCode != http.StatusBadRequest {
		t.Errorf("StatusCode = %v, want %v", err.StatusCode, http.StatusBadRequest)
	}
}
