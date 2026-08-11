package db

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/iamraydoan/factoryos/services/analytics-engine/config"
)

type mockInserter struct {
	mu          sync.Mutex
	copiedRows  [][]any
	copyCalls   int
	shouldError bool
}

func (m *mockInserter) CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.shouldError {
		return 0, pgx.ErrTxClosed
	}

	m.copyCalls++
	var count int64
	for rowSrc.Next() {
		values, err := rowSrc.Values()
		if err != nil {
			return count, err
		}
		m.copiedRows = append(m.copiedRows, values)
		count++
	}
	return count, rowSrc.Err()
}
// waitForCondition polls check() every 5ms until it returns true or timeout elapses.
// Use this instead of time.Sleep to avoid flaky tests on slow CI runners.
func waitForCondition(t *testing.T, timeout time.Duration, check func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if check() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("condition not met within timeout")
}


func TestBatchWriter_FlushOnThreshold(t *testing.T) {
	mock := &mockInserter{}
	cfg := config.IngestionConfig{BatchSize: 3, FlushInterval: 1 * time.Second, ChannelCapacityMultiplier: 4}
	writer := NewBatchWriter(mock, cfg, "raw_telemetry")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	writer.Start(ctx)

	now := time.Now()
	// Enqueue 3 records to trigger size threshold
	writer.Enqueue(TelemetryRecord{Time: now, PhysicalAssetID: "asset-1", MetricName: "temp", Value: 75.0, Quality: "GOOD"})
	writer.Enqueue(TelemetryRecord{Time: now, PhysicalAssetID: "asset-1", MetricName: "pressure", Value: 101.3, Quality: "GOOD"})
	writer.Enqueue(TelemetryRecord{Time: now, PhysicalAssetID: "asset-1", MetricName: "vibe", Value: 2.1, Quality: "GOOD"})

	// Wait until the worker has flushed the batch.
	waitForCondition(t, 2*time.Second, func() bool {
		mock.mu.Lock()
		defer mock.mu.Unlock()
		return len(mock.copiedRows) == 3
	})

	mock.mu.Lock()
	defer mock.mu.Unlock()

	if len(mock.copiedRows) != 3 {
		t.Fatalf("expected 3 copied rows, got: %d", len(mock.copiedRows))
	}
	if mock.copyCalls != 1 {
		t.Fatalf("expected 1 copy call, got: %d", mock.copyCalls)
	}

	writer.Stop()
}

func TestBatchWriter_FlushOnTicker(t *testing.T) {
	mock := &mockInserter{}
	cfg := config.IngestionConfig{BatchSize: 100, FlushInterval: 50 * time.Millisecond, ChannelCapacityMultiplier: 4}
	writer := NewBatchWriter(mock, cfg, "raw_telemetry")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	writer.Start(ctx)

	// Enqueue 1 record (less than batchSize 100)
	writer.Enqueue(TelemetryRecord{Time: time.Now(), PhysicalAssetID: "asset-2", MetricName: "rpm", Value: 3000.0, Quality: "GOOD"})

	// Wait until the ticker fires and the worker flushes the single record.
	waitForCondition(t, 2*time.Second, func() bool {
		mock.mu.Lock()
		defer mock.mu.Unlock()
		return len(mock.copiedRows) == 1
	})

	mock.mu.Lock()
	if len(mock.copiedRows) != 1 {
		t.Fatalf("expected 1 copied row on ticker flush, got: %d", len(mock.copiedRows))
	}
	mock.mu.Unlock()

	writer.Stop()

	inserted, batches := writer.Stats()
	if inserted != 1 || batches != 1 {
		t.Fatalf("expected Stats=(1, 1), got (%d, %d)", inserted, batches)
	}
}

func TestBatchWriter_ErrorHandling(t *testing.T) {
	mock := &mockInserter{shouldError: true}
	cfg := config.IngestionConfig{BatchSize: 2, FlushInterval: 50 * time.Millisecond, ChannelCapacityMultiplier: 4}
	writer := NewBatchWriter(mock, cfg, "raw_telemetry")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	writer.Start(ctx)

	writer.Enqueue(TelemetryRecord{Time: time.Now(), PhysicalAssetID: "asset-err", MetricName: "err", Value: 0.0})
	writer.Enqueue(TelemetryRecord{Time: time.Now(), PhysicalAssetID: "asset-err", MetricName: "err2", Value: 0.0})

	// The worker will attempt a flush after the batch threshold is hit.
	// Stop() drains and flushes, then waits for the goroutine to exit.
	writer.Stop()

	// Error occurred, so inserted count should remain 0
	inserted, _ := writer.Stats()
	if inserted != 0 {
		t.Fatalf("expected 0 inserted on error, got: %d", inserted)
	}
}

func TestNewBatchWriter_DefaultConfig(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://test:test@localhost:5432/testdb?sslmode=disable")

	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}
	writer := NewBatchWriter(nil, cfg.Ingestion, cfg.Database.TableName)
	if writer.batchSize != 500 {
		t.Fatalf("expected default batchSize 500, got: %d", writer.batchSize)
	}
	if writer.flushInterval != 200*time.Millisecond {
		t.Fatalf("expected default flushInterval 200ms, got: %v", writer.flushInterval)
	}
	if writer.tableName != "raw_telemetry" {
		t.Fatalf("expected default tableName raw_telemetry, got: %s", writer.tableName)
	}
}

func TestBatchWriter_FlushNilInserter(t *testing.T) {
	cfg := config.IngestionConfig{BatchSize: 2, FlushInterval: 50 * time.Millisecond, ChannelCapacityMultiplier: 4}
	writer := NewBatchWriter(nil, cfg, "raw_telemetry")
	writer.Enqueue(TelemetryRecord{Time: time.Now(), PhysicalAssetID: "asset-nil", MetricName: "metric", Value: 1.0})
	
	err := writer.Flush(context.Background())
	if err != nil {
		t.Fatalf("expected nil error when flushing with nil inserter, got: %v", err)
	}
}



func TestBatchWriter_ChannelBufferFull(t *testing.T) {
	mock := &mockInserter{}
	cfg := config.IngestionConfig{BatchSize: 1, FlushInterval: 50 * time.Millisecond, ChannelCapacityMultiplier: 1}
	writer := NewBatchWriter(mock, cfg, "raw_telemetry")

	// Channel capacity is 1*1 = 1. Without worker consuming, 1st succeeds, 2nd drops.
	err1 := writer.Enqueue(TelemetryRecord{Time: time.Now(), PhysicalAssetID: "asset-1", MetricName: "metric", Value: 1.0})
	if err1 != nil {
		t.Fatalf("expected first enqueue to succeed, got: %v", err1)
	}

	err2 := writer.Enqueue(TelemetryRecord{Time: time.Now(), PhysicalAssetID: "asset-2", MetricName: "metric", Value: 2.0})
	if err2 != ErrBufferFull {
		t.Fatalf("expected ErrBufferFull on saturated channel, got: %v", err2)
	}

	if writer.DroppedCount() != 1 {
		t.Fatalf("expected DroppedCount == 1 during buffer saturation, got: %d", writer.DroppedCount())
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	writer.Start(ctx)
	writer.Stop()

	inserted, _ := writer.Stats()
	if inserted != 1 {
		t.Fatalf("expected 1 inserted record, got: %d", inserted)
	}
}

