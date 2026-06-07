package handler

import (
	"hrms/dbhelper"
	"hrms/model"
	"hrms/utils"
	"net/http"
)

func CreatePerformanceReview(w http.ResponseWriter, r *http.Request) {
	reviewerID := r.Header.Get("X-User-ID")

	var req model.CreatePerformanceReviewRequest
	if err := utils.ParseBody(r.Body, &req); err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "Failed to parse request body")
		return
	}

	if errs := utils.ValidateStruct(req); errs != nil {
		utils.RespondValidationError(w, errs)
		return
	}
	if req.Rating < 1 || req.Rating > 5 {
		utils.RespondError(w, http.StatusBadRequest, nil, "rating must be between 1 and 5")
		return
	}

	review, err := dbhelper.CreatePerformanceReview(req.EmployeeID, reviewerID, req.ReviewPeriod, req.Rating, req.Feedback, req.Goals, req.Status)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to create performance review")
		return
	}

	utils.RespondJSON(w, http.StatusCreated, review)
}

func GetMyPerformanceReviews(w http.ResponseWriter, r *http.Request) {
	employeeID := r.Header.Get("X-User-ID")

	reviews, err := dbhelper.GetPerformanceReviewsByEmployee(employeeID)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to fetch performance reviews")
		return
	}

	utils.RespondJSON(w, http.StatusOK, reviews)
}

func GetAllPerformanceReviews(w http.ResponseWriter, r *http.Request) {
	reviews, err := dbhelper.GetAllPerformanceReviews()
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to fetch performance reviews")
		return
	}

	utils.RespondJSON(w, http.StatusOK, reviews)
}

func GetEmployeePerformanceReviews(w http.ResponseWriter, r *http.Request) {
	employeeID := r.PathValue("employeeId")
	if employeeID == "" {
		utils.RespondError(w, http.StatusBadRequest, nil, "employeeId is required")
		return
	}

	reviews, err := dbhelper.GetPerformanceReviewsByEmployee(employeeID)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to fetch performance reviews")
		return
	}

	utils.RespondJSON(w, http.StatusOK, reviews)
}

func UpdatePerformanceReview(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		utils.RespondError(w, http.StatusBadRequest, nil, "id is required")
		return
	}

	var req model.UpdatePerformanceReviewRequest
	if err := utils.ParseBody(r.Body, &req); err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "Failed to parse request body")
		return
	}
	if req.Rating < 0 || req.Rating > 5 {
		utils.RespondError(w, http.StatusBadRequest, nil, "rating must be between 1 and 5")
		return
	}

	review, err := dbhelper.UpdatePerformanceReview(id, req.ReviewPeriod, req.Rating, req.Feedback, req.Goals, req.Status)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to update performance review")
		return
	}

	utils.RespondJSON(w, http.StatusOK, review)
}

func DeletePerformanceReview(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		utils.RespondError(w, http.StatusBadRequest, nil, "id is required")
		return
	}

	if err := dbhelper.DeletePerformanceReview(id); err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to delete performance review")
		return
	}

	utils.RespondJSON(w, http.StatusOK, map[string]string{"message": "Performance review deleted successfully"})
}
