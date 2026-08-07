module github.com/iamraydoan/factoryos/platform/edge-runtime

go 1.22

require (
	github.com/eclipse/paho.mqtt.golang v1.4.3
	github.com/iamraydoan/factoryos/platform/platform-sdk v0.0.0
	github.com/mattn/go-sqlite3 v1.14.22
	google.golang.org/protobuf v1.34.1
)

replace github.com/iamraydoan/factoryos/platform/platform-sdk => ../platform-sdk
