// Package main is the entrypoint for the FactoryOS Resource Service.
// It wires configuration, database, migrations, and gRPC server together,
// and handles graceful shutdown on SIGINT/SIGTERM.
package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"

	"github.com/iamraydoan/factoryos/services/resource-service/config"
	"github.com/iamraydoan/factoryos/services/resource-service/db"
	"github.com/iamraydoan/factoryos/services/resource-service/server"
	resourcev1 "github.com/iamraydoan/factoryos/platform/platform-sdk/go/gen/resource/v1"
)

func main() {
	log.Println("================================================================")
	log.Println(" FactoryOS Resource Service (ISA-95 Equipment Hierarchy)")
	log.Println("================================================================")

	// 1. Load configuration from environment variables / .env file.
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("[Config][FATAL] Failed to load configuration: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 2. Connect to PostgreSQL.
	database, err := db.NewDB(ctx, cfg.Database)
	if err != nil {
		log.Fatalf("[PostgreSQL][FATAL] Failed to connect: %v", err)
	}
	defer database.Close()

	// 3. Run database migrations (goose).
	if err := database.RunMigrations(ctx); err != nil {
		log.Fatalf("[PostgreSQL][FATAL] Failed to run migrations: %v", err)
	}

	// 4. Create repository and register gRPC server.
	repo := db.NewPostgresEquipmentRepository(database.Pool())
	grpcServer := grpc.NewServer()
	resourcev1.RegisterEquipmentServiceServer(grpcServer, server.NewEquipmentService(repo))

	// 5. Start HTTP sidecar for health checks.
	httpServer := startHTTPServer(cfg.Server.MetricsPort)
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("[HTTP][ERROR] %v", err)
		}
	}()

	// 6. Listen on gRPC port.
	lis, err := net.Listen("tcp", cfg.Server.Port)
	if err != nil {
		log.Fatalf("[gRPC][FATAL] Failed to listen on %s: %v", cfg.Server.Port, err)
	}
	go func() {
		log.Printf("[gRPC] Server listening on %s", cfg.Server.Port)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("[gRPC][FATAL] Server error: %v", err)
		}
	}()

	// 7. Wait for shutdown signal (SIGINT or SIGTERM).
	shutdownSig := make(chan os.Signal, 1)
	signal.Notify(shutdownSig, syscall.SIGINT, syscall.SIGTERM)
	<-shutdownSig

	// 8. Graceful shutdown: finish in-flight requests, then exit.
	log.Println("[Shutdown] Signal received. Stopping gracefully...")
	cancel()

	// Stop gRPC server (finish in-flight RPCs)
	grpcServer.GracefulStop()

	// Stop HTTP server
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer shutdownCancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("[HTTP][WARN] Shutdown error: %v", err)
	}

	log.Println("[Shutdown] Resource Service terminated cleanly.")
}

// startHTTPServer creates a lightweight HTTP server for health checks.
func startHTTPServer(port string) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"HEALTHY","service":"resource-service"}`))
	})

	server := &http.Server{
		Addr:    "127.0.0.1" + port,
		Handler: mux,
	}

	log.Printf("[HTTP] Health server listening on 127.0.0.1%s /healthz", port)
	return server
}
