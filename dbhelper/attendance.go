package dbhelper

import (
	"hrms/database"
	"hrms/model"
)

// ─── Check In ─────────────────────────────────────────────────────────────────

func CheckIn(employeeID string) (*model.Attendance, error) {
	var a model.Attendance
	err := database.DB.Get(&a, `
		INSERT INTO attendance (employee_id, check_in, date)
		VALUES ($1, NOW(), CURRENT_DATE)
		RETURNING id::text, employee_id::text, check_in, check_out, date, COALESCE(created_at, NOW()) AS created_at
	`, employeeID)
	return &a, err
}

// ─── Already checked in today? ────────────────────────────────────────────────

func HasCheckedInToday(employeeID string) (bool, error) {
	var count int
	err := database.DB.Get(&count, `
		SELECT COUNT(*) FROM attendance
		WHERE employee_id::text = $1 AND date = CURRENT_DATE
	`, employeeID)
	return count > 0, err
}

// ─── Already checked out today? ───────────────────────────────────────────────

func HasCheckedOutToday(employeeID string) (bool, error) {
	var count int
	err := database.DB.Get(&count, `
		SELECT COUNT(*) FROM attendance
		WHERE employee_id::text = $1 AND date = CURRENT_DATE AND check_out IS NOT NULL
	`, employeeID)
	return count > 0, err
}

// ─── Check Out ────────────────────────────────────────────────────────────────

func CheckOut(employeeID string) (*model.Attendance, error) {
	var a model.Attendance
	err := database.DB.Get(&a, `
		UPDATE attendance
		SET check_out = NOW()
		WHERE employee_id::text = $1 AND date = CURRENT_DATE AND check_out IS NULL
		RETURNING id::text, employee_id::text, check_in, check_out, date, COALESCE(created_at, NOW()) AS created_at
	`, employeeID)
	return &a, err
}

// ─── Get my attendance ────────────────────────────────────────────────────────

func GetMyAttendance(employeeID string) ([]model.Attendance, error) {
	var records []model.Attendance
	err := database.DB.Select(&records, `
		SELECT id::text, employee_id::text, check_in, check_out, date, COALESCE(created_at, NOW()) AS created_at
		FROM attendance
		WHERE employee_id::text = $1
		ORDER BY date DESC
	`, employeeID)
	return records, err
}
