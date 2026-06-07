package database

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

var DB *sqlx.DB

func Connect() {
	sslMode := strings.TrimSpace(os.Getenv("DB_SSLMODE"))
	if sslMode == "" {
		sslMode = "require"
		if host := strings.ToLower(strings.TrimSpace(os.Getenv("DB_HOST"))); host == "localhost" || host == "127.0.0.1" {
			sslMode = "disable"
		}
	}
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		sslMode,
	)

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	DB = db
	log.Println("Database connected successfully")

	if err := EnsureAuthSchema(); err != nil {
		log.Fatalf("Failed to prepare auth schema: %v", err)
	}
	if err := EnsureRecruitmentSchema(); err != nil {
		log.Fatalf("Failed to prepare recruitment schema: %v", err)
	}
	if err := EnsureInterviewSchema(); err != nil {
		log.Fatalf("Failed to prepare interview schema: %v", err)
	}
}

func EnsureAuthSchema() error {
	stmts := []string{
		`CREATE EXTENSION IF NOT EXISTS pgcrypto`,
		`CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(255) NOT NULL,
			email VARCHAR(255) NOT NULL UNIQUE,
			role VARCHAR(50) NOT NULL,
			phone_no VARCHAR(30) NOT NULL DEFAULT '',
			status VARCHAR(20) NOT NULL DEFAULT 'active',
			password_hash TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS name VARCHAR(255) NOT NULL DEFAULT ''`,
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS phone_no VARCHAR(30) NOT NULL DEFAULT ''`,
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS status VARCHAR(20) NOT NULL DEFAULT 'active'`,
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS password_hash TEXT`,
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`,
		`CREATE TABLE IF NOT EXISTS employees (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID UNIQUE REFERENCES users(id) ON DELETE CASCADE,
			employee_code VARCHAR(50) NOT NULL UNIQUE,
			name VARCHAR(255) NOT NULL,
			email VARCHAR(255) NOT NULL,
			phone VARCHAR(30) NOT NULL DEFAULT '',
			department VARCHAR(100),
			designation VARCHAR(100),
			joining_date DATE NOT NULL DEFAULT CURRENT_DATE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS sessions (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			token TEXT,
			is_active BOOLEAN NOT NULL DEFAULT TRUE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			expires_at TIMESTAMPTZ NOT NULL
		)`,
		`ALTER TABLE sessions ADD COLUMN IF NOT EXISTS token TEXT`,
		`ALTER TABLE sessions ADD COLUMN IF NOT EXISTS is_active BOOLEAN NOT NULL DEFAULT TRUE`,
		`ALTER TABLE sessions ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`,
		`ALTER TABLE sessions ADD COLUMN IF NOT EXISTS expires_at TIMESTAMPTZ`,
	}

	for _, stmt := range stmts {
		if _, err := DB.Exec(stmt); err != nil {
			return err
		}
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	adminSQL := `
		INSERT INTO users (name, email, role, phone_no, status, password_hash)
		VALUES ('Admin', 'admin@hrms.com', 'admin', '0000000000', 'active', $1)
		ON CONFLICT (email) DO UPDATE
		SET
			name = COALESCE(NULLIF(users.name, ''), EXCLUDED.name),
			role = COALESCE(NULLIF(users.role, ''), EXCLUDED.role),
			phone_no = COALESCE(NULLIF(users.phone_no, ''), EXCLUDED.phone_no),
			status = COALESCE(NULLIF(users.status, ''), EXCLUDED.status),
			password_hash = COALESCE(users.password_hash, EXCLUDED.password_hash)
		RETURNING id::text
	`
	if tableColumnExists("users", "password") {
		adminSQL = `
			INSERT INTO users (name, email, role, phone_no, status, password_hash, password)
			VALUES ('Admin', 'admin@hrms.com', 'admin', '0000000000', 'active', $1, $2)
			ON CONFLICT (email) DO UPDATE
			SET
				name = COALESCE(NULLIF(users.name, ''), EXCLUDED.name),
				role = COALESCE(NULLIF(users.role, ''), EXCLUDED.role),
				phone_no = COALESCE(NULLIF(users.phone_no, ''), EXCLUDED.phone_no),
				status = COALESCE(NULLIF(users.status, ''), EXCLUDED.status),
				password_hash = COALESCE(users.password_hash, EXCLUDED.password_hash),
				password = COALESCE(users.password, EXCLUDED.password)
			RETURNING id::text
		`
	}

	var adminID string
	args := []interface{}{string(hashed)}
	if tableColumnExists("users", "password") {
		args = append(args, string(hashed))
	}
	if err := DB.Get(&adminID, adminSQL, args...); err != nil {
		return err
	}

	_, err = DB.Exec(`
		INSERT INTO employees (user_id, employee_code, name, email, phone, joining_date)
		VALUES ($1, 'ADMIN-001', 'Admin', 'admin@hrms.com', '0000000000', CURRENT_DATE)
		ON CONFLICT (user_id) DO UPDATE
		SET name = EXCLUDED.name, email = EXCLUDED.email, phone = EXCLUDED.phone
	`, adminID)
	return err
}

func tableColumnExists(tableName, columnName string) bool {
	var exists bool
	err := DB.Get(&exists, `
		SELECT EXISTS (
			SELECT 1
			FROM information_schema.columns
			WHERE table_schema = 'public'
			  AND table_name = $1
			  AND column_name = $2
		)
	`, tableName, columnName)
	return err == nil && exists
}

func EnsureRecruitmentSchema() error {
	stmts := []string{
		`ALTER TABLE jobs ADD COLUMN IF NOT EXISTS location VARCHAR(255) NOT NULL DEFAULT ''`,
		`ALTER TABLE jobs ADD COLUMN IF NOT EXISTS required_skills TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE candidates ADD COLUMN IF NOT EXISTS phone VARCHAR(30) NOT NULL DEFAULT ''`,
		`ALTER TABLE candidates ADD COLUMN IF NOT EXISTS resume_url TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE candidates ADD COLUMN IF NOT EXISTS resume_text TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE candidates ADD COLUMN IF NOT EXISTS ai_score NUMERIC(5, 2)`,
		`ALTER TABLE candidates ADD COLUMN IF NOT EXISTS ai_report JSONB NOT NULL DEFAULT '{}'::jsonb`,
		`ALTER TABLE candidates ADD COLUMN IF NOT EXISTS cover_letter TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE candidates ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`,
	}
	for _, stmt := range stmts {
		if _, err := DB.Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}

func EnsureInterviewSchema() error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS interviews (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			candidate_id TEXT NOT NULL,
			job_id TEXT NOT NULL,
			questions JSONB NOT NULL DEFAULT '[]'::jsonb,
			answers JSONB NOT NULL DEFAULT '[]'::jsonb,
			technical_score NUMERIC(5,2),
			communication_score NUMERIC(5,2),
			problem_solving_score NUMERIC(5,2),
			overall_score NUMERIC(5,2),
			recommendation VARCHAR(20) NOT NULL DEFAULT '',
			status VARCHAR(20) NOT NULL DEFAULT 'in_progress',
			ai_report JSONB NOT NULL DEFAULT '{}'::jsonb,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`ALTER TABLE interviews ADD COLUMN IF NOT EXISTS candidate_id TEXT`,
		`ALTER TABLE interviews ADD COLUMN IF NOT EXISTS job_id TEXT`,
		`ALTER TABLE interviews ADD COLUMN IF NOT EXISTS questions JSONB NOT NULL DEFAULT '[]'::jsonb`,
		`ALTER TABLE interviews ADD COLUMN IF NOT EXISTS answers JSONB NOT NULL DEFAULT '[]'::jsonb`,
		`ALTER TABLE interviews ADD COLUMN IF NOT EXISTS technical_score NUMERIC(5,2)`,
		`ALTER TABLE interviews ADD COLUMN IF NOT EXISTS communication_score NUMERIC(5,2)`,
		`ALTER TABLE interviews ADD COLUMN IF NOT EXISTS problem_solving_score NUMERIC(5,2)`,
		`ALTER TABLE interviews ADD COLUMN IF NOT EXISTS overall_score NUMERIC(5,2)`,
		`ALTER TABLE interviews ADD COLUMN IF NOT EXISTS recommendation VARCHAR(20) NOT NULL DEFAULT ''`,
		`ALTER TABLE interviews ADD COLUMN IF NOT EXISTS status VARCHAR(20) NOT NULL DEFAULT 'in_progress'`,
		`ALTER TABLE interviews ADD COLUMN IF NOT EXISTS ai_report JSONB NOT NULL DEFAULT '{}'::jsonb`,
		`ALTER TABLE interviews ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`,
		`ALTER TABLE interviews ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`,
		`CREATE INDEX IF NOT EXISTS idx_interviews_candidate_id ON interviews (candidate_id)`,
		`CREATE INDEX IF NOT EXISTS idx_interviews_job_id ON interviews (job_id)`,
		`CREATE INDEX IF NOT EXISTS idx_interviews_status ON interviews (status)`,
	}

	for _, stmt := range stmts {
		if _, err := DB.Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}

// Tx runs a function inside a transaction. Rolls back on error, commits on success.
func Tx(fn func(tx *sqlx.Tx) error) error {
	tx, err := DB.Beginx()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}
