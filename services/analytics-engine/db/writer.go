package db

import (
	"context"
	"crypto/sha1"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/iamraydoan/factoryos/services/analytics-engine/config"
)

// ErrBufferFull is returned by Enqueue when the ingestion channel is at capacity.
var ErrBufferFull = errors.New("batchwriter: ingestion channel is full, record dropped")

// TelemetryRecord represents a single flattened sensor reading to be inserted into TimescaleDB.
type TelemetryRecord struct {
	Time            time.Time
	PhysicalAssetID string
	MetricName      string
	Value           float64
	Quality         string
}

// BatchInserter defines the interface for executing bulk CopyFrom operations.
type BatchInserter interface {
	CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error)
}

// BatchWriter buffers incoming TelemetryRecords and periodically flushes them to TimescaleDB via pgx.CopyFrom.
type BatchWriter struct {
	inserter      BatchInserter
	tableName     string
	batchSize     int
	flushInterval time.Duration

	mu      sync.Mutex
	buffer  []TelemetryRecord
	records chan TelemetryRecord
	done    chan struct{}
	wg      sync.WaitGroup

	// Metrics counters
	totalInserted int64
	totalBatches  int64
	totalDropped  int64
}

// NewBatchWriter creates a new BatchWriter instance from validated IngestionConfig and target table name.
func NewBatchWriter(inserter BatchInserter, cfg config.IngestionConfig, tableName string) *BatchWriter {
	return &BatchWriter{
		inserter:      inserter,
		tableName:     tableName,
		batchSize:     cfg.BatchSize,
		flushInterval: cfg.FlushInterval,
		buffer:        make([]TelemetryRecord, 0, cfg.BatchSize),
		records:       make(chan TelemetryRecord, cfg.BatchSize*cfg.ChannelCapacityMultiplier),
		done:          make(chan struct{}),
	}
}

// Start launches the background worker that processes incoming records and periodic timer flushes.
func (w *BatchWriter) Start(ctx context.Context) {
	w.wg.Add(1)
	go w.workerLoop(ctx)
	log.Printf("[BatchWriter] Started worker loop (Table: %s, BatchSize: %d, FlushInterval: %v)",
		w.tableName, w.batchSize, w.flushInterval)
}

// Enqueue attempts a non-blocking send of a TelemetryRecord into the ingestion channel.
// If the channel is at capacity, the record is dropped and ErrBufferFull is returned.
// Callers should log or count dropped records at their discretion.
func (w *BatchWriter) Enqueue(record TelemetryRecord) error {
	select {
	case w.records <- record:
		return nil
	default:
		w.mu.Lock()
		w.totalDropped++
		dropped := w.totalDropped
		w.mu.Unlock()
		log.Printf("[BatchWriter][WARN] Channel buffer full — record DROPPED (metric=%s, asset=%s, total_dropped=%d)",
			record.MetricName, record.PhysicalAssetID, dropped)
		return ErrBufferFull
	}
}

// workerLoop continuously processes incoming records and periodic timer ticks.
func (w *BatchWriter) workerLoop(ctx context.Context) {
	defer w.wg.Done()
	ticker := time.NewTicker(w.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-w.done:
			w.drainAndFlush(context.Background())
			return

		case <-ctx.Done():
			w.drainAndFlush(context.Background())
			return

		case record, ok := <-w.records:
			if !ok {
				w.drainAndFlush(context.Background())
				return
			}
			w.mu.Lock()
			w.buffer = append(w.buffer, record)
			shouldFlush := len(w.buffer) >= w.batchSize
			w.mu.Unlock()

			if shouldFlush {
				if err := w.Flush(ctx); err != nil {
					log.Printf("[BatchWriter][ERROR] Error flushing batch on size threshold: %v", err)
				}
			}

		case <-ticker.C:
			w.mu.Lock()
			hasItems := len(w.buffer) > 0
			w.mu.Unlock()

			if hasItems {
				if err := w.Flush(ctx); err != nil {
					log.Printf("[BatchWriter][ERROR] Error flushing batch on timer ticker: %v", err)
				}
			}
		}
	}
}

// Flush writes the current in-memory buffer to TimescaleDB using pgx.CopyFrom.
func (w *BatchWriter) Flush(ctx context.Context) error {
	w.mu.Lock()
	if len(w.buffer) == 0 {
		w.mu.Unlock()
		return nil
	}

	// Swap buffer
	currentBatch := w.buffer
	w.buffer = make([]TelemetryRecord, 0, w.batchSize)
	w.mu.Unlock()

	if w.inserter == nil {
		return nil
	}

	startTime := time.Now()

	// Prepare CopyFromRows
	rows := make([][]any, len(currentBatch))
	for i, rec := range currentBatch {
		rows[i] = []any{
			rec.Time,
			ensureUUID(rec.PhysicalAssetID),
			rec.MetricName,
			rec.Value,
			rec.Quality,
		}
	}

	count, err := w.inserter.CopyFrom(
		ctx,
		pgx.Identifier{w.tableName},
		[]string{"time", "physical_asset_id", "metric_name", "value", "quality"},
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		return fmt.Errorf("CopyFrom failed: %w", err)
	}

	duration := time.Since(startTime)
	w.mu.Lock()
	w.totalInserted += count
	w.totalBatches++
	w.mu.Unlock()

	log.Printf("[BatchWriter] Successfully flushed %d records to %s in %v (Total: %d)",
		count, w.tableName, duration, w.totalInserted)
	return nil
}

func (w *BatchWriter) drainAndFlush(ctx context.Context) {
	for {
		select {
		case record := <-w.records:
			w.mu.Lock()
			w.buffer = append(w.buffer, record)
			w.mu.Unlock()
		default:
			goto flushRemaining
		}
	}

flushRemaining:
	if err := w.Flush(ctx); err != nil {
		log.Printf("[BatchWriter][ERROR] Final drain flush failed: %v", err)
	}
}

// Stop signals the writer to flush pending records and wait for worker completion.
func (w *BatchWriter) Stop() {
	close(w.done)
	w.wg.Wait()
	log.Println("[BatchWriter] Stopped gracefully.")
}

// Stats returns total inserted count, total batch count, and total dropped record count.
func (w *BatchWriter) Stats() (int64, int64) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.totalInserted, w.totalBatches
}

// DroppedCount returns the number of records dropped due to a full ingestion channel.
func (w *BatchWriter) DroppedCount() int64 {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.totalDropped
}

// ensureUUID formats or deterministically maps a string identifier into a valid pgtype.UUID.
func ensureUUID(id string) pgtype.UUID {
	var u pgtype.UUID
	if err := u.Scan(id); err == nil && u.Valid {
		return u
	}
	// Deterministic UUIDv5-like mapping from arbitrary slug string
	h := sha1.New()
	h.Write([]byte("factoryos-asset-namespace"))
	h.Write([]byte(id))
	bs := h.Sum(nil)[:16]
	bs[6] = (bs[6] & 0x0f) | 0x50 // version 5
	bs[8] = (bs[8] & 0x3f) | 0x80 // RFC 4122 variant
	copy(u.Bytes[:], bs)
	u.Valid = true
	return u
}

