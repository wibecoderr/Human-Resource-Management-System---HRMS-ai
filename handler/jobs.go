package handler

import (
	"database/sql"
	"hrms/dbhelper"
	"hrms/model"
	"hrms/utils"
	"net/http"
)

func CreateJob(w http.ResponseWriter, r *http.Request) {
	postedBy := r.Header.Get("X-User-ID")
	var req model.CreateJobRequest
	if err := utils.ParseBody(r.Body, &req); err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "Failed to parse request body")
		return
	}
	if errs := utils.ValidateStruct(req); errs != nil {
		utils.RespondValidationError(w, errs)
		return
	}
	job, err := dbhelper.CreateJob(req.Title, req.Department, req.RequiredSkills, req.Description, req.Location, postedBy)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to create job")
		return
	}
	utils.RespondJSON(w, http.StatusCreated, job)
}

func GetJobs(w http.ResponseWriter, r *http.Request) {
	showAll := r.URL.Query().Get("all") == "true"
	role := r.Header.Get("X-User-Role")
	isStaff := role == "admin" || role == "hr_recruiter" || role == "senior_manager"

	var (
		jobs []model.Job
		err  error
	)
	if showAll && isStaff {
		jobs, err = dbhelper.GetAllJobs()
	} else {
		jobs, err = dbhelper.GetOpenJobs()
	}
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to fetch jobs")
		return
	}
	utils.RespondJSON(w, http.StatusOK, jobs)
}

func GetJob(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	job, err := dbhelper.GetJobByID(id)
	if err == sql.ErrNoRows {
		utils.RespondError(w, http.StatusNotFound, nil, "Job not found")
		return
	}
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to fetch job")
		return
	}
	utils.RespondJSON(w, http.StatusOK, job)
}

func UpdateJob(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req model.UpdateJobRequest
	if err := utils.ParseBody(r.Body, &req); err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "Failed to parse request body")
		return
	}
	_, err := dbhelper.GetJobByID(id)
	if err == sql.ErrNoRows {
		utils.RespondError(w, http.StatusNotFound, nil, "Job not found")
		return
	}
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to fetch job")
		return
	}
	job, err := dbhelper.UpdateJob(id, req.Title, req.Department, req.RequiredSkills, req.Description, req.Location, req.Status)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to update job")
		return
	}
	utils.RespondJSON(w, http.StatusOK, job)
}

func DeleteJob(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	_, err := dbhelper.GetJobByID(id)
	if err == sql.ErrNoRows {
		utils.RespondError(w, http.StatusNotFound, nil, "Job not found")
		return
	}
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to fetch job")
		return
	}
	if err := dbhelper.DeleteJob(id); err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to close job")
		return
	}
	utils.RespondJSON(w, http.StatusOK, map[string]string{"message": "Job closed successfully"})
}
