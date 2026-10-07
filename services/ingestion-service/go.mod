module github.com/iamraydoan/factoryos/services/ingestion-service

go 1.25.0

require (
	github.com/caarlos0/env/v11 v11.3.1
	github.com/go-playground/validator/v10 v10.25.0
	github.com/iamraydoan/factoryos/platform/platform-sdk v0.0.0
	github.com/joho/godotenv v1.5.1
	github.com/segmentio/kafka-go v0.4.47
	google.golang.org/grpc v1.84.0
	google.golang.org/protobuf v1.36.12
)

require (
	github.com/gabriel-vasile/mimetype v1.4.8 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/klauspost/compress v1.15.9 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/pierrec/lz4/v4 v4.1.15 // indirect
	golang.org/x/crypto v0.54.0 // indirect
	golang.org/x/net v0.57.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
	golang.org/x/text v0.40.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260706201446-f0a921348800 // indirect
)

replace github.com/iamraydoan/factoryos/platform/platform-sdk => ../../platform/platform-sdk
