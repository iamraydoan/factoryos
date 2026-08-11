package main

import (
	"context"
	"net"
	"os"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/timestamppb"

	telemetryv1 "github.com/iamraydoan/factoryos/platform/platform-sdk/go/gen/telemetry/v1"
	"github.com/iamraydoan/factoryos/services/ingestion-service/server"
)

type mockE2EProducer struct {
	batches []*telemetryv1.RecordBatch
}

func (m *mockE2EProducer) WriteBatch(ctx context.Context, batch *telemetryv1.RecordBatch) error {
	m.batches = append(m.batches, batch)
	return nil
}

func (m *mockE2EProducer) Close() error {
	return nil
}

func TestE2E_TelemetryPipeline(t *testing.T) {
	// If live integration test requested with running daemon:
	grpcAddr := os.Getenv("TEST_GRPC_ADDR")
	var conn *grpc.ClientConn
	var err error

	if grpcAddr != "" {
		conn, err = grpc.NewClient(grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			t.Fatalf("failed to connect to live gRPC at %s: %v", grpcAddr, err)
		}
		defer conn.Close()
	} else {
		// Self-contained in-memory gRPC server
		bufferSize := 1024 * 1024
		lis := bufconn.Listen(bufferSize)
		s := grpc.NewServer()
		mockProd := &mockE2EProducer{}
		srv := server.NewIngestionServer(mockProd)
		telemetryv1.RegisterTelemetryIngestionServiceServer(s, srv)

		go func() {
			_ = s.Serve(lis)
		}()
		defer s.Stop()

		conn, err = grpc.NewClient("passthrough://bufnet",
			grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
				return lis.Dial()
			}),
			grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			t.Fatalf("failed to dial bufnet: %v", err)
		}
		defer conn.Close()
	}

	client := telemetryv1.NewTelemetryIngestionServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	batch := &telemetryv1.RecordBatch{
		BatchId:      "e2e-live-batch-001",
		EdgeNodeId:   "edge-node-live-01",
		DispatchedAt: timestamppb.Now(),
		Payloads: []*telemetryv1.TelemetryPayload{
			{
				PhysicalAssetId: "cnc-machine-01",
				EdgeTimestamp:   timestamppb.Now(),
				Readings: []*telemetryv1.SensorReading{
					{MetricName: "temperature_celsius", Value: 64.2, Quality: "GOOD"},
					{MetricName: "spindle_vibration", Value: 0.035, Quality: "GOOD"},
					{MetricName: "cycle_count", Value: 1250, Quality: "GOOD"},
				},
			},
			{
				PhysicalAssetId: "robotic-arm-02",
				EdgeTimestamp:   timestamppb.Now(),
				Readings: []*telemetryv1.SensorReading{
					{MetricName: "motor_torque_nm", Value: 85.0, Quality: "GOOD"},
					{MetricName: "coolant_flow_rate", Value: 12.5, Quality: "GOOD"},
				},
			},
		},
	}

	resp, err := client.IngestBatch(ctx, batch)
	if err != nil {
		t.Fatalf("E2E IngestBatch failed: %v", err)
	}

	if resp.GetStatus() != telemetryv1.IngestionStatus_INGESTION_STATUS_ACCEPTED {
		t.Fatalf("expected status ACCEPTED, got %v (msg: %s)", resp.GetStatus(), resp.GetMessage())
	}

	if resp.GetRecordsIngested() != 5 {
		t.Fatalf("expected 5 records ingested, got %d", resp.GetRecordsIngested())
	}

	t.Logf("E2E Batch %s successfully accepted by Ingestion Service (%d records)", resp.GetBatchId(), resp.GetRecordsIngested())
}
