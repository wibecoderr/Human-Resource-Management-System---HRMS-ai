package dbhelper

import (
	"database/sql"
	"hrms/database"
	"hrms/model"
)

func CountActiveEmployees() (int, error) {
	var count int
	if !columnExists("users", "status") {
		err := database.DB.Get(&count, `SELECT COUNT(*) FROM users WHERE COALESCE(is_active, TRUE) = TRUE`)
		return count, err
	}
	err := database.DB.Get(&count, `SELECT COUNT(*) FROM users WHERE status = 'active'`)
	return count, err
}

func CountOpenJobs() (int, error) {
	var count int
	// If you don't have a jobs table yet, returns 0 gracefully
	err := database.DB.Get(&count, `
		SELECT COUNT(*) FROM jobs WHERE status = 'open'
	`)
	if err != nil {
		return 0, nil // table may not exist yet
	}
	return count, nil
}

func CountCandidates() (int, error) {
	var count int
	err := database.DB.Get(&count, `SELECT COUNT(*) FROM candidates`)
	if err != nil {
		return 0, nil // table may not exist yet
	}
	return count, nil
}

func CountPendingInterviews() (int, error) {
	var count int
	err := database.DB.Get(&count, `
		SELECT COUNT(*) FROM candidates WHERE status IN ('shortlisted', 'interviewing', 'interviewed')
	`)
	if err != nil {
		return 0, nil
	}
	return count, nil
}

func CountPerformanceReviews() (int, error) {
	var count int
	err := database.DB.Get(&count, `SELECT COUNT(*) FROM performance_reviews`)
	if err != nil {
		return 0, nil
	}
	return count, nil
}

func CountEmployeeAttendance(employeeID string) (int, error) {
	var count int
	err := database.DB.Get(&count, `
		SELECT COUNT(*) FROM attendance WHERE employee_id::text = $1
	`, employeeID)
	return count, err
}

func GetLeaveBalance(employeeID string) (int, error) {
	var used int
	if tableExists("leave_requests") && !tableExists("leaves") {
		err := database.DB.Get(&used, `
			SELECT COALESCE(SUM((end_date - start_date) + 1), 0)
			FROM leave_requests
			WHERE employee_id::text = $1 AND status = 'approved'
		`, employeeID)
		if err != nil {
			return 0, err
		}
		const annualAllowance = 24
		if used >= annualAllowance {
			return 0, nil
		}
		return annualAllowance - used, nil
	}
	err := database.DB.Get(&used, `
		SELECT COALESCE(SUM((end_date - start_date) + 1), 0)
		FROM leaves
		WHERE employee_id::text = $1 AND status = 'approved'
	`, employeeID)
	if err != nil {
		return 0, err
	}

	const annualAllowance = 24
	if used >= annualAllowance {
		return 0, nil
	}
	return annualAllowance - used, nil
}

func GetRecentPayroll(employeeID string) (*model.Payroll, error) {
	var payroll model.Payroll
	err := database.DB.Get(&payroll, `
		SELECT p.id::text, p.employee_id::text, COALESCE(e.name, u.email) AS employee_name,
		       p.month, p.year, p.basic_salary, COALESCE(p.created_at, NOW()) AS created_at
		FROM payroll p
		JOIN users u ON u.id = p.employee_id
		LEFT JOIN employees e ON e.user_id = u.id
		WHERE p.employee_id::text = $1
		ORDER BY p.year DESC, p.month DESC, p.created_at DESC
		LIMIT 1
	`, employeeID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &payroll, nil
}
