package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/iamraydoan/factoryos/services/analytics-engine/config"
	"github.com/iamraydoan/factoryos/services/analytics-engine/consumer"
	"github.com/iamraydoan/factoryos/services/analytics-engine/db"
	"github.com/iamraydoan/factoryos/services/analytics-engine/processor"
)

func main() {
	log.Println("================================================================")
	log.Println(" FactoryOS Real-Time Analytics Engine & Telemetry Ingestion")
	log.Println("================================================================")

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("[Config][FATAL] Failed to load configuration: %v", err)
	}


	log.Printf("[Config] Kafka Brokers: %v | Topic: %s | Group: %s",
		cfg.Kafka.Brokers, cfg.Kafka.Topic, cfg.Kafka.GroupID)
	log.Printf("[Config] Database Table: %s | MaxConns: %d | Retention: %d days",
		cfg.Database.TableName, cfg.Database.MaxConns, cfg.Database.RawRetentionDays)
	log.Printf("[Config] Batch Size: %d | Flush Interval: %v | Capacity Multiplier: %d",
		cfg.Ingestion.BatchSize, cfg.Ingestion.FlushInterval, cfg.Ingestion.ChannelCapacityMultiplier)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 1. Initialize TimescaleDB Connection Pool
	database, err := db.NewDB(ctx, cfg.Database)
	if err != nil {
		log.Printf("[TimescaleDB][WARN] Database connection failed (will retry in background): %v", err)
	} else {
		defer database.Close()
		// Run initial migrations
		if err := database.RunMigrations(ctx); err != nil {
			log.Printf("[TimescaleDB][WARN] Failed applying migrations: %v", err)
		}
	}

	// 2. Initialize Batch Writer
	var batchWriter *db.BatchWriter
	if database != nil {
		batchWriter = db.NewBatchWriter(database.AsBatchInserter(), cfg.Ingestion, cfg.Database.TableName)
	} else {
		batchWriter = db.NewBatchWriter(nil, cfg.Ingestion, cfg.Database.TableName)
	}
	batchWriter.Start(ctx)

	// 3. Initialize Real-Time Dynamic Alert Evaluator
	evaluator := processor.NewAlertEvaluator(nil, nil)

	// 4. Initialize Real-Time OEE Streaming Aggregator
	oeeCfg := processor.OEEConfig{
		WindowDuration:        cfg.OEE.WindowDuration,
		SnapshotInterval:      cfg.OEE.SnapshotInterval,
		DefaultIdealCycleTime: cfg.OEE.DefaultIdealCycleTime,
	}
	oeeAggregator := processor.NewOEEAggregator(oeeCfg, nil)
	oeeAggregator.Start(ctx)

	// 5. Initialize Kafka Consumer
	kafkaReader := consumer.NewKafkaReader(cfg.Kafka)
	telemetryConsumer := consumer.NewTelemetryConsumer(kafkaReader, batchWriter, evaluator, oeeAggregator, cfg.Kafka.RetryBackoff)
	telemetryConsumer.Start(ctx)

	// 6. Start HTTP Metrics & Health Endpoint
	httpServer := startHTTPServer(cfg.Server.MetricsPort, telemetryConsumer, batchWriter, oeeAggregator)

	// 7. Graceful Shutdown Listener
	shutdownSig := make(chan os.Signal, 1)
	signal.Notify(shutdownSig, syscall.SIGINT, syscall.SIGTERM)

	<-shutdownSig
	log.Println("\n[Shutdown] Termination signal received. Stopping services gracefully...")

	// Cancel context for child routines
	cancel()

	// Stop Consumer first (stop accepting new Kafka messages)
	telemetryConsumer.Stop()

	// Stop OEE Aggregator
	oeeAggregator.Stop()

	// Flush and stop Batch Writer
	batchWriter.Stop()

	// Shutdown HTTP Server
	shutdownTimeout := cfg.Server.ShutdownTimeout
	if shutdownTimeout <= 0 {
		shutdownTimeout = 5 * time.Second
	}
	shutdownCtx, httpCancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer httpCancel()
	_ = httpServer.Shutdown(shutdownCtx)

	log.Println("[Shutdown] FactoryOS Analytics Engine terminated cleanly.")
}

