package server

import (
	"context"
	"fmt"
	"io"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	telemetryv1 "github.com/iamraydoan/factoryos/platform/platform-sdk/go/gen/telemetry/v1"
)

// MockProducer records published batches for testing.
type MockProducer struct {
	batches []*telemetryv1.RecordBatch
	err     error
}

func (m *MockProducer) WriteBatch(ctx context.Context, batch *telemetryv1.RecordBatch) error {
	if m.err != nil {
		return m.err
	}
	m.batches = append(m.batches, batch)
	return nil
}

func (m *MockProducer) Close() error {
	return nil
}

func TestIngestionServer_IngestBatch_Success(t *testing.T) {
	mockProd := &MockProducer{}
	srv := NewIngestionServer(mockProd)

	batch := &telemetryv1.RecordBatch{
		BatchId:      "batch-001",
		EdgeNodeId:   "edge-node-01",
		DispatchedAt: timestamppb.Now(),
		Payloads: []*telemetryv1.TelemetryPayload{
			{
				PhysicalAssetId: "cnc-machine-01",
				EdgeTimestamp:   timestamppb.Now(),
				Readings: []*telemetryv1.SensorReading{
					{MetricName: "temperature_celsius", Value: 68.5, Quality: "GOOD"},
					{MetricName: "vibration_rms", Value: 0.04, Quality: "GOOD"},
				},
			},
		},
	}

	resp, err := srv.IngestBatch(context.Background(), &telemetryv1.IngestBatchRequest{Batch: batch})
	if err != nil {
		t.Fatalf("IngestBatch failed: %v", err)
	}

	result := resp.GetResult()
	if result.GetStatus() != telemetryv1.IngestionStatus_INGESTION_STATUS_ACCEPTED {
		t.Errorf("expected status ACCEPTED, got %v", result.GetStatus())
	}
	if result.GetRecordsIngested() != 2 {
		t.Errorf("expected 2 records ingested, got %d", result.GetRecordsIngested())
	}
	if result.GetBatchId() != "batch-001" {
		t.Errorf("expected batch-001, got %s", result.GetBatchId())
	}

	batches, records, errs := srv.Stats()
	if batches != 1 || records != 2 || errs != 0 {
		t.Errorf("unexpected stats: batches=%d, records=%d, errs=%d", batches, records, errs)
	}
}

func TestIngestionServer_IngestBatch_InvalidRequests(t *testing.T) {
	mockProd := &MockProducer{}
	srv := NewIngestionServer(mockProd)

	// 1. Nil request
	_, err := srv.IngestBatch(context.Background(), nil)
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument for nil request, got %v", err)
	}

	// 2. Missing batch_id
	_, err = srv.IngestBatch(context.Background(), &telemetryv1.IngestBatchRequest{Batch: &telemetryv1.RecordBatch{}})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument for missing batch_id, got %v", err)
	}
}

func TestIngestionServer_IngestBatch_ProducerError(t *testing.T) {
	mockProd := &MockProducer{err: fmt.Errorf("kafka broker connection timeout")}
	srv := NewIngestionServer(mockProd)

	batch := &telemetryv1.RecordBatch{
		BatchId:    "batch-err",
		EdgeNodeId: "edge-01",
	}

	resp, err := srv.IngestBatch(context.Background(), &telemetryv1.IngestBatchRequest{Batch: batch})
	if err == nil {
		t.Fatalf("expected error when producer fails, got nil")
	}
	if resp.GetResult().GetStatus() != telemetryv1.IngestionStatus_INGESTION_STATUS_REJECTED {
		t.Errorf("expected REJECTED status, got %v", resp.GetResult().GetStatus())
	}
}

// MockStreamServer simulates a gRPC client-side stream.
type MockStreamServer struct {
	grpc.ServerStream
	batches []*telemetryv1.RecordBatch
	idx     int
	resp    *telemetryv1.StreamTelemetryResponse
	ctx     context.Context
}

func (m *MockStreamServer) Context() context.Context {
	if m.ctx != nil {
		return m.ctx
	}
	return context.Background()
}

func (m *MockStreamServer) Recv() (*telemetryv1.StreamTelemetryRequest, error) {
	if m.idx >= len(m.batches) {
		return nil, io.EOF
	}
	b := m.batches[m.idx]
	m.idx++
	return &telemetryv1.StreamTelemetryRequest{Batch: b}, nil
}

func (m *MockStreamServer) SendAndClose(resp *telemetryv1.StreamTelemetryResponse) error {
	m.resp = resp
	return nil
}

func TestIngestionServer_StreamTelemetry(t *testing.T) {
	mockProd := &MockProducer{}
	srv := NewIngestionServer(mockProd)

	stream := &MockStreamServer{
		batches: []*telemetryv1.RecordBatch{
			{
				BatchId: "batch-s1",
				Payloads: []*telemetryv1.TelemetryPayload{
					{Readings: []*telemetryv1.SensorReading{{MetricName: "temp", Value: 50}}},
				},
			},
			{
				BatchId: "batch-s2",
				Payloads: []*telemetryv1.TelemetryPayload{
					{Readings: []*telemetryv1.SensorReading{{MetricName: "vib", Value: 0.1}}},
				},
			},
		},
	}

	if err := srv.StreamTelemetry(stream); err != nil {
		t.Fatalf("StreamTelemetry failed: %v", err)
	}

	if stream.resp == nil {
		t.Fatalf("expected stream response, got nil")
	}
	if stream.resp.GetResult().GetRecordsIngested() != 2 {
		t.Errorf("expected 2 records, got %d", stream.resp.GetResult().GetRecordsIngested())
	}
	if len(mockProd.batches) != 2 {
		t.Errorf("expected 2 batches written to producer, got %d", len(mockProd.batches))
	}
}

func TestIngestionServer_Stats(t *testing.T) {
	srv := NewIngestionServer(&MockProducer{})
	b, r, e := srv.Stats()
	if b != 0 || r != 0 || e != 0 {
		t.Errorf("expected zero stats, got b=%d, r=%d, e=%d", b, r, e)
	}
	_ = time.Second
}
