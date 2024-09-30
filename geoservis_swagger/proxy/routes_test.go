package main

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestRoutes(t *testing.T) {

	tests := []struct {
		name       string
		path       string
		statusCode int
		body       []byte
	}{
		{"Handler1", "/api/address/search", http.StatusForbidden, []byte(`{"lat": "55.878", "lng": "37.653"}`)},
		{"Handler2", "/api/address/search", http.StatusForbidden, []byte(` "lat": "55.878", "lng": "37.653"`)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			req := httptest.NewRequest("POST", tt.path, bytes.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
			app := &application{
				geo:    NewGeoService("fc47d9338dbcf9a2199f193ec2e5e57857e37378", "954baf5559aa44c49bde9a4dc572801bf48b69e9"),
				logger: logger,
			}
			r := app.setupRouter()

			r.ServeHTTP(w, req)
			if w.Code != tt.statusCode {
				t.Errorf("expected status code %d but got %d", tt.statusCode, w.Code)
			}
		})
	}
}
