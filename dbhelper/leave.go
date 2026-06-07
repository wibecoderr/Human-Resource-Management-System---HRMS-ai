package dbhelper

import (
	"hrms/database"
	"hrms/model"
)

// ─── Apply leave ──────────────────────────────────────────────────────────────

func ApplyLeave(employeeID, leaveType, startDate, endDate, reason string) (*model.Leave, error) {
	var l model.Leave
	if tableExists("leave_requests") && !tableExists("leaves") {
		dbEmployeeID, err := employeeRecordID(employeeID)
		if err != nil {
			return nil, err
		}
		err = database.DB.Get(&l, `
			INSERT INTO leave_requests (employee_id, leave_type, start_date, end_date, reason, status)
			VALUES ($1, $2, $3, $4, $5, 'pending')
			RETURNING id::text, $6::text AS employee_id, leave_type, start_date, end_date, COALESCE(reason, '') AS reason,
			          COALESCE(status, 'pending') AS status, approved_by::text, COALESCE(applied_at, NOW()) AS created_at
		`, dbEmployeeID, dbLeaveType(leaveType), startDate, endDate, reason, employeeID)
		return &l, err
	}
	err := database.DB.Get(&l, `
		INSERT INTO leaves (employee_id, leave_type, start_date, end_date, reason, status)
		VALUES ($1, $2, $3, $4, $5, 'pending')
		RETURNING id::text, employee_id::text, leave_type, start_date, end_date, reason, status, approved_by::text, created_at
	`, employeeID, leaveType, startDate, endDate, reason)
	return &l, err
}

// ─── Get my leaves ────────────────────────────────────────────────────────────

func GetMyLeaves(employeeID string) ([]model.Leave, error) {
	var leaves []model.Leave
	if tableExists("leave_requests") && !tableExists("leaves") {
		dbEmployeeID, err := employeeRecordID(employeeID)
		if err != nil {
			return nil, err
		}
		err = database.DB.Select(&leaves, `
			SELECT lr.id::text, COALESCE(e.user_id::text, lr.employee_id::text) AS employee_id, lr.leave_type, lr.start_date, lr.end_date, COALESCE(lr.reason, '') AS reason,
			       COALESCE(status, 'pending') AS status, approved_by::text, COALESCE(applied_at, NOW()) AS created_at
			FROM leave_requests lr
			LEFT JOIN employees e ON e.id = lr.employee_id
			WHERE lr.employee_id::text = $1
			ORDER BY lr.applied_at DESC
		`, dbEmployeeID)
		return leaves, err
	}
	err := database.DB.Select(&leaves, `
		SELECT id::text, employee_id::text, leave_type, start_date, end_date, reason, status, approved_by::text, created_at
		FROM leaves
		WHERE employee_id = $1
		ORDER BY created_at DESC
	`, employeeID)
	return leaves, err
}

// ─── Get all pending leaves (for HR / manager) ────────────────────────────────

func GetAllLeaves() ([]model.Leave, error) {
	var leaves []model.Leave
	if tableExists("leave_requests") && !tableExists("leaves") {
		err := database.DB.Select(&leaves, `
			SELECT lr.id::text, COALESCE(e.user_id::text, lr.employee_id::text) AS employee_id, lr.leave_type, lr.start_date, lr.end_date, COALESCE(lr.reason, '') AS reason,
			       COALESCE(lr.status, 'pending') AS status, approver.user_id::text AS approved_by, COALESCE(lr.applied_at, NOW()) AS created_at
			FROM leave_requests lr
			LEFT JOIN employees e ON e.id = lr.employee_id
			LEFT JOIN employees approver ON approver.id = lr.approved_by
			ORDER BY lr.applied_at DESC
		`)
		return leaves, err
	}
	err := database.DB.Select(&leaves, `
		SELECT id::text, employee_id::text, leave_type, start_date, end_date, reason, status, approved_by::text, created_at
		FROM leaves
		ORDER BY created_at DESC
	`)
	return leaves, err
}

// ─── Get leave by ID ─────────────────────────────────────────────────────────

func GetLeaveByID(id string) (*model.Leave, error) {
	var l model.Leave
	if tableExists("leave_requests") && !tableExists("leaves") {
		err := database.DB.Get(&l, `
			SELECT lr.id::text, COALESCE(e.user_id::text, lr.employee_id::text) AS employee_id, lr.leave_type, lr.start_date, lr.end_date, COALESCE(lr.reason, '') AS reason,
			       COALESCE(lr.status, 'pending') AS status, approver.user_id::text AS approved_by, COALESCE(lr.applied_at, NOW()) AS created_at
			FROM leave_requests lr
			LEFT JOIN employees e ON e.id = lr.employee_id
			LEFT JOIN employees approver ON approver.id = lr.approved_by
			WHERE lr.id::text = $1
		`, id)
		if err != nil {
			return nil, err
		}
		return &l, nil
	}
	err := database.DB.Get(&l, `
		SELECT id::text, employee_id::text, leave_type, start_date, end_date, reason, status, approved_by::text, created_at
		FROM leaves WHERE id::text = $1
	`, id)
	if err != nil {
		return nil, err
	}
	return &l, nil
}

// ─── Approve or reject leave ──────────────────────────────────────────────────

func UpdateLeaveStatus(leaveID, status, approverID string) (*model.Leave, error) {
	var leave model.Leave
	if tableExists("leave_requests") && !tableExists("leaves") {
		dbApproverID, err := employeeRecordID(approverID)
		if err != nil {
			return nil, err
		}
		err = database.DB.Get(&leave, `
			UPDATE leave_requests
			SET status = $2, approved_by = $3, updated_at = NOW()
			WHERE id::text = $1
			RETURNING id::text, employee_id::text, leave_type, start_date, end_date, COALESCE(reason, '') AS reason,
			          COALESCE(status, 'pending') AS status, $4::text AS approved_by, COALESCE(applied_at, NOW()) AS created_at
		`, leaveID, status, dbApproverID, approverID)
		if err != nil {
			return nil, err
		}
		return &leave, nil
	}
	err := database.DB.Get(&leave, `
		UPDATE leaves
		SET status = $2, approved_by = $3
		WHERE id::text = $1
		RETURNING id::text, employee_id::text, leave_type, start_date, end_date, reason, status, approved_by::text, created_at
	`, leaveID, status, approverID)
	if err != nil {
		return nil, err
	}
	return &leave, nil
}

func employeeRecordID(userOrEmployeeID string) (string, error) {
	if !tableExists("employees") {
		return userOrEmployeeID, nil
	}
	var id string
	err := database.DB.Get(&id, `
		SELECT id::text
		FROM employees
		WHERE user_id::text = $1 OR id::text = $1
		ORDER BY CASE WHEN user_id::text = $1 THEN 0 ELSE 1 END
		LIMIT 1
	`, userOrEmployeeID)
	if err != nil {
		return userOrEmployeeID, nil
	}
	return id, nil
}

func dbLeaveType(leaveType string) string {
	if leaveType == "earned" {
		return "paid"
	}
	return leaveType
}

// ─── Count pending leaves (for dashboard) ─────────────────────────────────────

func CountPendingLeaves() (int, error) {
	var count int
	if tableExists("leave_requests") && !tableExists("leaves") {
		err := database.DB.Get(&count, `SELECT COUNT(*) FROM leave_requests WHERE status = 'pending'`)
		return count, err
	}
	err := database.DB.Get(&count, `SELECT COUNT(*) FROM leaves WHERE status = 'pending'`)
	return count, err
}
