package consumer

import (
	"context"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/segmentio/kafka-go"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	telemetryv1 "github.com/iamraydoan/factoryos/platform/platform-sdk/go/gen/telemetry/v1"
	"github.com/iamraydoan/factoryos/services/analytics-engine/config"
	"github.com/iamraydoan/factoryos/services/analytics-engine/db"
	"github.com/iamraydoan/factoryos/services/analytics-engine/processor"
)

type mockReader struct {
	messages []kafka.Message
	index    int
	stopCh   chan struct{}
	once     sync.Once
	mu       sync.Mutex
}

func newMockReader(messages []kafka.Message) *mockReader {
	return &mockReader{
		messages: messages,
		stopCh:   make(chan struct{}),
	}
}

func (m *mockReader) ReadMessage(ctx context.Context) (kafka.Message, error) {
	m.mu.Lock()
	if m.index < len(m.messages) {
		msg := m.messages[m.index]
		m.index++
		m.mu.Unlock()
		return msg, nil
	}
	m.mu.Unlock()

	select {
	case <-ctx.Done():
		return kafka.Message{}, ctx.Err()
	case <-m.stopCh:
		return kafka.Message{}, io.EOF
	}
}

func (m *mockReader) Close() error {
	m.once.Do(func() {
		close(m.stopCh)
	})
	return nil
}

// waitForCondition polls check() every 5ms until it returns true or timeout elapses.
func waitForCondition(t *testing.T, timeout time.Duration, check func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if check() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("condition not met within timeout")
}


func TestTelemetryConsumer_ProcessPayload(t *testing.T) {
	var evaluatedAlerts []processor.TelemetryAlert
	evaluator := processor.NewAlertEvaluator(nil, func(alert processor.TelemetryAlert) {
		evaluatedAlerts = append(evaluatedAlerts, alert)
	})

	consumer := NewTelemetryConsumer(nil, nil, evaluator, 50*time.Millisecond)

	now := time.Now().UTC()
	payload := &telemetryv1.TelemetryPayload{
		PhysicalAssetId: "cnc-machine-01",
		EdgeTimestamp:   timestamppb.New(now),
		Readings: []*telemetryv1.SensorReading{
			{MetricName: "spindle_temp", Value: 99.5, Quality: "GOOD"},
			{MetricName: "motor_current", Value: 14.2, Quality: "GOOD"},
		},
	}

	data, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal proto: %v", err)
	}

	if err := consumer.ProcessPayload(data); err != nil {
		t.Fatalf("failed to process valid payload: %v", err)
	}

	_, readings, errors := consumer.Stats()
	if readings != 2 {
		t.Fatalf("expected 2 readings parsed, got: %d", readings)
	}
	if errors != 0 {
		t.Fatalf("expected 0 errors, got: %d", errors)
	}

	// Overheat alert must have triggered
	if len(evaluatedAlerts) != 1 {
		t.Fatalf("expected 1 overheat alert, got: %d", len(evaluatedAlerts))
	}
	if evaluatedAlerts[0].Severity != processor.SeverityCritical {
		t.Fatalf("expected CRITICAL alert, got: %s", evaluatedAlerts[0].Severity)
	}
}

func TestTelemetryConsumer_MalformedPayload(t *testing.T) {
	consumer := NewTelemetryConsumer(nil, nil, nil, 0)

	// Invalid bytes
	err := consumer.ProcessPayload([]byte("invalid-random-data"))
	if err == nil {
		t.Fatalf("expected error for malformed protobuf, got nil")
	}

	// Payload with empty asset ID
	emptyAssetPayload := &telemetryv1.TelemetryPayload{
		PhysicalAssetId: "",
		Readings: []*telemetryv1.SensorReading{
			{MetricName: "temp", Value: 50.0},
		},
	}
	data, _ := proto.Marshal(emptyAssetPayload)
	err = consumer.ProcessPayload(data)
	if err == nil {
		t.Fatalf("expected error for empty physical_asset_id, got nil")
	}
}

func TestTelemetryConsumer_ConsumeLoop(t *testing.T) {
	now := time.Now().UTC()
	payload := &telemetryv1.TelemetryPayload{
		PhysicalAssetId: "press-01",
		EdgeTimestamp:   timestamppb.New(now),
		Readings: []*telemetryv1.SensorReading{
			{MetricName: "pressure_bar", Value: 210.0, Quality: "GOOD"},
		},
	}
	data, _ := proto.Marshal(payload)

	reader := newMockReader([]kafka.Message{
		{Value: data},
	})


	fullCfg, _ := config.LoadConfig()
	writer := db.NewBatchWriter(nil, fullCfg.Ingestion, fullCfg.Database.TableName)
	consumer := NewTelemetryConsumer(reader, writer, nil, 50*time.Millisecond)


	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	writer.Start(ctx)
	defer writer.Stop()

	consumer.Start(ctx)

	// Wait until the single message has been consumed and parsed.
	waitForCondition(t, 2*time.Second, func() bool {
		msgs, _, _ := consumer.Stats()
		return msgs >= 1
	})

	consumer.Stop()

	msgs, readings, _ := consumer.Stats()
	if msgs != 1 || readings != 1 {
		t.Fatalf("expected msgs=1, readings=1, got msgs=%d, readings=%d", msgs, readings)
	}
}


