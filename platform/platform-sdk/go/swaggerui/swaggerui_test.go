package swaggerui

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandler_HTML(t *testing.T) {
	spec := []byte("openapi: 3.0.3\ninfo:\n  title: Test API")
	handler := Handler("Telemetry API", spec)

	req := httptest.NewRequest("GET", "/docs", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	body := w.Body.String()
	if !strings.Contains(body, "Telemetry API — FactoryOS API Documentation") {
		t.Errorf("expected HTML title in body, got: %s", body)
	}
}

func TestHandler_SpecFile(t *testing.T) {
	spec := []byte("openapi: 3.0.3\ninfo:\n  title: Test API")
	handler := Handler("Telemetry API", spec)

	req := httptest.NewRequest("GET", "/openapi.yaml", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	body := w.Body.String()
	if !strings.Contains(body, "title: Test API") {
		t.Errorf("expected spec content in body, got: %s", body)
	}
}
