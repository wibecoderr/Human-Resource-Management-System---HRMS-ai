package dbhelper

import (
	"hrms/database"
	"hrms/model"
)

func SetPayroll(employeeID string, month, year int, basicSalary float64) (*model.Payroll, error) {
	var p model.Payroll
	dbEmployeeID, err := payrollEmployeeID(employeeID)
	if err != nil {
		return nil, err
	}
	err = database.DB.Get(&p, `
		INSERT INTO payroll (employee_id, month, year, basic_salary)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (employee_id, month, year)
		DO UPDATE SET basic_salary = EXCLUDED.basic_salary
		RETURNING id::text, $5::text AS employee_id, month, year, basic_salary, COALESCE(created_at, NOW()) AS created_at
	`, dbEmployeeID, month, year, basicSalary, employeeID)
	return &p, err
}

func GetMyPayroll(employeeID string) ([]model.Payroll, error) {
	var records []model.Payroll
	dbEmployeeID, err := payrollEmployeeID(employeeID)
	if err != nil {
		return nil, err
	}
	err = database.DB.Select(&records, `
		SELECT p.id::text, COALESCE(e.user_id::text, p.employee_id::text) AS employee_id, COALESCE(e.name, u.email) AS employee_name,
		       p.month, p.year, p.basic_salary, COALESCE(p.created_at, NOW()) AS created_at
		FROM payroll p
		LEFT JOIN employees e ON e.id = p.employee_id
		LEFT JOIN users u ON u.id = e.user_id
		WHERE p.employee_id::text = $1
		ORDER BY p.year DESC, p.month DESC
	`, dbEmployeeID)
	return records, err
}

func GetPayrollByEmployee(employeeID string) ([]model.Payroll, error) {
	var records []model.Payroll
	dbEmployeeID, err := payrollEmployeeID(employeeID)
	if err != nil {
		return nil, err
	}
	err = database.DB.Select(&records, `
		SELECT p.id::text, COALESCE(e.user_id::text, p.employee_id::text) AS employee_id, COALESCE(e.name, u.email) AS employee_name,
		       p.month, p.year, p.basic_salary, COALESCE(p.created_at, NOW()) AS created_at
		FROM payroll p
		LEFT JOIN employees e ON e.id = p.employee_id
		LEFT JOIN users u ON u.id = e.user_id
		WHERE p.employee_id::text = $1
		ORDER BY p.year DESC, p.month DESC
	`, dbEmployeeID)
	return records, err
}

func GetAllPayroll() ([]model.Payroll, error) {
	var records []model.Payroll
	err := database.DB.Select(&records, `
		SELECT p.id::text, COALESCE(e.user_id::text, p.employee_id::text) AS employee_id, COALESCE(e.name, u.email) AS employee_name,
		       p.month, p.year, p.basic_salary, COALESCE(p.created_at, NOW()) AS created_at
		FROM payroll p
		LEFT JOIN employees e ON e.id = p.employee_id
		LEFT JOIN users u ON u.id = e.user_id
		ORDER BY p.year DESC, p.month DESC, COALESCE(e.name, u.email) ASC
	`)
	return records, err
}

func UpdatePayroll(id string, basicSalary float64) (*model.Payroll, error) {
	var p model.Payroll
	err := database.DB.Get(&p, `
		UPDATE payroll SET basic_salary = $2
		WHERE id::text = $1
		RETURNING id::text, employee_id::text, month, year, basic_salary, COALESCE(created_at, NOW()) AS created_at
	`, id, basicSalary)
	if err != nil {
		return nil, err
	}
	return GetPayrollByID(p.ID)
}

func GetPayrollByID(id string) (*model.Payroll, error) {
	var p model.Payroll
	err := database.DB.Get(&p, `
		SELECT p.id::text, COALESCE(e.user_id::text, p.employee_id::text) AS employee_id, COALESCE(e.name, u.email) AS employee_name,
		       p.month, p.year, p.basic_salary, COALESCE(p.created_at, NOW()) AS created_at
		FROM payroll p
		LEFT JOIN employees e ON e.id = p.employee_id
		LEFT JOIN users u ON u.id = e.user_id
		WHERE p.id::text = $1
	`, id)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func payrollEmployeeID(employeeID string) (string, error) {
	if !tableExists("employees") {
		return employeeID, nil
	}
	var id string
	err := database.DB.Get(&id, `
		SELECT id::text
		FROM employees
		WHERE user_id::text = $1 OR id::text = $1
		ORDER BY CASE WHEN user_id::text = $1 THEN 0 ELSE 1 END
		LIMIT 1
	`, employeeID)
	if err != nil {
		return employeeID, nil
	}
	return id, nil
}
