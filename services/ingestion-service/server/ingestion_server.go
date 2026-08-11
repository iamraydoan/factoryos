package server

import (
	"context"
	"fmt"
	"io"
	"log"
	"sync"
	"sync/atomic"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	telemetryv1 "github.com/iamraydoan/factoryos/platform/platform-sdk/go/gen/telemetry/v1"
	"github.com/iamraydoan/factoryos/services/ingestion-service/kafka"
)

// IngestionServer implements the gRPC TelemetryIngestionServiceServer interface.
type IngestionServer struct {
	telemetryv1.UnimplementedTelemetryIngestionServiceServer
	producer kafka.TelemetryProducer

	batchesReceived int64
	recordsReceived int64
	errorsCount     int64
	mu              sync.Mutex
}

// NewIngestionServer creates a new IngestionServer instance with the provided Kafka producer.
func NewIngestionServer(producer kafka.TelemetryProducer) *IngestionServer {
	return &IngestionServer{
		producer: producer,
	}
}

// IngestBatch handles single IngestBatchRequest unary RPC calls from Edge Nodes.
func (s *IngestionServer) IngestBatch(ctx context.Context, req *telemetryv1.IngestBatchRequest) (*telemetryv1.IngestBatchResponse, error) {
	if req == nil || req.GetBatch() == nil {
		atomic.AddInt64(&s.errorsCount, 1)
		return nil, status.Error(codes.InvalidArgument, "IngestBatchRequest or batch cannot be nil")
	}

	batch := req.GetBatch()

	if batch.GetBatchId() == "" {
		atomic.AddInt64(&s.errorsCount, 1)
		return nil, status.Error(codes.InvalidArgument, "batch_id is required")
	}

	payloads := batch.GetPayloads()
	var totalReadings int64
	for _, p := range payloads {
		totalReadings += int64(len(p.GetReadings()))
	}

	if err := s.producer.WriteBatch(ctx, batch); err != nil {
		atomic.AddInt64(&s.errorsCount, 1)
		log.Printf("[IngestionServer][ERROR] Failed to produce batch %s to Kafka: %v", batch.GetBatchId(), err)
		return &telemetryv1.IngestBatchResponse{
			Result: &telemetryv1.IngestTelemetryResponse{
				Status:          telemetryv1.IngestionStatus_INGESTION_STATUS_REJECTED,
				BatchId:         batch.GetBatchId(),
				RecordsIngested: 0,
				Message:         fmt.Sprintf("internal error producing to event bus: %v", err),
				AcknowledgedAt:  timestamppb.Now(),
			},
		}, status.Errorf(codes.Internal, "failed to persist batch: %v", err)
	}

	atomic.AddInt64(&s.batchesReceived, 1)
	atomic.AddInt64(&s.recordsReceived, totalReadings)

	return &telemetryv1.IngestBatchResponse{
		Result: &telemetryv1.IngestTelemetryResponse{
			Status:          telemetryv1.IngestionStatus_INGESTION_STATUS_ACCEPTED,
			BatchId:         batch.GetBatchId(),
			RecordsIngested: totalReadings,
			Message:         "Batch accepted and enqueued to Kafka event stream",
			AcknowledgedAt:  timestamppb.Now(),
		},
	}, nil
}

// StreamTelemetry handles persistent client-side telemetry streaming from Edge Nodes.
func (s *IngestionServer) StreamTelemetry(stream telemetryv1.TelemetryIngestionService_StreamTelemetryServer) error {
	var totalStreamReadings int64
	var lastBatchID string

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			// Stream completed by client, send summary response
			return stream.SendAndClose(&telemetryv1.StreamTelemetryResponse{
				Result: &telemetryv1.IngestTelemetryResponse{
					Status:          telemetryv1.IngestionStatus_INGESTION_STATUS_ACCEPTED,
					BatchId:         lastBatchID,
					RecordsIngested: totalStreamReadings,
					Message:         "Telemetry stream ingested successfully",
					AcknowledgedAt:  timestamppb.Now(),
				},
			})
		}
		if err != nil {
			atomic.AddInt64(&s.errorsCount, 1)
			return status.Errorf(codes.Unknown, "error reading telemetry stream: %v", err)
		}

		if req == nil || req.GetBatch() == nil {
			continue
		}

		batch := req.GetBatch()
		lastBatchID = batch.GetBatchId()
		var batchReadings int64
		for _, p := range batch.GetPayloads() {
			batchReadings += int64(len(p.GetReadings()))
		}

		if err := s.producer.WriteBatch(stream.Context(), batch); err != nil {
			atomic.AddInt64(&s.errorsCount, 1)
			log.Printf("[IngestionServer][ERROR] Failed streaming batch %s to Kafka: %v", lastBatchID, err)
			return status.Errorf(codes.Internal, "failed to write stream batch to Kafka: %v", err)
		}

		atomic.AddInt64(&s.batchesReceived, 1)
		atomic.AddInt64(&s.recordsReceived, batchReadings)
		totalStreamReadings += batchReadings
	}
}

// Stats returns the real-time counters for processed batches, records, and errors.
func (s *IngestionServer) Stats() (batches int64, records int64, errs int64) {
	return atomic.LoadInt64(&s.batchesReceived), atomic.LoadInt64(&s.recordsReceived), atomic.LoadInt64(&s.errorsCount)
}
