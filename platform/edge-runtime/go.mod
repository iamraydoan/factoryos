module github.com/iamraydoan/factoryos/platform/edge-runtime

go 1.25.0

require (
	github.com/eclipse/paho.mqtt.golang v1.4.3
	github.com/iamraydoan/factoryos/platform/platform-sdk v0.0.0
	github.com/mattn/go-sqlite3 v1.14.22
	google.golang.org/grpc v1.84.0
	google.golang.org/protobuf v1.36.12
)

require (
	github.com/gorilla/websocket v1.5.0 // indirect
	golang.org/x/net v0.57.0 // indirect
	golang.org/x/sync v0.22.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
	golang.org/x/text v0.40.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260706201446-f0a921348800 // indirect
)

replace github.com/iamraydoan/factoryos/platform/platform-sdk => ../platform-sdk
