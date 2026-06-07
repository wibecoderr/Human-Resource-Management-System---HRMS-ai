package handler

import (
	"database/sql"
	"hrms/dbhelper"
	"hrms/model"
	"hrms/utils"
	"net/http"
)

// POST /api/leave
// Any employee can apply
func ApplyLeave(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")

	var req model.ApplyLeaveRequest
	if err := utils.ParseBody(r.Body, &req); err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "Failed to parse request body")
		return
	}

	if errs := utils.ValidateStruct(req); errs != nil {
		utils.RespondValidationError(w, errs)
		return
	}

	leave, err := dbhelper.ApplyLeave(userID, req.LeaveType, req.StartDate, req.EndDate, req.Reason)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to apply for leave")
		return
	}

	utils.RespondJSON(w, http.StatusCreated, leave)
}

// GET /api/leave/me
// Employee sees their own leaves
func GetMyLeaves(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")

	leaves, err := dbhelper.GetMyLeaves(userID)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to fetch leaves")
		return
	}

	utils.RespondJSON(w, http.StatusOK, leaves)
}

// GET /api/leave
// Roles: admin, senior_manager, hr_recruiter — see all leaves
func GetAllLeaves(w http.ResponseWriter, r *http.Request) {
	leaves, err := dbhelper.GetAllLeaves()
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to fetch leaves")
		return
	}

	utils.RespondJSON(w, http.StatusOK, leaves)
}

// PUT /api/leave/{id}/approve
// Roles: admin, senior_manager — approve or reject
func UpdateLeaveStatus(w http.ResponseWriter, r *http.Request) {
	leaveID := r.PathValue("id")
	approverID := r.Header.Get("X-User-ID")

	var req model.ApproveLeaveRequest
	if err := utils.ParseBody(r.Body, &req); err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "Failed to parse request body")
		return
	}

	if errs := utils.ValidateStruct(req); errs != nil {
		utils.RespondValidationError(w, errs)
		return
	}

	_, err := dbhelper.GetLeaveByID(leaveID)
	if err == sql.ErrNoRows {
		utils.RespondError(w, http.StatusNotFound, nil, "Leave request not found")
		return
	}
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to fetch leave")
		return
	}

	leave, err := dbhelper.UpdateLeaveStatus(leaveID, req.Status, approverID)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to update leave status")
		return
	}

	utils.RespondJSON(w, http.StatusOK, leave)
}

// PUT /api/leave/{id}/reject
// Roles: admin, senior_manager
func RejectLeave(w http.ResponseWriter, r *http.Request) {
	leaveID := r.PathValue("id")
	approverID := r.Header.Get("X-User-ID")

	_, err := dbhelper.GetLeaveByID(leaveID)
	if err == sql.ErrNoRows {
		utils.RespondError(w, http.StatusNotFound, nil, "Leave request not found")
		return
	}
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to fetch leave")
		return
	}

	leave, err := dbhelper.UpdateLeaveStatus(leaveID, "rejected", approverID)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to reject leave")
		return
	}

	utils.RespondJSON(w, http.StatusOK, leave)
}
