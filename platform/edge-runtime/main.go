package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/iamraydoan/factoryos/platform/edge-runtime/buffer"
	"github.com/iamraydoan/factoryos/platform/edge-runtime/collector"
	"github.com/iamraydoan/factoryos/platform/edge-runtime/forwarder"
	"github.com/iamraydoan/factoryos/platform/edge-runtime/mqtt"
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
// The returned WaitGroup tracks all background goroutines; the caller must
// wait on it after cancelling the context to ensure clean shutdown.
func run(ctx context.Context, dbPath string) (*sync.WaitGroup, error) {
	log.Printf("Initializing local SQLite store-and-forward buffer at %s...", dbPath)
	buf, err := buffer.NewSQLiteBuffer(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize SQLite buffer: %w", err)
	}

	var publisher collector.CloudPublisher
	gatewayURL := os.Getenv("INGESTION_GATEWAY_URL")
	if gatewayURL == "" {
		gatewayURL = "localhost:50051"
	}
	edgeNodeID := os.Getenv("EDGE_NODE_ID")
	if edgeNodeID == "" {
		edgeNodeID = "factory-edge-site-01"
	}

	grpcFwd, err := forwarder.NewGRPCForwarder(forwarder.GRPCForwarderConfig{
		TargetAddress: gatewayURL,
		EdgeNodeID:    edgeNodeID,
		Timeout:       5 * time.Second,
	})
	if err != nil {
		log.Printf("Warning: Failed to create gRPC forwarder: %v (falling back to console publisher)", err)
		publisher = &ConsoleCloudPublisher{}
	} else {
		defer grpcFwd.Close()
		publisher = grpcFwd
	}

	telCollector := collector.NewTelemetryCollector(buf, publisher)

	var wg sync.WaitGroup

	// Start background cloud sync worker (every 2s, batch size up to 50)
	wg.Add(1)
	go func() {
		defer wg.Done()
		telCollector.StartSyncWorker(ctx, 2*time.Second, 50)
	}()

	// Initialize MQTT Subscriber
	brokerURL := os.Getenv("MQTT_BROKER_URL")
	if brokerURL == "" {
		brokerURL = "tcp://localhost:1883"
	}

	mqttSub, err := mqtt.NewMQTTSubscriber(brokerURL, "factoryos-edge-runtime", telCollector)
	if err != nil {
		log.Printf("Warning: Could not initialize MQTT subscriber: %v", err)
	} else {
		// Run MQTT connect in background so offline broker doesn't block startup or unit tests
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := mqttSub.Connect(); err != nil {
				log.Printf("Warning: MQTT Broker at %s unreachable (%v). Subscriber retrying in background...", brokerURL, err)
			}
		}()
	}

	return &wg, nil
}

func main() {
	log.Println("FactoryOS Edge Runtime starting...")

	dbPath := os.Getenv("EDGE_SQLITE_PATH")
	if dbPath == "" {
		dbPath = "edge_buffer.db"
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg, err := run(ctx, dbPath)
	if err != nil {
		log.Fatalf("Runtime startup error: %v", err)
	}

	log.Println("FactoryOS Edge Runtime successfully running. Press Ctrl+C to terminate.")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down FactoryOS Edge Runtime gracefully...")
	cancel()
	wg.Wait()
	log.Println("FactoryOS Edge Runtime stopped.")
}
