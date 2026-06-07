package handler

import (
	"database/sql"
	"hrms/database"
	"hrms/dbhelper"
	"hrms/model"
	"hrms/utils"
	"net/http"

	"github.com/jmoiron/sqlx"
)

// GET /api/employees
// GET /api/employees?q=searchterm
// Roles: admin, senior_manager, hr_recruiter
func GetEmployees(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")

	var (
		employees []model.User
		err       error
	)

	if q != "" {
		employees, err = dbhelper.SearchEmployees(q)
	} else {
		employees, err = dbhelper.GetAllEmployees()
	}

	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to fetch employees")
		return
	}

	utils.RespondJSON(w, http.StatusOK, employees)
}

// GET /api/employees/{id}
// Admin/HR/Manager → any employee. Employee → only themselves.
func GetEmployee(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	role := r.Header.Get("X-User-Role")
	userID := r.Header.Get("X-User-ID")

	// Employee can only view their own profile
	if role == "employee" && id != userID {
		utils.RespondError(w, http.StatusForbidden, nil, "You can only view your own profile")
		return
	}

	employee, err := dbhelper.GetEmployeeByID(id)
	if err == sql.ErrNoRows {
		utils.RespondError(w, http.StatusNotFound, nil, "Employee not found")
		return
	}
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to fetch employee")
		return
	}

	utils.RespondJSON(w, http.StatusOK, employee)
}

// POST /api/employees
// Roles: admin only
func CreateEmployee(w http.ResponseWriter, r *http.Request) {
	var req model.CreateEmployeeRequest
	if err := utils.ParseBody(r.Body, &req); err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "Failed to parse request body")
		return
	}

	if errs := utils.ValidateStruct(req); errs != nil {
		utils.RespondValidationError(w, errs)
		return
	}

	exists, err := dbhelper.UserExist(req.Email)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to check user")
		return
	}
	if exists {
		utils.RespondError(w, http.StatusConflict, nil, "User with this email already exists")
		return
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to hash password")
		return
	}

	var userID string
	err = database.Tx(func(tx *sqlx.Tx) error {
		var txErr error
		userID, txErr = dbhelper.CreateEmployee(tx, req.Name, req.Email, req.Role, req.PhoneNo, hashedPassword)
		return txErr
	})
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to create employee")
		return
	}

	employee, err := dbhelper.GetEmployeeByID(userID)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Employee created but failed to fetch")
		return
	}

	utils.RespondJSON(w, http.StatusCreated, employee)
}

// PUT /api/employees/{id}
// Admin → can update all fields. Employee → can only update own name/phone.
func UpdateEmployee(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	role := r.Header.Get("X-User-Role")
	userID := r.Header.Get("X-User-ID")

	// Employees can only update themselves
	if role == "employee" && id != userID {
		utils.RespondError(w, http.StatusForbidden, nil, "You can only update your own profile")
		return
	}
	if role != "admin" && role != "hr_recruiter" && role != "senior_manager" && id != userID {
		utils.RespondError(w, http.StatusForbidden, nil, "You are not allowed to update this employee")
		return
	}

	var req model.UpdateEmployeeRequest
	if err := utils.ParseBody(r.Body, &req); err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "Failed to parse request body")
		return
	}

	// Non-admin cannot change role or status
	if role != "admin" {
		req.Role = ""
		req.Status = ""
	}

	// Check employee exists
	_, err := dbhelper.GetEmployeeByID(id)
	if err == sql.ErrNoRows {
		utils.RespondError(w, http.StatusNotFound, nil, "Employee not found")
		return
	}
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to fetch employee")
		return
	}

	if err := dbhelper.UpdateEmployee(id, req.Name, req.PhoneNo, req.Role, req.Status); err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to update employee")
		return
	}

	employee, _ := dbhelper.GetEmployeeByID(id)
	utils.RespondJSON(w, http.StatusOK, employee)
}

// DELETE /api/employees/{id}
// Roles: admin only — soft delete (status = inactive)
func DeleteEmployee(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	_, err := dbhelper.GetEmployeeByID(id)
	if err == sql.ErrNoRows {
		utils.RespondError(w, http.StatusNotFound, nil, "Employee not found")
		return
	}
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to fetch employee")
		return
	}

	if err := dbhelper.DeleteEmployee(id); err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to delete employee")
		return
	}

	utils.RespondJSON(w, http.StatusOK, map[string]string{"message": "Employee deactivated successfully"})
}
