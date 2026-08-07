package buffer

import (
	"context"
	"testing"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	telemetryv1 "github.com/iamraydoan/factoryos/platform/platform-sdk/go/gen/telemetry/v1"
)

func newTestBuffer(t *testing.T, dbName string) *SQLiteBuffer {
	buf, err := NewSQLiteBuffer("file:" + dbName + "?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("failed to create sqlite memory buffer: %v", err)
	}
	return buf
}

func TestSQLiteBuffer_Enqueue(t *testing.T) {
	ctx := context.Background()
	buf := newTestBuffer(t, "db_enqueue")
	defer buf.Close()

	payload := &telemetryv1.TelemetryPayload{
		PhysicalAssetId: "asset-enqueue-01",
		EdgeTimestamp:   timestamppb.Now(),
		Readings: []*telemetryv1.SensorReading{
			{MetricName: "temp", Value: 36.6, Quality: "GOOD"},
		},
	}

	id, err := buf.Enqueue(ctx, payload)
	if err != nil {
		t.Fatalf("Enqueue failed: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected positive insert id, got %d", id)
	}

	count, err := buf.GetPendingCount(ctx)
	if err != nil {
		t.Fatalf("GetPendingCount failed: %v", err)
	}
	if count != 1 {
		t.Errorf("expected pending count 1, got %d", count)
	}
}

func TestSQLiteBuffer_Enqueue_NilTimestamp(t *testing.T) {
	ctx := context.Background()
	buf := newTestBuffer(t, "db_enqueue_nil_ts")
	defer buf.Close()

	// Enqueue with nil EdgeTimestamp -> automatically sets timestamp
	payload := &telemetryv1.TelemetryPayload{
		PhysicalAssetId: "asset-nil-ts",
		EdgeTimestamp:   nil,
	}

	id, err := buf.Enqueue(ctx, payload)
	if err != nil {
		t.Fatalf("Enqueue failed with nil timestamp: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected positive insert id, got %d", id)
	}
}

func TestSQLiteBuffer_DequeueBatch(t *testing.T) {
	ctx := context.Background()
	buf := newTestBuffer(t, "db_dequeue")
	defer buf.Close()

	payload := &telemetryv1.TelemetryPayload{
		PhysicalAssetId: "asset-dequeue-01",
		EdgeTimestamp:   timestamppb.Now(),
		Readings: []*telemetryv1.SensorReading{
			{MetricName: "vibration", Value: 0.05, Quality: "GOOD"},
		},
	}

	id, err := buf.Enqueue(ctx, payload)
	if err != nil {
		t.Fatalf("Enqueue failed: %v", err)
	}

	records, err := buf.DequeueBatch(ctx, 10)
	if err != nil {
		t.Fatalf("DequeueBatch failed: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}

	rec := records[0]
	if rec.ID != id || rec.PhysicalAssetID != "asset-dequeue-01" {
		t.Errorf("unexpected record data: %+v", rec)
	}

	var unmarshaled telemetryv1.TelemetryPayload
	if err := proto.Unmarshal(rec.PayloadBytes, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal payload bytes: %v", err)
	}
	if unmarshaled.Readings[0].MetricName != "vibration" {
		t.Errorf("expected metric 'vibration', got %s", unmarshaled.Readings[0].MetricName)
	}
}

func TestSQLiteBuffer_IncrementRetryCount(t *testing.T) {
	ctx := context.Background()
	buf := newTestBuffer(t, "db_retry")
	defer buf.Close()

	payload := &telemetryv1.TelemetryPayload{PhysicalAssetId: "asset-retry-01"}
	id, err := buf.Enqueue(ctx, payload)
	if err != nil {
		t.Fatalf("Enqueue failed: %v", err)
	}

	if err := buf.IncrementRetryCount(ctx, []int64{id}); err != nil {
		t.Fatalf("IncrementRetryCount failed: %v", err)
	}

	records, err := buf.DequeueBatch(ctx, 1)
	if err != nil {
		t.Fatalf("DequeueBatch failed: %v", err)
	}
	if records[0].RetryCount != 1 {
		t.Errorf("expected RetryCount 1, got %d", records[0].RetryCount)
	}
}

func TestSQLiteBuffer_MarkSent(t *testing.T) {
	ctx := context.Background()
	buf := newTestBuffer(t, "db_mark_sent")
	defer buf.Close()

	payload := &telemetryv1.TelemetryPayload{PhysicalAssetId: "asset-sent-01"}
	id, err := buf.Enqueue(ctx, payload)
	if err != nil {
		t.Fatalf("Enqueue failed: %v", err)
	}

	if err := buf.MarkSent(ctx, []int64{id}); err != nil {
		t.Fatalf("MarkSent failed: %v", err)
	}

	count, err := buf.GetPendingCount(ctx)
	if err != nil {
		t.Fatalf("GetPendingCount failed: %v", err)
	}
	if count != 0 {
		t.Errorf("expected pending count 0 after MarkSent, got %d", count)
	}
}

func TestSQLiteBuffer_EmptyOperations(t *testing.T) {
	ctx := context.Background()
	buf := newTestBuffer(t, "db_empty_ops")
	defer buf.Close()

	if err := buf.IncrementRetryCount(ctx, nil); err != nil {
		t.Errorf("expected nil error for empty IncrementRetryCount, got %v", err)
	}
	if err := buf.MarkSent(ctx, nil); err != nil {
		t.Errorf("expected nil error for empty MarkSent, got %v", err)
	}
}

func TestSQLiteBuffer_InvalidPath(t *testing.T) {
	_, err := NewSQLiteBuffer("/invalid_dir/cannot_create.db")
	if err == nil {
		t.Error("expected error when opening invalid db path, got nil")
	}
}

func TestSQLiteBuffer_ClosedDBErrors(t *testing.T) {
	ctx := context.Background()
	buf := newTestBuffer(t, "db_closed_err")
	buf.Close() // Close database immediately

	payload := &telemetryv1.TelemetryPayload{PhysicalAssetId: "asset-closed"}
	if _, err := buf.Enqueue(ctx, payload); err == nil {
		t.Error("expected error on Enqueue with closed db, got nil")
	}

	if _, err := buf.DequeueBatch(ctx, 10); err == nil {
		t.Error("expected error on DequeueBatch with closed db, got nil")
	}

	if err := buf.MarkSent(ctx, []int64{1}); err == nil {
		t.Error("expected error on MarkSent with closed db, got nil")
	}

	if err := buf.IncrementRetryCount(ctx, []int64{1}); err == nil {
		t.Error("expected error on IncrementRetryCount with closed db, got nil")
	}

	if _, err := buf.GetPendingCount(ctx); err == nil {
		t.Error("expected error on GetPendingCount with closed db, got nil")
	}
}
