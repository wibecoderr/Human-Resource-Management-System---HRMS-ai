package handler

import (
	"database/sql"
	"hrms/dbhelper"
	"hrms/model"
	"hrms/utils"
	"net/http"
)

func SetPayroll(w http.ResponseWriter, r *http.Request) {
	var req model.SetPayrollRequest
	if err := utils.ParseBody(r.Body, &req); err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "Failed to parse request body")
		return
	}
	if errs := utils.ValidateStruct(req); errs != nil {
		utils.RespondValidationError(w, errs)
		return
	}
	if req.BasicSalary <= 0 {
		utils.RespondError(w, http.StatusBadRequest, nil, "basic_salary must be greater than 0")
		return
	}
	if req.Month < 1 || req.Month > 12 {
		utils.RespondError(w, http.StatusBadRequest, nil, "month must be between 1 and 12")
		return
	}
	_, err := dbhelper.GetEmployeeByID(req.EmployeeID)
	if err == sql.ErrNoRows {
		utils.RespondError(w, http.StatusNotFound, nil, "Employee not found")
		return
	}
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to verify employee")
		return
	}
	payroll, err := dbhelper.SetPayroll(req.EmployeeID, req.Month, req.Year, req.BasicSalary)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to set payroll")
		return
	}
	utils.RespondJSON(w, http.StatusCreated, payroll)
}

func GetMyPayroll(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	records, err := dbhelper.GetMyPayroll(userID)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to fetch payroll")
		return
	}
	utils.RespondJSON(w, http.StatusOK, records)
}

func GetAllPayroll(w http.ResponseWriter, r *http.Request) {
	records, err := dbhelper.GetAllPayroll()
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to fetch payroll")
		return
	}
	utils.RespondJSON(w, http.StatusOK, records)
}

func GetEmployeePayroll(w http.ResponseWriter, r *http.Request) {
	employeeID := r.PathValue("employeeId")
	_, err := dbhelper.GetEmployeeByID(employeeID)
	if err == sql.ErrNoRows {
		utils.RespondError(w, http.StatusNotFound, nil, "Employee not found")
		return
	}
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to verify employee")
		return
	}
	records, err := dbhelper.GetPayrollByEmployee(employeeID)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to fetch payroll")
		return
	}
	utils.RespondJSON(w, http.StatusOK, records)
}

func UpdatePayroll(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req model.UpdatePayrollRequest
	if err := utils.ParseBody(r.Body, &req); err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "Failed to parse request body")
		return
	}
	if req.BasicSalary <= 0 {
		utils.RespondError(w, http.StatusBadRequest, nil, "basic_salary must be greater than 0")
		return
	}
	_, err := dbhelper.GetPayrollByID(id)
	if err == sql.ErrNoRows {
		utils.RespondError(w, http.StatusNotFound, nil, "Payroll record not found")
		return
	}
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to fetch payroll")
		return
	}
	updated, err := dbhelper.UpdatePayroll(id, req.BasicSalary)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to update payroll")
		return
	}
	utils.RespondJSON(w, http.StatusOK, updated)
}
