package mqtt

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"google.golang.org/protobuf/proto"

	"github.com/iamraydoan/factoryos/platform/edge-runtime/collector"
	telemetryv1 "github.com/iamraydoan/factoryos/platform/platform-sdk/go/gen/telemetry/v1"
)

// MQTTSubscriber manages connection to an MQTT broker and processes telemetry payloads.
type MQTTSubscriber struct {
	client    mqtt.Client
	collector *collector.TelemetryCollector
}

// TelemetryJSONMessage represents a JSON fallback payload sent by lightweight MQTT sensors/simulators.
type TelemetryJSONMessage struct {
	PhysicalAssetID string              `json:"physical_asset_id"`
	Readings        []SensorReadingJSON `json:"readings"`
}

type SensorReadingJSON struct {
	MetricName string  `json:"metric_name"`
	Value      float64 `json:"value"`
	Quality    string  `json:"quality"`
}

// NewMQTTSubscriber initializes a new MQTT subscriber client.
func NewMQTTSubscriber(brokerURL string, clientID string, telCollector *collector.TelemetryCollector) (*MQTTSubscriber, error) {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(brokerURL)
	opts.SetClientID(clientID)
	opts.SetAutoReconnect(true)
	opts.SetConnectRetry(true)
	opts.SetConnectTimeout(1 * time.Second) // Fast timeout for offline/test environments
	opts.SetConnectRetryInterval(2 * time.Second)

	sub := &MQTTSubscriber{collector: telCollector}
	opts.SetOnConnectHandler(sub.onConnect)

	client := mqtt.NewClient(opts)
	sub.client = client
	return sub, nil
}

func (s *MQTTSubscriber) onConnect(c mqtt.Client) {
	log.Println("[MQTT Subscriber] Connected to MQTT Broker")
	topic := "factoryos/telemetry/+/readings"
	if token := c.Subscribe(topic, 1, s.handleMessage); token.Wait() && token.Error() != nil {
		log.Printf("[MQTT Subscriber] Error subscribing to topic %s: %v", topic, token.Error())
	} else {
		log.Printf("[MQTT Subscriber] Subscribed to topic: %s", topic)
	}
}

// Connect initiates connection to the MQTT broker asynchronously.
func (s *MQTTSubscriber) Connect() error {
	if token := s.client.Connect(); token.WaitTimeout(1*time.Second) && token.Error() != nil {
		return fmt.Errorf("failed to connect to MQTT broker: %w", token.Error())
	}
	return nil
}

func (s *MQTTSubscriber) handleMessage(client mqtt.Client, msg mqtt.Message) {
	payloadBytes := msg.Payload()
	ctx := context.Background()

	// 1. Try unmarshaling as Protobuf
	var pbPayload telemetryv1.TelemetryPayload
	if err := proto.Unmarshal(payloadBytes, &pbPayload); err == nil && pbPayload.PhysicalAssetId != "" && len(pbPayload.Readings) > 0 {
		readingsMap := make(map[string]float64)
		for _, r := range pbPayload.Readings {
			readingsMap[r.MetricName] = r.Value
		}
		if err := s.collector.RecordBatch(ctx, pbPayload.PhysicalAssetId, readingsMap); err != nil {
			log.Printf("[MQTT Subscriber] Error recording protobuf batch: %v", err)
		} else {
			log.Printf("[MQTT Subscriber] Processed Protobuf payload from asset %s (%d metrics)", pbPayload.PhysicalAssetId, len(readingsMap))
		}
		return
	}

	// 2. Try unmarshaling as JSON
	var jsonMsg TelemetryJSONMessage
	if err := json.Unmarshal(payloadBytes, &jsonMsg); err == nil && jsonMsg.PhysicalAssetID != "" {
		readingsMap := make(map[string]float64)
		for _, r := range jsonMsg.Readings {
			readingsMap[r.MetricName] = r.Value
		}
		if err := s.collector.RecordBatch(ctx, jsonMsg.PhysicalAssetID, readingsMap); err != nil {
			log.Printf("[MQTT Subscriber] Error recording JSON batch: %v", err)
		} else {
			log.Printf("[MQTT Subscriber] Processed JSON payload from asset %s (%d metrics)", jsonMsg.PhysicalAssetID, len(readingsMap))
		}
		return
	}

	log.Printf("[MQTT Subscriber] Received unparseable payload on topic %s: %s", msg.Topic(), string(payloadBytes))
}

// Disconnect cleanly closes the MQTT broker connection.
func (s *MQTTSubscriber) Disconnect(quiesce uint) {
	if s.client != nil && s.client.IsConnected() {
		s.client.Disconnect(quiesce)
	}
}
