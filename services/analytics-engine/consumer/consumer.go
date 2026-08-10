package consumer

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
	_ "github.com/segmentio/kafka-go/snappy"
	"google.golang.org/protobuf/proto"

	telemetryv1 "github.com/iamraydoan/factoryos/platform/platform-sdk/go/gen/telemetry/v1"
	"github.com/iamraydoan/factoryos/services/analytics-engine/config"
	"github.com/iamraydoan/factoryos/services/analytics-engine/db"
	"github.com/iamraydoan/factoryos/services/analytics-engine/processor"
)

// MessageReader defines the interface for reading Kafka messages.
type MessageReader interface {
	ReadMessage(ctx context.Context) (kafka.Message, error)
	Close() error
}

// TelemetryConsumer consumes telemetry messages from Kafka, decodes Protobuf payloads,
// runs real-time alert evaluations, and enqueues records into the BatchWriter.
type TelemetryConsumer struct {
	reader       MessageReader
	writer       *db.BatchWriter
	evaluator    *processor.AlertEvaluator
	retryBackoff time.Duration

	done chan struct{}
	wg   sync.WaitGroup

	messagesConsumed int64
	readingsParsed   int64
	errorsCount      int64
	mu               sync.Mutex
}

// NewTelemetryConsumer creates a consumer with validated dependencies.
func NewTelemetryConsumer(reader MessageReader, writer *db.BatchWriter, evaluator *processor.AlertEvaluator, retryBackoff time.Duration) *TelemetryConsumer {
	if evaluator == nil {
		evaluator = processor.NewAlertEvaluator(nil, nil)
	}
	return &TelemetryConsumer{
		reader:       reader,
		writer:       writer,
		evaluator:    evaluator,
		retryBackoff: retryBackoff,
		done:         make(chan struct{}),
	}
}

// NewKafkaReader initializes a production Kafka Reader from validated KafkaConfig.
func NewKafkaReader(cfg config.KafkaConfig) *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:        cfg.Brokers,
		Topic:          cfg.Topic,
		GroupID:        cfg.GroupID,
		MinBytes:       cfg.MinBytes,
		MaxBytes:       cfg.MaxBytes,
		MaxWait:        cfg.MaxWait,
		CommitInterval: cfg.CommitInterval,
		StartOffset:    kafka.LastOffset,
	})
}

// Start begins the consumer loop in a separate goroutine.
// The provided context controls the lifetime of the consumer loop.
func (c *TelemetryConsumer) Start(parentCtx context.Context) {
	c.wg.Add(1)
	go c.consumeLoop(parentCtx)
	log.Println("[TelemetryConsumer] Ingestion consumer started.")
}

func (c *TelemetryConsumer) consumeLoop(ctx context.Context) {
	defer c.wg.Done()

	for {
		select {
		case <-c.done:
			log.Println("[TelemetryConsumer] Worker received stop signal.")
			return
		case <-ctx.Done():
			log.Println("[TelemetryConsumer] Context cancelled.")
			return
		default:
			msg, err := c.reader.ReadMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				c.mu.Lock()
				c.errorsCount++
				c.mu.Unlock()
				log.Printf("[TelemetryConsumer][ERROR] Failed to read message from Kafka: %v", err)
				select {
				case <-ctx.Done():
					return
				case <-c.done:
					return
				case <-time.After(c.retryBackoff):
					continue
				}
			}

			c.mu.Lock()
			c.messagesConsumed++
			c.mu.Unlock()

			if err := c.ProcessPayload(msg.Value); err != nil {
				c.mu.Lock()
				c.errorsCount++
				c.mu.Unlock()
				log.Printf("[TelemetryConsumer][WARN] Malformed payload received: %v", err)
			}
		}
	}
}

// ProcessPayload unmarshals Protobuf bytes, evaluates alarms, and feeds the batch writer.
func (c *TelemetryConsumer) ProcessPayload(data []byte) error {
	var payload telemetryv1.TelemetryPayload
	if err := proto.Unmarshal(data, &payload); err != nil {
		return fmt.Errorf("protobuf unmarshal failed: %w", err)
	}

	assetID := payload.GetPhysicalAssetId()
	if assetID == "" {
		return fmt.Errorf("missing physical_asset_id in payload")
	}

	var readingTime time.Time
	if payload.GetEdgeTimestamp() != nil {
		readingTime = payload.GetEdgeTimestamp().AsTime()
	} else {
		readingTime = time.Now().UTC()
	}

	for _, reading := range payload.GetReadings() {
		metricName := reading.GetMetricName()
		val := reading.GetValue()
		quality := processor.SensorQuality(reading.GetQuality())
		if quality == "" {
			quality = processor.QualityGood
		}

		// 1. Real-time alert evaluation in memory
		c.evaluator.EvaluateReading(assetID, metricName, val, quality, readingTime)

		// 2. Feed into high-throughput TimescaleDB batch writer
		if c.writer != nil {
			if err := c.writer.Enqueue(db.TelemetryRecord{
				Time:            readingTime,
				PhysicalAssetID: assetID,
				MetricName:      metricName,
				Value:           val,
				Quality:         string(quality),
			}); err != nil {
				// ErrBufferFull — record was dropped; consumer already logged it.
				// Do not count as a processing error; it is a back-pressure event.
				_ = err
			}
		}

		c.mu.Lock()
		c.readingsParsed++
		c.mu.Unlock()
	}

	return nil
}

// Stop signals the consumer to finish and waits for the goroutine to exit.
// The parent context cancellation (passed to Start) is the primary shutdown trigger;
// Stop additionally closes the done channel and the Kafka reader.
func (c *TelemetryConsumer) Stop() {
	close(c.done)
	if c.reader != nil {
		_ = c.reader.Close()
	}
	c.wg.Wait()
	log.Println("[TelemetryConsumer] Consumer stopped gracefully.")
}

// Stats returns processing statistics.
func (c *TelemetryConsumer) Stats() (messages, readings, errors int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.messagesConsumed, c.readingsParsed, c.errorsCount
}
