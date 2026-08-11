module github.com/iamraydoan/factoryos/services/ingestion-service

go 1.22

require (
	github.com/caarlos0/env/v11 v11.3.1
	github.com/go-playground/validator/v10 v10.25.0
	github.com/iamraydoan/factoryos/platform/platform-sdk v0.0.0
	github.com/joho/godotenv v1.5.1
	github.com/segmentio/kafka-go v0.4.47
	google.golang.org/grpc v1.64.0
	google.golang.org/protobuf v1.34.1
)

replace github.com/iamraydoan/factoryos/platform/platform-sdk => ../../platform/platform-sdk
