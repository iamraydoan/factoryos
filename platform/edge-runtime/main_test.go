package main

import (
	"context"
	"testing"
	"time"

	telemetryv1 "github.com/iamraydoan/factoryos/platform/platform-sdk/go/gen/telemetry/v1"
)

func TestConsoleCloudPublisher_PublishBatch(t *testing.T) {
	pub := &ConsoleCloudPublisher{}
	ctx := context.Background()

	payloads := []*telemetryv1.TelemetryPayload{
		{
			PhysicalAssetId: "test-console-asset",
			Readings: []*telemetryv1.SensorReading{
				{MetricName: "voltage", Value: 220.0, Quality: "GOOD"},
			},
		},
	}

	if err := pub.PublishBatch(ctx, payloads); err != nil {
		t.Fatalf("ConsoleCloudPublisher.PublishBatch failed: %v", err)
	}
}

func TestRun_Success(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dbPath := "file:main_test_db?mode=memory&cache=shared"
	if err := run(ctx, dbPath); err != nil {
		t.Fatalf("run failed: %v", err)
	}

	time.Sleep(50 * time.Millisecond)
}

func TestRun_InvalidDBPath(t *testing.T) {
	ctx := context.Background()
	invalidDbPath := "/invalid_dir/cannot_create.db"

	if err := run(ctx, invalidDbPath); err == nil {
		t.Error("expected error when running with invalid DB path, got nil")
	}
}
