package db

import "github.com/jackc/pgx/v5/pgxpool"

// PostgresEquipmentRepository implements all repository interfaces using PostgreSQL.
// A single struct is used to share the connection pool across domains, which is
// required for transactional operations that span multiple tables (e.g., InstallAsset).
type PostgresEquipmentRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresEquipmentRepository creates a new repository with the given connection pool.
// The returned struct implements WorkUnitRepository, EquipmentClassRepository,
// CapabilityRepository, PhysicalAssetRepository, and InstallationRepository.
func NewPostgresEquipmentRepository(pool *pgxpool.Pool) *PostgresEquipmentRepository {
	return &PostgresEquipmentRepository{pool: pool}
}
