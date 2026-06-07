package handler

import (
	"database/sql"
	"encoding/json"
	"hrms/dbhelper"
	"hrms/model"
	"hrms/services"
	"hrms/utils"
	"net/http"
	"strings"
)

func StartInterview(w http.ResponseWriter, r *http.Request) {
	var req model.StartInterviewRequest
	if err := utils.ParseBody(r.Body, &req); err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "Failed to parse request body")
		return
	}
	if errs := utils.ValidateStruct(req); errs != nil {
		utils.RespondValidationError(w, errs)
		return
	}

	candidate, err := dbhelper.GetCandidateByID(req.CandidateID)
	if err == sql.ErrNoRows {
		utils.RespondError(w, http.StatusNotFound, nil, "Candidate not found")
		return
	}
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to fetch candidate")
		return
	}

	job, err := dbhelper.GetJobByID(candidate.JobID)
	if err == sql.ErrNoRows {
		utils.RespondError(w, http.StatusNotFound, nil, "Candidate job not found")
		return
	}
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to fetch job")
		return
	}

	resumeReport := json.RawMessage(candidate.AIReport)
	if len(resumeReport) == 0 || string(resumeReport) == "null" {
		resumeReport = json.RawMessage(`{}`)
	}

	_, questionsJSON, err := services.GenerateInterviewQuestions(job.Title, job.Description, resumeReport)
	if err != nil {
		utils.RespondError(w, http.StatusBadGateway, err, "Failed to generate interview questions with Gemini")
		return
	}

	interview, err := dbhelper.CreateInterview(candidate.ID, job.ID, questionsJSON)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to save interview")
		return
	}

	utils.RespondJSON(w, http.StatusCreated, interview)
}

func GetInterviews(w http.ResponseWriter, r *http.Request) {
	candidateID := strings.TrimSpace(r.URL.Query().Get("candidate_id"))
	var (
		interviews []model.Interview
		err        error
	)
	if candidateID != "" {
		interviews, err = dbhelper.GetInterviewsByCandidate(candidateID)
	} else {
		interviews, err = dbhelper.GetAllInterviews()
	}
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to fetch interviews")
		return
	}
	utils.RespondJSON(w, http.StatusOK, interviews)
}

func GetInterview(w http.ResponseWriter, r *http.Request) {
	interview, err := dbhelper.GetInterviewByID(r.PathValue("id"))
	if err == sql.ErrNoRows {
		utils.RespondError(w, http.StatusNotFound, nil, "Interview not found")
		return
	}
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to fetch interview")
		return
	}
	utils.RespondJSON(w, http.StatusOK, interview)
}

func SubmitInterview(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req model.SubmitInterviewRequest
	if err := utils.ParseBody(r.Body, &req); err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "Failed to parse request body")
		return
	}
	if len(req.Answers) == 0 {
		utils.RespondError(w, http.StatusBadRequest, nil, "Interview answers are required")
		return
	}
	for _, answer := range req.Answers {
		if strings.TrimSpace(answer.Question) == "" || strings.TrimSpace(answer.Answer) == "" {
			utils.RespondError(w, http.StatusBadRequest, nil, "Every answer must include a question and answer")
			return
		}
	}

	interview, err := dbhelper.GetInterviewByID(id)
	if err == sql.ErrNoRows {
		utils.RespondError(w, http.StatusNotFound, nil, "Interview not found")
		return
	}
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to fetch interview")
		return
	}

	candidate, err := dbhelper.GetCandidateByID(interview.CandidateID)
	if err == sql.ErrNoRows {
		utils.RespondError(w, http.StatusNotFound, nil, "Candidate not found")
		return
	}
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to fetch candidate")
		return
	}

	job, err := dbhelper.GetJobByID(interview.JobID)
	if err == sql.ErrNoRows {
		utils.RespondError(w, http.StatusNotFound, nil, "Interview job not found")
		return
	}
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to fetch job")
		return
	}

	resumeReport := json.RawMessage(candidate.AIReport)
	if len(resumeReport) == 0 || string(resumeReport) == "null" {
		resumeReport = json.RawMessage(`{}`)
	}

	evaluation, report, err := services.EvaluateInterviewAnswers(job.Title, job.Description, resumeReport, req.Answers)
	if err != nil {
		utils.RespondError(w, http.StatusBadGateway, err, "Failed to evaluate interview with Gemini")
		return
	}

	answersJSON, err := json.Marshal(req.Answers)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to save interview answers")
		return
	}
	updated, err := dbhelper.CompleteInterview(id, answersJSON, evaluation, report)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to save interview evaluation")
		return
	}

	utils.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"interview":  updated,
		"evaluation": evaluation,
	})
}
