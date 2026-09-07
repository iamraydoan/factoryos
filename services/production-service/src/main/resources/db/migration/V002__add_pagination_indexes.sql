-- Composite index for cursor-based pagination with createdAt DESC, id DESC sort.
-- Covers both single-key (id) and composite-key (created_at, id) pagination.
CREATE INDEX IF NOT EXISTS idx_work_orders_created_at_id ON work_orders(created_at, id);
