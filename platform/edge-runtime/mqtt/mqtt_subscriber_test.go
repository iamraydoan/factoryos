package mqtt

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"google.golang.org/protobuf/proto"

	"github.com/iamraydoan/factoryos/platform/edge-runtime/buffer"
	"github.com/iamraydoan/factoryos/platform/edge-runtime/collector"
	telemetryv1 "github.com/iamraydoan/factoryos/platform/platform-sdk/go/gen/telemetry/v1"
)

type mockMessage struct {
	topic   string
	payload []byte
}

func (m *mockMessage) Duplicate() bool   { return false }
func (m *mockMessage) Qos() byte         { return 0 }
func (m *mockMessage) Retained() bool    { return false }
func (m *mockMessage) Topic() string     { return m.topic }
func (m *mockMessage) MessageID() uint16 { return 1 }
func (m *mockMessage) Payload() []byte   { return m.payload }
func (m *mockMessage) Ack()              {}

type mockMQTTToken struct {
	err error
}

func (t *mockMQTTToken) Wait() bool                     { return true }
func (t *mockMQTTToken) WaitTimeout(d time.Duration) bool { return true }
func (t *mockMQTTToken) Done() <-chan struct{}          { return nil }
func (t *mockMQTTToken) Error() error                   { return t.err }

type mockMQTTClient struct {
	mqtt.Client
	connected      bool
	subscribed     bool
	subscribeError error
}

func (m *mockMQTTClient) IsConnected() bool {
	return m.connected
}

func (m *mockMQTTClient) Connect() mqtt.Token {
	return &mockMQTTToken{err: nil}
}

func (m *mockMQTTClient) Disconnect(quiesce uint) {
	m.connected = false
}

func (m *mockMQTTClient) Subscribe(topic string, qos byte, callback mqtt.MessageHandler) mqtt.Token {
	m.subscribed = true
	return &mockMQTTToken{err: m.subscribeError}
}

func TestNewMQTTSubscriber_And_Disconnect(t *testing.T) {
	buf, err := buffer.NewSQLiteBuffer("file:mqtt_new_sub?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("failed to create buffer: %v", err)
	}
	defer buf.Close()

	telCollector := collector.NewTelemetryCollector(buf, nil)
	sub, err := NewMQTTSubscriber("tcp://localhost:1883", "test-client-id", telCollector)
	if err != nil {
		t.Fatalf("NewMQTTSubscriber failed: %v", err)
	}
	if sub == nil {
		t.Fatal("expected non-nil subscriber instance")
	}

	// Inject connected mock client to test Disconnect branch
	mockClient := &mockMQTTClient{connected: true}
	sub.client = mockClient

	// Test Disconnect when connected
	sub.Disconnect(10)
	if mockClient.connected {
		t.Error("expected client to be disconnected")
	}

	// Test Connect with mock client
	if err := sub.Connect(); err != nil {
		t.Errorf("expected Connect to succeed with mock client, got %v", err)
	}
}

func TestMQTTSubscriber_OnConnectCallback(t *testing.T) {
	buf, err := buffer.NewSQLiteBuffer("file:mqtt_on_connect?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("failed to create buffer: %v", err)
	}
	defer buf.Close()

	telCollector := collector.NewTelemetryCollector(buf, nil)
	sub := &MQTTSubscriber{collector: telCollector}

	// 1. Test successful onConnect
	mockClientSuccess := &mockMQTTClient{connected: true, subscribeError: nil}
	sub.onConnect(mockClientSuccess)
	if !mockClientSuccess.subscribed {
		t.Error("expected mockClient to be subscribed")
	}

	// 2. Test onConnect with subscribe error
	mockClientError := &mockMQTTClient{connected: true, subscribeError: errors.New("subscription permission denied")}
	sub.onConnect(mockClientError)
}

func TestMQTTSubscriber_HandleMessage_JSON(t *testing.T) {
	ctx := context.Background()
	buf, err := buffer.NewSQLiteBuffer("file:mqtt_test_json?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("failed to create buffer: %v", err)
	}
	defer buf.Close()

	telCollector := collector.NewTelemetryCollector(buf, nil)
	sub := &MQTTSubscriber{collector: telCollector}

	jsonPayload := TelemetryJSONMessage{
		PhysicalAssetID: "asset-json-01",
		Readings: []SensorReadingJSON{
			{MetricName: "temperature", Value: 55.4, Quality: "GOOD"},
		},
	}
	bytes, _ := json.Marshal(jsonPayload)

	msg := &mockMessage{topic: "factoryos/telemetry/asset-json-01/readings", payload: bytes}
	sub.handleMessage(nil, msg)

	count, err := buf.GetPendingCount(ctx)
	if err != nil {
		t.Fatalf("GetPendingCount failed: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 record in buffer after JSON MQTT message, got %d", count)
	}
}

func TestMQTTSubscriber_HandleMessage_Protobuf(t *testing.T) {
	ctx := context.Background()
	buf, err := buffer.NewSQLiteBuffer("file:mqtt_test_pb?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("failed to create buffer: %v", err)
	}
	defer buf.Close()

	telCollector := collector.NewTelemetryCollector(buf, nil)
	sub := &MQTTSubscriber{collector: telCollector}

	pbPayload := &telemetryv1.TelemetryPayload{
		PhysicalAssetId: "asset-pb-01",
		Readings: []*telemetryv1.SensorReading{
			{MetricName: "vibration", Value: 0.02, Quality: "GOOD"},
		},
	}
	bytes, _ := proto.Marshal(pbPayload)

	msg := &mockMessage{topic: "factoryos/telemetry/asset-pb-01/readings", payload: bytes}
	sub.handleMessage(nil, msg)

	count, err := buf.GetPendingCount(ctx)
	if err != nil {
		t.Fatalf("GetPendingCount failed: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 record in buffer after Protobuf MQTT message, got %d", count)
	}
}

func TestMQTTSubscriber_HandleMessage_RecordBatchError(t *testing.T) {
	buf, err := buffer.NewSQLiteBuffer("file:mqtt_test_db_err?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("failed to create buffer: %v", err)
	}
	buf.Close() // Close DB immediately to trigger RecordBatch error

	telCollector := collector.NewTelemetryCollector(buf, nil)
	sub := &MQTTSubscriber{collector: telCollector}

	jsonPayload := TelemetryJSONMessage{
		PhysicalAssetID: "asset-err-01",
		Readings: []SensorReadingJSON{
			{MetricName: "temperature", Value: 55.4, Quality: "GOOD"},
		},
	}
	bytes, _ := json.Marshal(jsonPayload)

	msg := &mockMessage{topic: "factoryos/telemetry/asset-err-01/readings", payload: bytes}
	sub.handleMessage(nil, msg) // Handles error gracefully without panic
}

func TestMQTTSubscriber_HandleMessage_Unparseable(t *testing.T) {
	buf, err := buffer.NewSQLiteBuffer("file:mqtt_test_unparseable?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("failed to create buffer: %v", err)
	}
	defer buf.Close()

	telCollector := collector.NewTelemetryCollector(buf, nil)
	sub := &MQTTSubscriber{collector: telCollector}

	// Send unparseable text payload
	msg := &mockMessage{topic: "factoryos/telemetry/bad/readings", payload: []byte("INVALID_DATA")}
	sub.handleMessage(nil, msg)
}
