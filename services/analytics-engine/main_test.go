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
	"github.com/iamraydoan/factoryos/services/analytics-engine/processor"
)



func TestHTTPServerEndpoints(t *testing.T) {
	cfg := config.IngestionConfig{BatchSize: 100, FlushInterval: 1 * time.Second}
	writer := db.NewBatchWriter(nil, cfg, "raw_telemetry")
	oeeCfg := processor.OEEConfig{
		WindowDuration:        1 * time.Hour,
		SnapshotInterval:      1 * time.Hour,
		DefaultIdealCycleTime: 30 * time.Second,
	}
	oeeAgg := processor.NewOEEAggregator(oeeCfg, nil)
	cons := consumer.NewTelemetryConsumer(nil, writer, nil, oeeAgg, 50*time.Millisecond)

	server := startHTTPServer(":0", cons, writer, oeeAgg)


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

	// 4. Test /oee (empty — no data)
	req = httptest.NewRequest(http.MethodGet, "/oee", nil)
	w = httptest.NewRecorder()
	server.Handler.ServeHTTP(w, req)

	resp = w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK for /oee, got: %d", resp.StatusCode)
	}
	var oeeSnapshots []processor.OEESnapshot
	if err := json.NewDecoder(resp.Body).Decode(&oeeSnapshots); err != nil {
		t.Fatalf("failed to decode /oee json: %v", err)
	}
	if len(oeeSnapshots) != 0 {
		t.Fatalf("expected 0 OEE snapshots with no data, got: %d", len(oeeSnapshots))
	}

	// 5. Test /oee?asset_id=unknown (404)
	req = httptest.NewRequest(http.MethodGet, "/oee?asset_id=unknown", nil)
	w = httptest.NewRecorder()
	server.Handler.ServeHTTP(w, req)

	resp = w.Result()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404 for unknown asset, got: %d", resp.StatusCode)
	}

	// 6. Test /oee with data
	oeeAgg.ProcessReading("test-asset", "machine_status", 1.0, "GOOD", time.Now())
	req = httptest.NewRequest(http.MethodGet, "/oee", nil)
	w = httptest.NewRecorder()
	server.Handler.ServeHTTP(w, req)

	resp = w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK for /oee with data, got: %d", resp.StatusCode)
	}
	if err := json.NewDecoder(resp.Body).Decode(&oeeSnapshots); err != nil {
		t.Fatalf("failed to decode /oee json with data: %v", err)
	}
	if len(oeeSnapshots) != 1 {
		t.Fatalf("expected 1 OEE snapshot, got: %d", len(oeeSnapshots))
	}

	// 7. Test /oee?asset_id=test-asset (200)
	req = httptest.NewRequest(http.MethodGet, "/oee?asset_id=test-asset", nil)
	w = httptest.NewRecorder()
	server.Handler.ServeHTTP(w, req)

	resp = w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK for /oee?asset_id=test-asset, got: %d", resp.StatusCode)
	}
	var snap processor.OEESnapshot
	if err := json.NewDecoder(resp.Body).Decode(&snap); err != nil {
		t.Fatalf("failed to decode /oee single asset json: %v", err)
	}
	if snap.PhysicalAssetID != "test-asset" {
		t.Fatalf("expected asset_id test-asset, got: %s", snap.PhysicalAssetID)
	}
}
