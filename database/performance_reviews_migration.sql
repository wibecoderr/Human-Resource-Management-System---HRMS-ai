ALTER TABLE performance_reviews ADD COLUMN IF NOT EXISTS status VARCHAR(20) NOT NULL DEFAULT 'draft';
ALTER TABLE performance_reviews ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

UPDATE performance_reviews
SET status = COALESCE(NULLIF(status, ''), 'draft');

UPDATE performance_reviews
SET updated_at = COALESCE(updated_at, created_at, NOW());
