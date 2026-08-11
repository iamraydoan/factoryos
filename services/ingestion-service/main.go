package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"

	telemetryv1 "github.com/iamraydoan/factoryos/platform/platform-sdk/go/gen/telemetry/v1"
	"github.com/iamraydoan/factoryos/services/ingestion-service/config"
	"github.com/iamraydoan/factoryos/services/ingestion-service/kafka"
	"github.com/iamraydoan/factoryos/services/ingestion-service/server"
)

func main() {
	log.Println("================================================================")
	log.Println(" FactoryOS Telemetry Ingestion Gateway Service (gRPC)")
	log.Println("================================================================")

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("[Config][FATAL] Failed to load configuration: %v", err)
	}

	log.Printf("[Config] gRPC Port: %s | Metrics Port: %s", cfg.Server.GRPCPort, cfg.Server.MetricsPort)
	log.Printf("[Config] Kafka Brokers: %v | Topic: %s", cfg.Kafka.Brokers, cfg.Kafka.Topic)

	// 1. Initialize Kafka Producer
	kafkaProducer := kafka.NewKafkaProducer(cfg.Kafka)
	defer kafkaProducer.Close()

	// 2. Initialize Ingestion Server Handler
	ingestionServer := server.NewIngestionServer(kafkaProducer)

	// 3. Start gRPC Server
	grpcListener, err := net.Listen("tcp", cfg.Server.GRPCPort)
	if err != nil {
		log.Fatalf("[gRPC][FATAL] Failed to bind gRPC port %s: %v", cfg.Server.GRPCPort, err)
	}

	grpcServer := grpc.NewServer()
	telemetryv1.RegisterTelemetryIngestionServiceServer(grpcServer, ingestionServer)

	go func() {
		log.Printf("[gRPC] Telemetry Ingestion Service listening on %s", cfg.Server.GRPCPort)
		if err := grpcServer.Serve(grpcListener); err != nil && err != grpc.ErrServerStopped {
			log.Fatalf("[gRPC][FATAL] gRPC server error: %v", err)
		}
	}()

	// 4. Start HTTP Health & Metrics Server
	httpServer := startHTTPServer(cfg.Server.MetricsPort, ingestionServer)

	// 5. Graceful Shutdown Listener
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Println("\n[Shutdown] Termination signal received. Stopping Ingestion Service gracefully...")

	// Graceful gRPC stop
	grpcServer.GracefulStop()

	// Graceful HTTP stop
	shutdownTimeout := cfg.Server.ShutdownTimeout
	if shutdownTimeout <= 0 {
		shutdownTimeout = 5 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	_ = httpServer.Shutdown(ctx)

	log.Println("[Shutdown] FactoryOS Ingestion Service terminated cleanly.")
}

func startHTTPServer(port string, srv *server.IngestionServer) *http.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"HEALTHY","service":"ingestion-service"}`))
	})

	mux.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		batches, records, errs := srv.Stats()
		stats := map[string]any{
			"batches_received": batches,
			"records_received": records,
			"errors_count":     errs,
			"timestamp":        time.Now().UTC().Format(time.RFC3339),
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(stats)
	})

	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		batches, records, errs := srv.Stats()
		resp := fmt.Sprintf("# HELP telemetry_ingestion_batches_total Total batches received\n"+
			"# TYPE telemetry_ingestion_batches_total counter\n"+
			"telemetry_ingestion_batches_total %d\n\n"+
			"# HELP telemetry_ingestion_records_total Total sensor readings received\n"+
			"# TYPE telemetry_ingestion_records_total counter\n"+
			"telemetry_ingestion_records_total %d\n\n"+
			"# HELP telemetry_ingestion_errors_total Total ingestion errors\n"+
			"# TYPE telemetry_ingestion_errors_total counter\n"+
			"telemetry_ingestion_errors_total %d\n",
			batches, records, errs)

		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		_, _ = w.Write([]byte(resp))
	})

	server := &http.Server{
		Addr:    "127.0.0.1" + port,
		Handler: mux,
	}

	go func() {
		log.Printf("[HTTP] Metrics & Health Server listening on 127.0.0.1%s (/healthz, /stats, /metrics)", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("[HTTP][WARN] HTTP server closed: %v", err)
		}
	}()

	return server
}