// startHTTPServer starts the metrics and health HTTP server on the loopback interface.
// It detects immediate startup failures (port conflict, permission denied) within a
// short window and calls log.Fatalf so the process exits cleanly instead of running zombie.
func startHTTPServer(port string, cons *consumer.TelemetryConsumer, writer *db.BatchWriter, oeeAgg *processor.OEEAggregator) *http.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"HEALTHY","service":"analytics-engine"}`))
	})

	mux.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		msgs, readings, errs := cons.Stats()
		inserted, batches := writer.Stats()

		stats := map[string]any{
			"kafka_messages_consumed": msgs,
			"sensor_readings_parsed":  readings,
			"consumer_errors":         errs,
			"timescaledb_inserted":    inserted,
			"timescaledb_batches":     batches,
			"timestamp":               time.Now().UTC().Format(time.RFC3339),
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(stats)
	})

	// TODO: Add authentication/authorization (API key or mTLS) before production deployment.
	// OEE data exposes per-asset production efficiency metrics which may be considered sensitive.
	mux.HandleFunc("/oee", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		assetID := r.URL.Query().Get("asset_id")
		if assetID != "" {
			snap := oeeAgg.GetSnapshotForAsset(assetID)
			if snap == nil {
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte(`{"error":"no OEE data for asset"}`))
				return
			}
			_ = json.NewEncoder(w).Encode(snap)
			return
		}

		snapshots := oeeAgg.GetSnapshots()
		_ = json.NewEncoder(w).Encode(snapshots)
	})

	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		msgs, readings, errs := cons.Stats()
		inserted, batches := writer.Stats()

		// Prometheus exposition format
		resp := fmt.Sprintf("# HELP telemetry_consumed_total Total messages read from Kafka\n"+
			"# TYPE telemetry_consumed_total counter\n"+
			"telemetry_consumed_total %d\n\n"+
			"# HELP telemetry_readings_parsed_total Total sensor readings parsed\n"+
			"# TYPE telemetry_readings_parsed_total counter\n"+
			"telemetry_readings_parsed_total %d\n\n"+
			"# HELP telemetry_consumer_errors_total Total parsing or consumer errors\n"+
			"# TYPE telemetry_consumer_errors_total counter\n"+
			"telemetry_consumer_errors_total %d\n\n"+
			"# HELP timescaledb_records_inserted_total Total records written to TimescaleDB\n"+
			"# TYPE timescaledb_records_inserted_total counter\n"+
			"timescaledb_records_inserted_total %d\n\n"+
			"# HELP timescaledb_batches_flushed_total Total CopyFrom batches flushed to TimescaleDB\n"+
			"# TYPE timescaledb_batches_flushed_total counter\n"+
			"timescaledb_batches_flushed_total %d\n",
			msgs, readings, errs, inserted, batches)

		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		_, _ = w.Write([]byte(resp))
	})

	// Bind to loopback only: prevents exposure of operational metrics to external networks.
	// Use a reverse proxy (nginx, Envoy) or a dedicated scrape network for Prometheus.
	server := &http.Server{
		Addr:    "127.0.0.1" + port,
		Handler: mux,
	}

	// Detect immediate startup failure within a short window.
	// This prevents the service from running as a zombie when the port is unavailable.
	errCh := make(chan error, 1)
	go func() {
		log.Printf("[HTTP] Metrics & Health Server listening on 127.0.0.1%s (/healthz, /stats, /metrics)", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case err, open := <-errCh:
		if open && err != nil {
			log.Fatalf("[HTTP][FATAL] Failed to start HTTP server on 127.0.0.1%s: %v", port, err)
		}
	case <-time.After(200 * time.Millisecond):
		// Server started successfully within the detection window.
	}

	return server
}
