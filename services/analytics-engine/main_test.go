package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/iamraydoan/factoryos/services/analytics-engine/config"
	"github.com/iamraydoan/factoryos/services/analytics-engine/consumer"
	"github.com/iamraydoan/factoryos/services/analytics-engine/db"
)



func TestHTTPServerEndpoints(t *testing.T) {
	cfg := config.IngestionConfig{BatchSize: 100, FlushInterval: 1 * time.Second}
	writer := db.NewBatchWriter(nil, cfg, "raw_telemetry")
	cons := consumer.NewTelemetryConsumer(nil, writer, nil, 50*time.Millisecond)

	server := startHTTPServer(":0", cons, writer)


	defer func() {
		_ = server.Close()
	}()

	// 1. Test /healthz
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	server.Handler.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK for /healthz, got: %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "HEALTHY") {
		t.Fatalf("expected HEALTHY in body, got: %s", string(body))
	}

	// 2. Test /stats
	req = httptest.NewRequest(http.MethodGet, "/stats", nil)
	w = httptest.NewRecorder()
	server.Handler.ServeHTTP(w, req)

	resp = w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK for /stats, got: %d", resp.StatusCode)
	}
	var stats map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		t.Fatalf("failed to decode /stats json: %v", err)
	}
	if stats["service"] != nil && stats["service"] != "analytics-engine" {
		t.Fatalf("unexpected stats payload: %+v", stats)
	}

	// 3. Test /metrics
	req = httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w = httptest.NewRecorder()
	server.Handler.ServeHTTP(w, req)

	resp = w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK for /metrics, got: %d", resp.StatusCode)
	}
	metricsBody, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(metricsBody), "telemetry_consumed_total") {
		t.Fatalf("expected Prometheus metric in body, got: %s", string(metricsBody))
	}
}
