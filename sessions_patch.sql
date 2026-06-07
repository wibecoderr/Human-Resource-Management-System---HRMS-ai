-- Run this in pgAdmin if you already ran schema.sql earlier.
-- This makes sessions.token nullable (we fill it AFTER JWT is generated).

ALTER TABLE sessions ALTER COLUMN token DROP NOT NULL;

-- Verify
SELECT column_name, is_nullable FROM information_schema.columns
WHERE table_name = 'sessions';
