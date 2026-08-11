package kafka

import (
	"context"
	"fmt"
	"log"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/snappy"
	"google.golang.org/protobuf/proto"

	telemetryv1 "github.com/iamraydoan/factoryos/platform/platform-sdk/go/gen/telemetry/v1"
	"github.com/iamraydoan/factoryos/services/ingestion-service/config"
)

// TelemetryProducer abstracts the publishing of RecordBatches to Kafka.
type TelemetryProducer interface {
	WriteBatch(ctx context.Context, batch *telemetryv1.RecordBatch) error
	Close() error
}

// KafkaBatchProducer produces Protobuf RecordBatches to an internal Kafka topic.
type KafkaBatchProducer struct {
	writer *kafka.Writer
	topic  string
}

// NewKafkaProducer creates a new KafkaBatchProducer from configuration.
func NewKafkaProducer(cfg config.KafkaConfig) *KafkaBatchProducer {
	writer := &kafka.Writer{
		Addr:                   kafka.TCP(cfg.Brokers...),
		Topic:                  cfg.Topic,
		Balancer:               &kafka.LeastBytes{},
		BatchSize:              cfg.BatchSize,
		BatchTimeout:           cfg.BatchTimeout,
		RequiredAcks:           kafka.RequiredAcks(cfg.RequiredAcks),
		Async:                  cfg.Async,
		Compression:            kafka.Snappy,
		AllowAutoTopicCreation: true,
	}

	// Register Snappy codec
	_ = snappy.NewCompressionCodec()

	log.Printf("[KafkaProducer] Initialized producer on brokers: %v | Topic: %s (Snappy compression)", cfg.Brokers, cfg.Topic)
	return &KafkaBatchProducer{
		writer: writer,
		topic:  cfg.Topic,
	}
}

// WriteBatch serializes a Protobuf RecordBatch and produces it to Kafka.
func (p *KafkaBatchProducer) WriteBatch(ctx context.Context, batch *telemetryv1.RecordBatch) error {
	if batch == nil {
		return fmt.Errorf("cannot write nil RecordBatch")
	}

	data, err := proto.Marshal(batch)
	if err != nil {
		return fmt.Errorf("failed to marshal RecordBatch: %w", err)
	}

	key := batch.GetEdgeNodeId()
	if key == "" {
		key = batch.GetBatchId()
	}

	msg := kafka.Message{
		Key:   []byte(key),
		Value: data,
		Headers: []kafka.Header{
			{Key: "Content-Type", Value: []byte("application/x-protobuf")},
			{Key: "Batch-ID", Value: []byte(batch.GetBatchId())},
			{Key: "Edge-Node-ID", Value: []byte(batch.GetEdgeNodeId())},
		},
	}

	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		return fmt.Errorf("failed to produce message to Kafka topic %s: %w", p.topic, err)
	}

	return nil
}

// Close closes the underlying Kafka writer.
func (p *KafkaBatchProducer) Close() error {
	if p.writer != nil {
		return p.writer.Close()
	}
	return nil
}
