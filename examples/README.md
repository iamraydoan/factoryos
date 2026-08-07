# Reference Examples & Simulators (`examples`)

Contains reference implementations, mock PLC telemetry simulators (OPC-UA/MQTT generators), example workflow scripts, and SDK integration starter templates.

---

## 🛠️ Mock PLC Telemetry Simulator (`examples/mock-plc-simulator`)

The `mock-plc-simulator` tool simulates real-world industrial factory equipment (CNC machines, robotic arms, stamping presses) publishing real-time telemetry sensor data over MQTT.

### ⚙️ Configuration (`config.json`)

The simulator reads configuration settings from [config.json](mock-plc-simulator/config.json):

```json
{
  "mqtt_broker": "tcp://localhost:1883",
  "publish_interval_ms": 2000,
  "machines": [
    {
      "asset_id": "cnc-machine-01",
      "base_temp": 45.0,
      "base_vib": 0.02,
      "part_count": 100,
      "machine_state": 1.0
    },
    {
      "asset_id": "cnc-machine-02",
      "base_temp": 52.0,
      "base_vib": 0.03,
      "part_count": 250,
      "machine_state": 1.0
    },
    {
      "asset_id": "robotic-arm-01",
      "base_temp": 38.0,
      "base_vib": 0.01,
      "part_count": 500,
      "machine_state": 1.0
    }
  ]
}
```

* `mqtt_broker`: Address of the local or remote MQTT Broker.
* `publish_interval_ms`: Frequency of telemetry payload generation (milliseconds).
* `machines`: Array of simulated equipment, each with initial temperature, vibration, part production count, and operational state (`1.0` = RUNNING, `2.0` = IDLE, `3.0` = FAULTED).

---

### 🚀 Running the Simulator

1. Ensure the MQTT Broker is running (e.g. via `docker compose up -d mosquitto`).
2. Run the simulator from the repository root:

```bash
cd examples/mock-plc-simulator
go run main.go
```

To run with a custom configuration file:

```bash
go run main.go -config /path/to/custom_config.json
```
