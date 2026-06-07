package dbhelper

import (
	"crypto/rand"
	"encoding/hex"
	"hrms/database"
	"hrms/model"
	"time"

	"github.com/jmoiron/sqlx"
)

// ─── User existence check ────────────────────────────────────────────────────

func UserExist(email string) (bool, error) {
	var count int
	err := database.DB.Get(&count, `SELECT COUNT(*) FROM users WHERE email = $1`, email)
	return count > 0, err
}

func AddEmployee(tx *sqlx.Tx, name, email, role, phoneNo, hashedPassword string) (string, error) {
	var id string
	if !columnExists("users", "password_hash") {
		err := tx.Get(&id, `
			INSERT INTO users (email, password, role, is_active)
			VALUES ($1, $2, $3, TRUE)
			RETURNING id::text
		`, email, hashedPassword, role)
		if err != nil {
			return "", err
		}
		_, err = tx.Exec(`
			INSERT INTO employees (user_id, employee_code, name, email, phone, joining_date)
			VALUES ($1, $2, $3, $4, $5, CURRENT_DATE)
			ON CONFLICT DO NOTHING
		`, id, "EMP"+id, name, email, phoneNo)
		return id, err
	}
	if columnExists("users", "password") {
		err := tx.Get(&id, `
			INSERT INTO users (name, email, role, phone_no, password_hash, password)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id::text
		`, name, email, role, phoneNo, hashedPassword, hashedPassword)
		return id, err
	}
	err := tx.Get(&id, `
		INSERT INTO users (name, email, role, phone_no, password_hash)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id::text
	`, name, email, role, phoneNo, hashedPassword)
	return id, err
}

// ─── Session management ──────────────────────────────────────────────────────

func CreateSession(tx *sqlx.Tx, userID string) (string, error) {
	var sessionID string
	expiresAt := time.Now().Add(24 * time.Hour)
	tokenPlaceholder := "pending-" + uniqueTokenSuffix()
	err := tx.Get(&sessionID, `
		INSERT INTO sessions (user_id, expires_at, token)
		VALUES ($1, $2, $3)
		RETURNING id::text
	`, userID, expiresAt, tokenPlaceholder)
	return sessionID, err
}

func uniqueTokenSuffix() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return time.Now().Format("20060102150405.000000000")
	}
	return hex.EncodeToString(b)
}

func SessionValid(sessionID string) (bool, error) {
	var count int
	err := database.DB.Get(&count, `
		SELECT COUNT(*) FROM sessions
		WHERE id::text = $1 AND is_active = TRUE AND expires_at > NOW()
	`, sessionID)
	return count > 0, err
}

func InvalidateSession(sessionID string) error {
	_, err := database.DB.Exec(`
		UPDATE sessions SET is_active = FALSE WHERE id::text = $1
	`, sessionID)
	return err
}

func InvalidateAllUserSessions(userID string) error {
	_, err := database.DB.Exec(`
		UPDATE sessions SET is_active = FALSE WHERE user_id::text = $1
	`, userID)
	return err
}

type UserWithPassword struct {
	model.User
	PasswordHash string `db:"password_hash"`
}

func GetUserByEmail(email string) (*UserWithPassword, error) {
	var u UserWithPassword
	query := `
		SELECT id::text, name, email, role, phone_no, status, created_at, password_hash
		FROM users
		WHERE email = $1
	`
	if !columnExists("users", "password_hash") {
		query = `
			SELECT u.id::text,
			       COALESCE(e.name, split_part(u.email, '@', 1)) AS name,
			       u.email,
			       u.role,
			       COALESCE(e.phone, '') AS phone_no,
			       CASE WHEN COALESCE(u.is_active, TRUE) THEN 'active' ELSE 'inactive' END AS status,
			       COALESCE(u.created_at, NOW()) AS created_at,
			       u.password AS password_hash
			FROM users u
			LEFT JOIN employees e ON e.user_id = u.id
			WHERE u.email = $1
		`
	}
	err := database.DB.Get(&u, query, email)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func GetUserByID(userID string) (*model.User, error) {
	var u model.User
	query := `
		SELECT id::text, name, email, role, phone_no, status, created_at
		FROM users
		WHERE id::text = $1
	`
	if !columnExists("users", "phone_no") {
		query = `
			SELECT u.id::text,
			       COALESCE(e.name, split_part(u.email, '@', 1)) AS name,
			       u.email,
			       u.role,
			       COALESCE(e.phone, '') AS phone_no,
			       CASE WHEN COALESCE(u.is_active, TRUE) THEN 'active' ELSE 'inactive' END AS status,
			       COALESCE(u.created_at, NOW()) AS created_at
			FROM users u
			LEFT JOIN employees e ON e.user_id = u.id
			WHERE u.id::text = $1
		`
	}
	err := database.DB.Get(&u, query, userID)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func AttachTokenToSession(sessionID, token string) error {
	_, err := database.DB.Exec(`
		UPDATE sessions SET token = $1 WHERE id::text = $2
	`, token, sessionID)
	return err
}

func InvalidateSessionByToken(token string) error {
	_, err := database.DB.Exec(`
		UPDATE sessions SET is_active = FALSE WHERE token = $1
	`, token)
	return err
}

func DeleteExpiredSessions() error {
	_, err := database.DB.Exec(`DELETE FROM sessions WHERE expires_at < NOW()`)
	return err
}
