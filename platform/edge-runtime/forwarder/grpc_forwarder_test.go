package forwarder

import (
	"context"
	"fmt"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"

	telemetryv1 "github.com/iamraydoan/factoryos/platform/platform-sdk/go/gen/telemetry/v1"
)

type MockIngestionClient struct {
	lastBatch *telemetryv1.RecordBatch
	err       error
	status    telemetryv1.IngestionStatus
}

func (m *MockIngestionClient) IngestBatch(ctx context.Context, in *telemetryv1.IngestBatchRequest, opts ...grpc.CallOption) (*telemetryv1.IngestBatchResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	batch := in.GetBatch()
	m.lastBatch = batch
	st := m.status
	if st == telemetryv1.IngestionStatus_INGESTION_STATUS_UNSPECIFIED {
		st = telemetryv1.IngestionStatus_INGESTION_STATUS_ACCEPTED
	}
	return &telemetryv1.IngestBatchResponse{
		Result: &telemetryv1.IngestTelemetryResponse{
			Status:          st,
			BatchId:         batch.GetBatchId(),
			RecordsIngested: int64(len(batch.GetPayloads())),
		},
	}, nil
}

func (m *MockIngestionClient) StreamTelemetry(ctx context.Context, opts ...grpc.CallOption) (telemetryv1.TelemetryIngestionService_StreamTelemetryClient, error) {
	return nil, nil
}

func TestGRPCForwarder_PublishBatch_Success(t *testing.T) {
	mockClient := &MockIngestionClient{}
	forwarder := NewGRPCForwarderWithClient(mockClient, "edge-site-alpha")

	payloads := []*telemetryv1.TelemetryPayload{
		{
			PhysicalAssetId: "asset-01",
			EdgeTimestamp:   timestamppb.Now(),
			Readings: []*telemetryv1.SensorReading{
				{MetricName: "temp", Value: 70.0, Quality: "GOOD"},
			},
		},
		{
			PhysicalAssetId: "asset-02",
			EdgeTimestamp:   timestamppb.Now(),
			Readings: []*telemetryv1.SensorReading{
				{MetricName: "vibe", Value: 0.05, Quality: "GOOD"},
			},
		},
	}

	err := forwarder.PublishBatch(context.Background(), payloads)
	if err != nil {
		t.Fatalf("PublishBatch failed: %v", err)
	}

	if mockClient.lastBatch == nil {
		t.Fatalf("expected batch to be received by mock client")
	}
	if mockClient.lastBatch.GetEdgeNodeId() != "edge-site-alpha" {
		t.Errorf("expected edge-site-alpha, got %s", mockClient.lastBatch.GetEdgeNodeId())
	}
	if len(mockClient.lastBatch.GetPayloads()) != 2 {
		t.Errorf("expected 2 payloads in batch, got %d", len(mockClient.lastBatch.GetPayloads()))
	}
}

func TestGRPCForwarder_PublishBatch_Empty(t *testing.T) {
	mockClient := &MockIngestionClient{}
	forwarder := NewGRPCForwarderWithClient(mockClient, "edge-site-alpha")

	err := forwarder.PublishBatch(context.Background(), nil)
	if err != nil {
		t.Fatalf("PublishBatch failed on empty: %v", err)
	}
	if mockClient.lastBatch != nil {
		t.Errorf("expected no batch dispatched for empty payloads")
	}
}

func TestGRPCForwarder_PublishBatch_NetworkError(t *testing.T) {
	mockClient := &MockIngestionClient{err: fmt.Errorf("connection refused")}
	forwarder := NewGRPCForwarderWithClient(mockClient, "edge-site-alpha")

	payloads := []*telemetryv1.TelemetryPayload{
		{PhysicalAssetId: "asset-01"},
	}

	err := forwarder.PublishBatch(context.Background(), payloads)
	if err == nil {
		t.Fatalf("expected error on network failure, got nil")
	}
}

func TestGRPCForwarder_PublishBatch_CloudRejected(t *testing.T) {
	mockClient := &MockIngestionClient{
		status: telemetryv1.IngestionStatus_INGESTION_STATUS_REJECTED,
	}
	forwarder := NewGRPCForwarderWithClient(mockClient, "edge-site-alpha")

	payloads := []*telemetryv1.TelemetryPayload{
		{PhysicalAssetId: "asset-01"},
	}

	err := forwarder.PublishBatch(context.Background(), payloads)
	if err == nil {
		t.Fatalf("expected error when cloud rejects batch, got nil")
	}
}

func TestGRPCForwarder_NewGRPCForwarder(t *testing.T) {
	f, err := NewGRPCForwarder(GRPCForwarderConfig{
		TargetAddress: "127.0.0.1:50051",
		EdgeNodeID:    "test-node",
		Timeout:       2 * time.Second,
	})
	if err != nil {
		t.Fatalf("NewGRPCForwarder error: %v", err)
	}
	defer f.Close()
}
