package dbhelper

import (
	"hrms/database"
	"hrms/model"

	"github.com/jmoiron/sqlx"
)

// ─── List all employees ───────────────────────────────────────────────────────

func GetAllEmployees() ([]model.User, error) {
	var users []model.User
	if !columnExists("users", "phone_no") {
		err := database.DB.Select(&users, `
			SELECT u.id::text,
			       COALESCE(e.name, split_part(u.email, '@', 1)) AS name,
			       u.email,
			       u.role,
			       COALESCE(e.phone, '') AS phone_no,
			       CASE WHEN COALESCE(u.is_active, TRUE) THEN 'active' ELSE 'inactive' END AS status,
			       COALESCE(u.created_at, NOW()) AS created_at
			FROM users u
			LEFT JOIN employees e ON e.user_id = u.id
			ORDER BY COALESCE(u.created_at, NOW()) DESC
		`)
		return users, err
	}
	err := database.DB.Select(&users, `
		SELECT id::text, name, email, role, phone_no, status, created_at
		FROM users
		ORDER BY created_at DESC
	`)
	return users, err
}

// ─── Search employees by name or email ───────────────────────────────────────

func SearchEmployees(query string) ([]model.User, error) {
	var users []model.User
	like := "%" + query + "%"
	if !columnExists("users", "phone_no") {
		err := database.DB.Select(&users, `
			SELECT u.id::text,
			       COALESCE(e.name, split_part(u.email, '@', 1)) AS name,
			       u.email,
			       u.role,
			       COALESCE(e.phone, '') AS phone_no,
			       CASE WHEN COALESCE(u.is_active, TRUE) THEN 'active' ELSE 'inactive' END AS status,
			       COALESCE(u.created_at, NOW()) AS created_at
			FROM users u
			LEFT JOIN employees e ON e.user_id = u.id
			WHERE COALESCE(e.name, '') ILIKE $1 OR u.email ILIKE $1
			ORDER BY COALESCE(u.created_at, NOW()) DESC
		`, like)
		return users, err
	}
	err := database.DB.Select(&users, `
		SELECT id::text, name, email, role, phone_no, status, created_at
		FROM users
		WHERE name ILIKE $1 OR email ILIKE $1
		ORDER BY created_at DESC
	`, like)
	return users, err
}

// ─── Get single employee ──────────────────────────────────────────────────────

func GetEmployeeByID(id string) (*model.User, error) {
	var u model.User
	if !columnExists("users", "phone_no") {
		err := database.DB.Get(&u, `
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
		`, id)
		if err != nil {
			return nil, err
		}
		return &u, nil
	}
	err := database.DB.Get(&u, `
		SELECT id::text, name, email, role, phone_no, status, created_at
		FROM users WHERE id::text = $1
	`, id)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// ─── Create employee (admin use — any role allowed) ───────────────────────────

func CreateEmployee(tx *sqlx.Tx, name, email, role, phoneNo, hashedPassword string) (string, error) {
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

// ─── Update employee ──────────────────────────────────────────────────────────

func UpdateEmployee(id, name, phoneNo, role, status string) error {
	if !columnExists("users", "phone_no") {
		_, err := database.DB.Exec(`
			UPDATE users
			SET
				role = COALESCE(NULLIF($2, ''), role),
				is_active = CASE
					WHEN $3 = 'active' THEN TRUE
					WHEN $3 = 'inactive' THEN FALSE
					ELSE is_active
				END
			WHERE id::text = $1
		`, id, role, status)
		if err != nil {
			return err
		}
		_, err = database.DB.Exec(`
			UPDATE employees
			SET
				name = COALESCE(NULLIF($2, ''), name),
				phone = COALESCE(NULLIF($3, ''), phone)
			WHERE user_id::text = $1
		`, id, name, phoneNo)
		return err
	}
	_, err := database.DB.Exec(`
		UPDATE users
		SET
			name     = COALESCE(NULLIF($2, ''), name),
			phone_no = COALESCE(NULLIF($3, ''), phone_no),
			role     = COALESCE(NULLIF($4, ''), role),
			status   = COALESCE(NULLIF($5, ''), status)
		WHERE id::text = $1
	`, id, name, phoneNo, role, status)
	return err
}

// ─── Soft delete (set status = inactive) ─────────────────────────────────────

func DeleteEmployee(id string) error {
	if !columnExists("users", "status") {
		_, err := database.DB.Exec(`
			UPDATE users SET is_active = FALSE WHERE id::text = $1
		`, id)
		return err
	}
	_, err := database.DB.Exec(`
		UPDATE users SET status = 'inactive' WHERE id::text = $1
	`, id)
	return err
}
