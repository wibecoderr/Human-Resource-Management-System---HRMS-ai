package handler

import (
	"hrms/dbhelper"
	"hrms/utils"
	"net/http"
)

// POST /api/attendance/checkin
func CheckIn(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")

	alreadyIn, err := dbhelper.HasCheckedInToday(userID)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to check attendance")
		return
	}
	if alreadyIn {
		utils.RespondError(w, http.StatusConflict, nil, "You have already checked in today")
		return
	}

	record, err := dbhelper.CheckIn(userID)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to check in")
		return
	}

	utils.RespondJSON(w, http.StatusCreated, record)
}

// POST /api/attendance/checkout
func CheckOut(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")

	alreadyIn, err := dbhelper.HasCheckedInToday(userID)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to check attendance")
		return
	}
	if !alreadyIn {
		utils.RespondError(w, http.StatusBadRequest, nil, "You have not checked in today")
		return
	}

	alreadyOut, err := dbhelper.HasCheckedOutToday(userID)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to check attendance")
		return
	}
	if alreadyOut {
		utils.RespondError(w, http.StatusConflict, nil, "You have already checked out today")
		return
	}

	record, err := dbhelper.CheckOut(userID)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to check out")
		return
	}

	utils.RespondJSON(w, http.StatusOK, record)
}

// GET /api/attendance/me
func GetMyAttendance(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")

	records, err := dbhelper.GetMyAttendance(userID)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to fetch attendance")
		return
	}

	utils.RespondJSON(w, http.StatusOK, records)
}

// GET /api/attendance/{employeeId}
func GetEmployeeAttendance(w http.ResponseWriter, r *http.Request) {
	employeeID := r.PathValue("employeeId")
	if employeeID == "" {
		utils.RespondError(w, http.StatusBadRequest, nil, "employeeId is required")
		return
	}

	records, err := dbhelper.GetMyAttendance(employeeID)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to fetch attendance")
		return
	}

	utils.RespondJSON(w, http.StatusOK, records)
}
