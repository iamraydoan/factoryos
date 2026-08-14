package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type Config struct {
	MQTTBroker        string              `json:"mqtt_broker"`
	PublishIntervalMs int                 `json:"publish_interval_ms"`
	Machines          []*SimulatedMachine `json:"machines"`
}

type SensorReadingJSON struct {
	MetricName string  `json:"metric_name"`
	Value      float64 `json:"value"`
	Quality    string  `json:"quality"`
}

type TelemetryJSONMessage struct {
	PhysicalAssetID string              `json:"physical_asset_id"`
	Readings        []SensorReadingJSON `json:"readings"`
}

type SimulatedMachine struct {
	AssetID      string  `json:"asset_id"`
	BaseTemp     float64 `json:"base_temp"`
	BaseVib      float64 `json:"base_vib"`
	PartCount    float64 `json:"part_count"`
	MachineState float64 `json:"machine_state"` // 1=RUNNING, 2=IDLE, 3=FAULTED
}

func loadConfig(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func defaultConfig() *Config {
	return &Config{
		MQTTBroker:        "tcp://localhost:1883",
		PublishIntervalMs: 2000,
		Machines: []*SimulatedMachine{
			{AssetID: "cnc-machine-01", BaseTemp: 45.0, BaseVib: 0.02, PartCount: 100, MachineState: 1.0},
			{AssetID: "cnc-machine-02", BaseTemp: 52.0, BaseVib: 0.03, PartCount: 250, MachineState: 1.0},
			{AssetID: "robotic-arm-01", BaseTemp: 38.0, BaseVib: 0.01, PartCount: 500, MachineState: 1.0},
		},
	}
}

func main() {
	log.Println("==================================================")
	log.Println(" FactoryOS - Configurable PLC Telemetry Simulator")
	log.Println("==================================================")

	configPathFlag := flag.String("config", "config.json", "Path to simulator JSON config file")
	flag.Parse()

	cfg, err := loadConfig(*configPathFlag)
	if err != nil {
		log.Printf("Notice: Could not load config file '%s' (%v). Using default configuration.", *configPathFlag, err)
		cfg = defaultConfig()
	} else {
		log.Printf("Successfully loaded configuration from '%s'", *configPathFlag)
	}

	// Environment variable overrides
	if envBroker := os.Getenv("MQTT_BROKER_URL"); envBroker != "" {
		cfg.MQTTBroker = envBroker
	}

	if cfg.PublishIntervalMs <= 0 {
		cfg.PublishIntervalMs = 2000
	}

	opts := mqtt.NewClientOptions()
	opts.AddBroker(cfg.MQTTBroker)
	opts.SetClientID("mock-plc-simulator")
	opts.SetAutoReconnect(true)

	log.Printf("Connecting to MQTT Broker at %s...", cfg.MQTTBroker)
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Printf("Warning: Failed to connect to MQTT broker (%v). Will keep retrying...", token.Error())
	} else {
		log.Println("Successfully connected to MQTT Broker!")
	}

	interval := time.Duration(cfg.PublishIntervalMs) * time.Millisecond
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	log.Printf("Simulating %d machines every %v. Press Ctrl+C to stop.", len(cfg.Machines), interval)

	for {
		select {
		case <-sigChan:
			log.Println("Stopping Mock PLC Telemetry Simulator...")
			client.Disconnect(250)
			return
		case <-ticker.C:
			for _, m := range cfg.Machines {
				// Simulate random fluctuations
				m.BaseTemp += (rand.Float64()*1.0 - 0.5)
				if m.BaseTemp < 35.0 {
					m.BaseTemp = 35.0
				} else if m.BaseTemp > 80.0 {
					m.BaseTemp = 80.0
				}

				m.BaseVib += (rand.Float64()*0.005 - 0.0025)
				if m.BaseVib < 0.005 {
					m.BaseVib = 0.005
				}

				m.PartCount += 1.0

				payload := TelemetryJSONMessage{
					PhysicalAssetID: m.AssetID,
					Readings: []SensorReadingJSON{
						{MetricName: "spindle_temperature_celsius", Value: mathRound(m.BaseTemp, 2), Quality: "GOOD"},
						{MetricName: "spindle_vibration_mm_s", Value: mathRound(m.BaseVib, 4), Quality: "GOOD"},
						{MetricName: "part_produced_count", Value: m.PartCount, Quality: "GOOD"},
						{MetricName: "machine_state", Value: m.MachineState, Quality: "GOOD"},
					},
				}

				payloadBytes, err := json.Marshal(payload)
				if err != nil {
					log.Printf("Error marshaling payload: %v", err)
					continue
				}

				topic := fmt.Sprintf("factoryos/telemetry/%s/readings", m.AssetID)
				if client.IsConnected() {
					token := client.Publish(topic, 1, false, payloadBytes)
					token.Wait()
					log.Printf("[Simulator] Published -> Asset: %s | Temp: %.2f°C | Vibration: %.4f | Parts: %.0f",
						m.AssetID, m.BaseTemp, m.BaseVib, m.PartCount)
				} else {
					log.Printf("[Simulator (Offline)] Generated payload for %s | Temp: %.2f°C", m.AssetID, m.BaseTemp)
				}
			}
		}
	}
}

func mathRound(val float64, precision int) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
}
