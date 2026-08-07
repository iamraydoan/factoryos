package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/iamraydoan/factoryos/platform/edge-runtime/buffer"
	"github.com/iamraydoan/factoryos/platform/edge-runtime/collector"
	telemetryv1 "github.com/iamraydoan/factoryos/platform/platform-sdk/go/gen/telemetry/v1"
)

// ConsoleCloudPublisher is a demonstration publisher logging telemetry to stdout.
type ConsoleCloudPublisher struct{}

func (p *ConsoleCloudPublisher) PublishBatch(ctx context.Context, payloads []*telemetryv1.TelemetryPayload) error {
	for _, payload := range payloads {
		fmt.Printf("[Cloud Publisher] Transmitted telemetry from asset=%s (%d readings)\n", payload.PhysicalAssetId, len(payload.Readings))
	}
	return nil
}

// run initializes and starts the Edge Runtime service components.
func run(ctx context.Context, dbPath string) error {
	log.Printf("Initializing local SQLite store-and-forward buffer at %s...", dbPath)
	buf, err := buffer.NewSQLiteBuffer(dbPath)
	if err != nil {
		return fmt.Errorf("failed to initialize SQLite buffer: %w", err)
	}

	publisher := &ConsoleCloudPublisher{}
	telCollector := collector.NewTelemetryCollector(buf, publisher)

	// Start background cloud sync worker
	go telCollector.StartSyncWorker(ctx, 5*time.Second, 50)
	return nil
}

func main() {
	log.Println("FactoryOS Edge Runtime starting...")

	dbPath := os.Getenv("EDGE_SQLITE_PATH")
	if dbPath == "" {
		dbPath = "edge_buffer.db"
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := run(ctx, dbPath); err != nil {
		log.Fatalf("Runtime startup error: %v", err)
	}

	log.Println("FactoryOS Edge Runtime successfully running. Press Ctrl+C to terminate.")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down FactoryOS Edge Runtime gracefully...")
}
