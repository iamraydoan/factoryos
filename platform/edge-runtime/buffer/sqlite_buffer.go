package buffer

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	telemetryv1 "github.com/iamraydoan/factoryos/platform/platform-sdk/go/gen/telemetry/v1"
)

// Record represents a single telemetry payload stored in the SQLite outbox.
type Record struct {
	ID              int64
	PhysicalAssetID string
	PayloadBytes    []byte
	Status          string
	RetryCount      int
	CreatedAt       time.Time
}

// SQLiteBuffer manages an offline-resilient store-and-forward SQLite database.
type SQLiteBuffer struct {
	db *sql.DB
	mu sync.Mutex
}

// NewSQLiteBuffer initializes a new SQLite buffer with the required schema.
func NewSQLiteBuffer(dbPath string) (*SQLiteBuffer, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database at %s: %w", dbPath, err)
	}

	// Enable WAL mode for high performance concurrency
	if _, err := db.Exec("PRAGMA journal_mode=WAL;"); err != nil {
		// Non-fatal if WAL is not supported in memory mode
	}

	buf := &SQLiteBuffer{db: db}
	if err := buf.initSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize sqlite schema: %w", err)
	}

	return buf, nil
}

func (b *SQLiteBuffer) initSchema() error {
	query := `
	CREATE TABLE IF NOT EXISTS telemetry_outbox (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		physical_asset_id TEXT NOT NULL,
		payload_bytes BLOB NOT NULL,
		status TEXT NOT NULL DEFAULT 'PENDING',
		retry_count INTEGER NOT NULL DEFAULT 0,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_telemetry_outbox_status ON telemetry_outbox(status);
	`
	_, err := b.db.Exec(query)
	return err
}

// Enqueue serializes a TelemetryPayload and saves it to the SQLite store-and-forward buffer.
func (b *SQLiteBuffer) Enqueue(ctx context.Context, payload *telemetryv1.TelemetryPayload) (int64, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if payload.EdgeTimestamp == nil {
		payload.EdgeTimestamp = timestamppb.Now()
	}

	payloadBytes, err := proto.Marshal(payload)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal telemetry payload: %w", err)
	}

	query := `INSERT INTO telemetry_outbox (physical_asset_id, payload_bytes, status, created_at) VALUES (?, ?, 'PENDING', ?)`
	res, err := b.db.ExecContext(ctx, query, payload.PhysicalAssetId, payloadBytes, time.Now().UTC())
	if err != nil {
		return 0, fmt.Errorf("failed to insert telemetry record into sqlite buffer: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

// DequeueBatch retrieves up to limit pending records for cloud transmission.
func (b *SQLiteBuffer) DequeueBatch(ctx context.Context, limit int) ([]Record, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	query := `SELECT id, physical_asset_id, payload_bytes, status, retry_count, created_at FROM telemetry_outbox WHERE status = 'PENDING' ORDER BY id ASC LIMIT ?`
	rows, err := b.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query pending records from sqlite buffer: %w", err)
	}
	defer rows.Close()

	var records []Record
	for rows.Next() {
		var rec Record
		if err := rows.Scan(&rec.ID, &rec.PhysicalAssetID, &rec.PayloadBytes, &rec.Status, &rec.RetryCount, &rec.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan telemetry record row: %w", err)
		}
		records = append(records, rec)
	}

	return records, rows.Err()
}

// MarkSent removes or updates acknowledged records from the SQLite buffer.
func (b *SQLiteBuffer) MarkSent(ctx context.Context, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	tx, err := b.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `DELETE FROM telemetry_outbox WHERE id = ?`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, id := range ids {
		if _, err := stmt.ExecContext(ctx, id); err != nil {
			return fmt.Errorf("failed to delete record %d: %w", id, err)
		}
	}

	return tx.Commit()
}

// IncrementRetryCount increments the retry count for records that failed to sync.
func (b *SQLiteBuffer) IncrementRetryCount(ctx context.Context, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	tx, err := b.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `UPDATE telemetry_outbox SET retry_count = retry_count + 1 WHERE id = ?`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, id := range ids {
		if _, err := stmt.ExecContext(ctx, id); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// GetPendingCount returns the number of records currently queued in the SQLite buffer.
func (b *SQLiteBuffer) GetPendingCount(ctx context.Context) (int64, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	var count int64
	row := b.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM telemetry_outbox WHERE status = 'PENDING'`)
	if err := row.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

// Close closes the underlying SQLite database connection.
func (b *SQLiteBuffer) Close() error {
	return b.db.Close()
}