type errorReader struct {
	hasReturnedError bool
	stopCh           chan struct{}
	once             sync.Once
	mu               sync.Mutex
}

func newErrorReader() *errorReader {
	return &errorReader{stopCh: make(chan struct{})}
}

func (e *errorReader) ReadMessage(ctx context.Context) (kafka.Message, error) {
	e.mu.Lock()
	if !e.hasReturnedError {
		e.hasReturnedError = true
		e.mu.Unlock()
		return kafka.Message{}, kafka.LeaderNotAvailable
	}
	e.mu.Unlock()

	select {
	case <-ctx.Done():
		return kafka.Message{}, ctx.Err()
	case <-e.stopCh:
		return kafka.Message{}, io.EOF
	}
}

func (e *errorReader) Close() error {
	e.once.Do(func() {
		close(e.stopCh)
	})
	return nil
}

func TestTelemetryConsumer_ReaderError(t *testing.T) {
	reader := newErrorReader()
	consumer := NewTelemetryConsumer(reader, nil, nil, 10*time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	consumer.Start(ctx)

	// Wait until the consumer has recorded at least one error.
	waitForCondition(t, 2*time.Second, func() bool {
		_, _, errs := consumer.Stats()
		return errs > 0
	})
	consumer.Stop()

	_, _, errCount := consumer.Stats()
	if errCount == 0 {
		t.Fatalf("expected at least 1 consumer error, got: %d", errCount)
	}
}


func TestTelemetryConsumer_NilTimestamp(t *testing.T) {
	consumer := NewTelemetryConsumer(nil, nil, nil, 0)
	payload := &telemetryv1.TelemetryPayload{
		PhysicalAssetId: "press-02",
		EdgeTimestamp:   nil, // Nil timestamp should use time.Now() fallback
		Readings: []*telemetryv1.SensorReading{
			{MetricName: "oil_level", Value: 88.0, Quality: ""},
		},
	}
	data, _ := proto.Marshal(payload)
	if err := consumer.ProcessPayload(data); err != nil {
		t.Fatalf("failed to process payload with nil timestamp: %v", err)
	}
}

func TestTelemetryConsumer_ContextCancel(t *testing.T) {
	reader := newMockReader([]kafka.Message{})
	consumer := NewTelemetryConsumer(reader, nil, nil, 50*time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	consumer.Start(ctx)

	// Cancel context immediately
	cancel()
	consumer.Stop()
}

func TestTelemetryConsumer_ConsumeLoop_MalformedPayload(t *testing.T) {
	reader := newMockReader([]kafka.Message{
		{Value: []byte("corrupt-bytes-not-protobuf")},
	})
	consumer := NewTelemetryConsumer(reader, nil, nil, 20*time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	consumer.Start(ctx)

	waitForCondition(t, 2*time.Second, func() bool {
		_, _, errs := consumer.Stats()
		return errs > 0
	})
	consumer.Stop()

	_, _, errs := consumer.Stats()
	if errs < 1 {
		t.Fatalf("expected at least 1 error count for malformed msg in loop, got: %d", errs)
	}
}

func TestTelemetryConsumer_ProcessRecordBatch(t *testing.T) {
	consumer := NewTelemetryConsumer(nil, nil, nil, 0)
	recordBatch := &telemetryv1.RecordBatch{
		BatchId:    "batch-test-101",
		EdgeNodeId: "node-101",
		Payloads: []*telemetryv1.TelemetryPayload{
			{
				PhysicalAssetId: "cnc-machine-01",
				EdgeTimestamp:   timestamppb.Now(),
				Readings: []*telemetryv1.SensorReading{
					{MetricName: "temp", Value: 42.0, Quality: "GOOD"},
					{MetricName: "vib", Value: 0.02, Quality: "GOOD"},
				},
			},
			{
				PhysicalAssetId: "cnc-machine-02",
				EdgeTimestamp:   timestamppb.Now(),
				Readings: []*telemetryv1.SensorReading{
					{MetricName: "speed", Value: 1500.0, Quality: "GOOD"},
				},
			},
		},
	}

	data, err := proto.Marshal(recordBatch)
	if err != nil {
		t.Fatalf("failed to marshal RecordBatch: %v", err)
	}

	if err := consumer.ProcessPayload(data); err != nil {
		t.Fatalf("ProcessPayload failed for RecordBatch: %v", err)
	}

	_, readings, errs := consumer.Stats()
	if readings != 3 || errs != 0 {
		t.Errorf("expected 3 readings parsed, got %d (errs: %d)", readings, errs)
	}
}


