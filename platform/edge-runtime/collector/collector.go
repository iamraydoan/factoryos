package collector

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/iamraydoan/factoryos/platform/edge-runtime/buffer"
	telemetryv1 "github.com/iamraydoan/factoryos/platform/platform-sdk/go/gen/telemetry/v1"
)

// CloudPublisher interface abstracts telemetry payload transmission to the cloud control plane.
type CloudPublisher interface {
	PublishBatch(ctx context.Context, payloads []*telemetryv1.TelemetryPayload) error
}

// TelemetryCollector coordinates sensor data collection, local SQLite buffering, and cloud sync.
type TelemetryCollector struct {
	buffer    *buffer.SQLiteBuffer
	publisher CloudPublisher
}

// NewTelemetryCollector creates a new TelemetryCollector instance.
func NewTelemetryCollector(buf *buffer.SQLiteBuffer, pub CloudPublisher) *TelemetryCollector {
	return &TelemetryCollector{
		buffer:    buf,
		publisher: pub,
	}
}

// RecordMetric receives a raw sensor metric reading and enqueues it to the local store-and-forward buffer.
func (c *TelemetryCollector) RecordMetric(ctx context.Context, physicalAssetID string, metricName string, value float64, quality string) error {
	payload := &telemetryv1.TelemetryPayload{
		PhysicalAssetId: physicalAssetID,
		EdgeTimestamp:   timestamppb.Now(),
		Readings: []*telemetryv1.SensorReading{
			{
				MetricName: metricName,
				Value:      value,
				Quality:    quality,
			},
		},
	}
	_, err := c.buffer.Enqueue(ctx, payload)
	return err
}

// RecordBatch receives multiple metric readings for an asset and enqueues them as a single batch payload.
func (c *TelemetryCollector) RecordBatch(ctx context.Context, physicalAssetID string, readings map[string]float64) error {
	var sensorReadings []*telemetryv1.SensorReading
	for metricName, val := range readings {
		sensorReadings = append(sensorReadings, &telemetryv1.SensorReading{
			MetricName: metricName,
			Value:      val,
			Quality:    "GOOD",
		})
	}

	payload := &telemetryv1.TelemetryPayload{
		PhysicalAssetId: physicalAssetID,
		EdgeTimestamp:   timestamppb.Now(),
		Readings:        sensorReadings,
	}

	_, err := c.buffer.Enqueue(ctx, payload)
	return err
}

// FlushBatch dequeues pending records from SQLite, attempts cloud transmission, and clears successfully sent records.
func (c *TelemetryCollector) FlushBatch(ctx context.Context, limit int) (int, error) {
	records, err := c.buffer.DequeueBatch(ctx, limit)
	if err != nil {
		return 0, fmt.Errorf("failed to dequeue batch from sqlite buffer: %w", err)
	}

	if len(records) == 0 {
		return 0, nil
	}

	var payloads []*telemetryv1.TelemetryPayload
	var recordIDs []int64

	for _, rec := range records {
		var payload telemetryv1.TelemetryPayload
		if err := proto.Unmarshal(rec.PayloadBytes, &payload); err != nil {
			log.Printf("corrupted payload in record %d, skipping: %v", rec.ID, err)
			continue
		}
		payloads = append(payloads, &payload)
		recordIDs = append(recordIDs, rec.ID)
	}

	if len(payloads) == 0 {
		return 0, nil
	}

	// Attempt cloud publication
	if err := c.publisher.PublishBatch(ctx, payloads); err != nil {
		// Log network / cloud connection failure and keep records in buffer for retry
		_ = c.buffer.IncrementRetryCount(ctx, recordIDs)
		return 0, fmt.Errorf("cloud sync failed, records retained in store-and-forward buffer: %w", err)
	}

	// Cloud sync succeeded, mark records as sent
	if err := c.buffer.MarkSent(ctx, recordIDs); err != nil {
		return len(payloads), fmt.Errorf("failed to mark records as sent: %w", err)
	}

	return len(payloads), nil
}

// StartSyncWorker runs a periodic background worker loop to flush store-and-forward buffer records to the cloud.
func (c *TelemetryCollector) StartSyncWorker(ctx context.Context, interval time.Duration, batchLimit int) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Sync worker stopping...")
			return
		case <-ticker.C:
			sentCount, err := c.FlushBatch(ctx, batchLimit)
			if err != nil {
				log.Printf("[Edge Runtime Offline Mode] %v", err)
			} else if sentCount > 0 {
				log.Printf("[Edge Runtime Online] Successfully synced %d telemetry batch payloads to cloud", sentCount)
			}
		}
	}
}
