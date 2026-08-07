package collector

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/iamraydoan/factoryos/platform/edge-runtime/buffer"
	telemetryv1 "github.com/iamraydoan/factoryos/platform/platform-sdk/go/gen/telemetry/v1"
)

type MockCloudPublisher struct {
	mu        sync.Mutex
	online    bool
	published []*telemetryv1.TelemetryPayload
}

func (m *MockCloudPublisher) PublishBatch(ctx context.Context, payloads []*telemetryv1.TelemetryPayload) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.online {
		return errors.New("simulated offline network outage")
	}
	m.published = append(m.published, payloads...)
	return nil
}

func (m *MockCloudPublisher) SetOnline(online bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.online = online
}

func (m *MockCloudPublisher) GetPublishedCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.published)
}

func newTestCollector(t *testing.T, dbName string, online bool) (*TelemetryCollector, *buffer.SQLiteBuffer, *MockCloudPublisher) {
	buf, err := buffer.NewSQLiteBuffer("file:" + dbName + "?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("failed to create sqlite buffer: %v", err)
	}
	pub := &MockCloudPublisher{online: online}
	coll := NewTelemetryCollector(buf, pub)
	return coll, buf, pub
}

func TestCollector_RecordMetric(t *testing.T) {
	ctx := context.Background()
	coll, buf, _ := newTestCollector(t, "coll_metric", true)
	defer buf.Close()

	if err := coll.RecordMetric(ctx, "asset-single-01", "temperature", 45.2, "GOOD"); err != nil {
		t.Fatalf("RecordMetric failed: %v", err)
	}

	count, err := buf.GetPendingCount(ctx)
	if err != nil {
		t.Fatalf("GetPendingCount failed: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 record in buffer, got %d", count)
	}
}

func TestCollector_RecordBatch(t *testing.T) {
	ctx := context.Background()
	coll, buf, _ := newTestCollector(t, "coll_batch", true)
	defer buf.Close()

	metrics := map[string]float64{
		"temperature": 75.2,
		"humidity":    45.0,
	}

	if err := coll.RecordBatch(ctx, "asset-batch-01", metrics); err != nil {
		t.Fatalf("RecordBatch failed: %v", err)
	}

	count, err := buf.GetPendingCount(ctx)
	if err != nil {
		t.Fatalf("GetPendingCount failed: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 batch record in buffer, got %d", count)
	}
}

func TestCollector_FlushBatch_Success(t *testing.T) {
	ctx := context.Background()
	coll, buf, _ := newTestCollector(t, "coll_flush_success", true)
	defer buf.Close()

	if err := coll.RecordMetric(ctx, "asset-flush-01", "pressure", 101.3, "GOOD"); err != nil {
		t.Fatalf("RecordMetric failed: %v", err)
	}

	sent, err := coll.FlushBatch(ctx, 10)
	if err != nil {
		t.Fatalf("FlushBatch failed: %v", err)
	}
	if sent != 1 {
		t.Errorf("expected 1 record sent, got %d", sent)
	}

	count, err := buf.GetPendingCount(ctx)
	if err != nil {
		t.Fatalf("GetPendingCount failed: %v", err)
	}
	if count != 0 {
		t.Errorf("expected buffer count 0 after flush, got %d", count)
	}
}

func TestCollector_FlushBatch_EmptyBuffer(t *testing.T) {
	ctx := context.Background()
	coll, buf, _ := newTestCollector(t, "coll_flush_empty", true)
	defer buf.Close()

	sent, err := coll.FlushBatch(ctx, 10)
	if err != nil {
		t.Fatalf("FlushBatch on empty buffer failed: %v", err)
	}
	if sent != 0 {
		t.Errorf("expected 0 records sent on empty buffer, got %d", sent)
	}
}

func TestCollector_FlushBatch_OfflineRetention(t *testing.T) {
	ctx := context.Background()
	coll, buf, pub := newTestCollector(t, "coll_flush_offline", false) // Start offline
	defer buf.Close()

	if err := coll.RecordMetric(ctx, "asset-offline-01", "speed", 250.0, "GOOD"); err != nil {
		t.Fatalf("RecordMetric failed: %v", err)
	}

	// Flush while offline -> should fail and retain record
	_, err := coll.FlushBatch(ctx, 10)
	if err == nil {
		t.Fatal("expected FlushBatch to fail offline, got nil")
	}

	count, err := buf.GetPendingCount(ctx)
	if err != nil {
		t.Fatalf("GetPendingCount failed: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected record to be retained in buffer during offline, got %d", count)
	}

	// Switch online and flush again
	pub.SetOnline(true)
	sent, err := coll.FlushBatch(ctx, 10)
	if err != nil {
		t.Fatalf("FlushBatch failed after switching online: %v", err)
	}
	if sent != 1 {
		t.Errorf("expected 1 record sent after online, got %d", sent)
	}
}

func TestCollector_SyncWorker(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	coll, buf, pub := newTestCollector(t, "coll_sync_worker", true)
	defer buf.Close()

	if err := coll.RecordMetric(ctx, "asset-worker-01", "current", 12.5, "GOOD"); err != nil {
		t.Fatalf("RecordMetric failed: %v", err)
	}

	go coll.StartSyncWorker(ctx, 20*time.Millisecond, 10)

	time.Sleep(100 * time.Millisecond)
	cancel()

	if pub.GetPublishedCount() != 1 {
		t.Errorf("expected 1 published payload from SyncWorker, got %d", pub.GetPublishedCount())
	}
}
