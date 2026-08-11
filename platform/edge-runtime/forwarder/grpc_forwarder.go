package forwarder

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"

	telemetryv1 "github.com/iamraydoan/factoryos/platform/platform-sdk/go/gen/telemetry/v1"
)

// GRPCForwarderConfig holds configuration for the Edge gRPC forwarder.
type GRPCForwarderConfig struct {
	TargetAddress string        // e.g. "localhost:50051" or "gateway.factoryos.com:443"
	EdgeNodeID    string        // e.g. "factory-edge-site-01"
	UseTLS        bool          // enable TLS encryption
	Timeout       time.Duration // RPC timeout per batch
}

// GRPCForwarder transmits micro-batches from local SQLite buffer to Cloud Ingestion Gateway via gRPC.
type GRPCForwarder struct {
	cfg        GRPCForwarderConfig
	client     telemetryv1.TelemetryIngestionServiceClient
	conn       *grpc.ClientConn
	mu         sync.Mutex
	batchCount int64
}

// NewGRPCForwarder creates a new forwarder and connects to the Ingestion Gateway.
func NewGRPCForwarder(cfg GRPCForwarderConfig) (*GRPCForwarder, error) {
	if cfg.Timeout <= 0 {
		cfg.Timeout = 5 * time.Second
	}
	if cfg.EdgeNodeID == "" {
		cfg.EdgeNodeID = "edge-node-default"
	}

	var dialOpt grpc.DialOption
	if cfg.UseTLS {
		dialOpt = grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{}))
	} else {
		dialOpt = grpc.WithTransportCredentials(insecure.NewCredentials())
	}

	conn, err := grpc.NewClient(cfg.TargetAddress, dialOpt)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC client for %s: %w", cfg.TargetAddress, err)
	}

	client := telemetryv1.NewTelemetryIngestionServiceClient(conn)
	log.Printf("[GRPCForwarder] Connected to Cloud Ingestion Gateway at %s (NodeID: %s, TLS: %v)",
		cfg.TargetAddress, cfg.EdgeNodeID, cfg.UseTLS)

	return &GRPCForwarder{
		cfg:    cfg,
		client: client,
		conn:   conn,
	}, nil
}

// NewGRPCForwarderWithClient initializes a forwarder with a custom/mock client (useful for unit tests).
func NewGRPCForwarderWithClient(client telemetryv1.TelemetryIngestionServiceClient, edgeNodeID string) *GRPCForwarder {
	return &GRPCForwarder{
		client: client,
		cfg: GRPCForwarderConfig{
			EdgeNodeID: edgeNodeID,
			Timeout:    5 * time.Second,
		},
	}
}

// PublishBatch packages payloads into a RecordBatch and transmits it to Cloud Ingestion Gateway.
func (f *GRPCForwarder) PublishBatch(ctx context.Context, payloads []*telemetryv1.TelemetryPayload) error {
	if len(payloads) == 0 {
		return nil
	}

	f.mu.Lock()
	f.batchCount++
	batchID := fmt.Sprintf("%s-batch-%d-%d", f.cfg.EdgeNodeID, time.Now().UnixNano(), f.batchCount)
	f.mu.Unlock()

	recordBatch := &telemetryv1.RecordBatch{
		BatchId:      batchID,
		EdgeNodeId:   f.cfg.EdgeNodeID,
		DispatchedAt: timestamppb.Now(),
		Payloads:     payloads,
	}

	rpcCtx, cancel := context.WithTimeout(ctx, f.cfg.Timeout)
	defer cancel()

	resp, err := f.client.IngestBatch(rpcCtx, &telemetryv1.IngestBatchRequest{Batch: recordBatch})
	if err != nil {
		return fmt.Errorf("gRPC IngestBatch RPC failed to %s: %w", f.cfg.TargetAddress, err)
	}

	result := resp.GetResult()
	if result.GetStatus() == telemetryv1.IngestionStatus_INGESTION_STATUS_REJECTED {
		return fmt.Errorf("cloud rejected telemetry batch %s: %s", batchID, result.GetMessage())
	}

	return nil
}

// Close gracefully terminates the gRPC connection.
func (f *GRPCForwarder) Close() error {
	if f.conn != nil {
		return f.conn.Close()
	}
	return nil
}
