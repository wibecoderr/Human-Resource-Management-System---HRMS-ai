package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"hrms/dbhelper"
	"hrms/model"
	"hrms/services"
	"hrms/utils"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func ApplyCandidate(w http.ResponseWriter, r *http.Request) {
	var req model.ApplyCandidateRequest
	if err := utils.ParseBody(r.Body, &req); err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "Failed to parse request body")
		return
	}
	if errs := utils.ValidateStruct(req); errs != nil {
		utils.RespondValidationError(w, errs)
		return
	}
	job, err := dbhelper.GetJobByID(req.JobID)
	if err == sql.ErrNoRows {
		utils.RespondError(w, http.StatusNotFound, nil, "Job not found")
		return
	}
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to verify job")
		return
	}
	if job.Status != "open" {
		utils.RespondError(w, http.StatusBadRequest, nil, "This job is no longer accepting applications")
		return
	}
	already, err := dbhelper.AlreadyApplied(req.Email, req.JobID)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to check application")
		return
	}
	if already {
		utils.RespondError(w, http.StatusConflict, nil, "You have already applied for this job")
		return
	}
	candidate, err := dbhelper.ApplyCandidate(req.Name, req.Email, req.Phone, req.JobID, req.ResumeURL, req.CoverLetter)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to submit application")
		return
	}
	utils.RespondJSON(w, http.StatusCreated, candidate)
}

func ApplyCandidateWithResume(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 10<<20)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "Application upload must be 10MB or smaller")
		return
	}

	req := model.ApplyCandidateRequest{
		Name:        strings.TrimSpace(r.FormValue("name")),
		Email:       strings.TrimSpace(r.FormValue("email")),
		Phone:       strings.TrimSpace(r.FormValue("phone")),
		JobID:       strings.TrimSpace(r.FormValue("job_id")),
		ResumeURL:   "pending-upload",
		CoverLetter: strings.TrimSpace(r.FormValue("cover_letter")),
	}
	if errs := utils.ValidateStruct(req); errs != nil {
		utils.RespondValidationError(w, errs)
		return
	}

	job, err := dbhelper.GetJobByID(req.JobID)
	if err == sql.ErrNoRows {
		utils.RespondError(w, http.StatusNotFound, nil, "Job not found")
		return
	}
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to verify job")
		return
	}
	if job.Status != "open" {
		utils.RespondError(w, http.StatusBadRequest, nil, "This job is no longer accepting applications")
		return
	}
	already, err := dbhelper.AlreadyApplied(req.Email, req.JobID)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to check application")
		return
	}
	if already {
		utils.RespondError(w, http.StatusConflict, nil, "You have already applied for this job")
		return
	}

	file, header, err := r.FormFile("resume")
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "PDF resume file is required")
		return
	}
	defer file.Close()

	if !strings.EqualFold(filepath.Ext(header.Filename), ".pdf") {
		utils.RespondError(w, http.StatusBadRequest, nil, "Only PDF resumes are allowed")
		return
	}
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		utils.RespondError(w, http.StatusBadRequest, err, "Failed to read resume")
		return
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to process resume")
		return
	}
	if http.DetectContentType(buffer[:n]) != "application/pdf" {
		utils.RespondError(w, http.StatusBadRequest, nil, "Uploaded file is not a valid PDF")
		return
	}

	candidate, err := dbhelper.ApplyCandidate(req.Name, req.Email, req.Phone, req.JobID, req.ResumeURL, req.CoverLetter)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to submit application")
		return
	}

	dir := filepath.Join("uploads", "resumes")
	if err := os.MkdirAll(dir, 0755); err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to prepare upload directory")
		return
	}
	filename := fmt.Sprintf("%s-%d.pdf", sanitizeFilePart(candidate.ID), time.Now().UnixNano())
	path := filepath.Join(dir, filename)
	dst, err := os.Create(path)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to store resume")
		return
	}
	defer dst.Close()
	if _, err := io.Copy(dst, file); err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to save resume")
		return
	}
	resumeText, err := services.ExtractResumeText(path)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "Failed to extract text from PDF")
		return
	}
	if strings.TrimSpace(resumeText) == "" {
		utils.RespondError(w, http.StatusBadRequest, nil, "PDF did not contain extractable text")
		return
	}

	updated, err := dbhelper.SaveCandidateResume(candidate.ID, "/"+filepath.ToSlash(path), resumeText)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to update candidate resume")
		return
	}
	utils.RespondJSON(w, http.StatusCreated, updated)
}

func GetCandidates(w http.ResponseWriter, r *http.Request) {
	jobID := r.URL.Query().Get("job_id")
	var (
		candidates []model.Candidate
		err        error
	)
	if jobID != "" {
		candidates, err = dbhelper.GetCandidatesByJob(jobID)
	} else {
		candidates, err = dbhelper.GetAllCandidates()
	}
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to fetch candidates")
		return
	}
	utils.RespondJSON(w, http.StatusOK, candidates)
}

func GetCandidate(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	candidate, err := dbhelper.GetCandidateByID(id)
	if err == sql.ErrNoRows {
		utils.RespondError(w, http.StatusNotFound, nil, "Candidate not found")
		return
	}
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to fetch candidate")
		return
	}
	utils.RespondJSON(w, http.StatusOK, candidate)
}

func UpdateCandidateStatus(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req model.UpdateCandidateStatusRequest
	if err := utils.ParseBody(r.Body, &req); err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "Failed to parse request body")
		return
	}
	if errs := utils.ValidateStruct(req); errs != nil {
		utils.RespondValidationError(w, errs)
		return
	}
	_, err := dbhelper.GetCandidateByID(id)
	if err == sql.ErrNoRows {
		utils.RespondError(w, http.StatusNotFound, nil, "Candidate not found")
		return
	}
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to fetch candidate")
		return
	}
	candidate, err := dbhelper.UpdateCandidateStatus(id, req.Status)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to update candidate status")
		return
	}
	utils.RespondJSON(w, http.StatusOK, candidate)
}

func DeleteCandidate(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	_, err := dbhelper.GetCandidateByID(id)
	if err == sql.ErrNoRows {
		utils.RespondError(w, http.StatusNotFound, nil, "Candidate not found")
		return
	}
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to fetch candidate")
		return
	}

	if err := dbhelper.DeleteCandidate(id); err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to delete candidate")
		return
	}

	utils.RespondJSON(w, http.StatusOK, map[string]string{"message": "Candidate deleted successfully"})
}

func UploadCandidateResume(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if _, err := dbhelper.GetCandidateByID(id); err == sql.ErrNoRows {
		utils.RespondError(w, http.StatusNotFound, nil, "Candidate not found")
		return
	} else if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to fetch candidate")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 10<<20)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "Resume upload must be 10MB or smaller")
		return
	}

	file, header, err := r.FormFile("resume")
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "PDF resume file is required")
		return
	}
	defer file.Close()

	if !strings.EqualFold(filepath.Ext(header.Filename), ".pdf") {
		utils.RespondError(w, http.StatusBadRequest, nil, "Only PDF resumes are allowed")
		return
	}

	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		utils.RespondError(w, http.StatusBadRequest, err, "Failed to read resume")
		return
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to process resume")
		return
	}
	if http.DetectContentType(buffer[:n]) != "application/pdf" {
		utils.RespondError(w, http.StatusBadRequest, nil, "Uploaded file is not a valid PDF")
		return
	}

	dir := filepath.Join("uploads", "resumes")
	if err := os.MkdirAll(dir, 0755); err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to prepare upload directory")
		return
	}

	filename := fmt.Sprintf("%s-%d.pdf", sanitizeFilePart(id), time.Now().UnixNano())
	path := filepath.Join(dir, filename)
	dst, err := os.Create(path)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to store resume")
		return
	}
	defer dst.Close()
	if _, err := io.Copy(dst, file); err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to save resume")
		return
	}

	resumeText, err := services.ExtractResumeText(path)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "Failed to extract text from PDF")
		return
	}
	if strings.TrimSpace(resumeText) == "" {
		utils.RespondError(w, http.StatusBadRequest, nil, "PDF did not contain extractable text")
		return
	}

	publicPath := "/" + filepath.ToSlash(path)
	candidate, err := dbhelper.SaveCandidateResume(id, publicPath, resumeText)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to update candidate resume")
		return
	}

	utils.RespondJSON(w, http.StatusOK, candidate)
}

func ScreenCandidateResume(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	candidate, err := dbhelper.GetCandidateByID(id)
	if err == sql.ErrNoRows {
		utils.RespondError(w, http.StatusNotFound, nil, "Candidate not found")
		return
	}
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to fetch candidate")
		return
	}
	if strings.TrimSpace(candidate.ResumeText) == "" {
		utils.RespondError(w, http.StatusBadRequest, nil, "Candidate resume text is missing. Upload a PDF resume first.")
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

	analysis, report, err := services.AnalyzeResume(job.Title, job.RequiredSkills, job.Description, candidate.ResumeText)
	if err != nil {
		utils.RespondError(w, http.StatusBadGateway, err, "Failed to analyze resume with Gemini")
		return
	}

	status := statusForScore(analysis.Score)
	updated, err := dbhelper.SaveCandidateAnalysis(id, analysis.Score, report, status)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to save resume analysis")
		return
	}

	utils.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"candidate": updated,
		"analysis":  analysis,
	})
}

func GetCandidateAnalysis(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	analysis, err := dbhelper.GetCandidateAnalysis(id)
	if err == sql.ErrNoRows {
		utils.RespondError(w, http.StatusNotFound, nil, "Candidate not found")
		return
	}
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to fetch candidate analysis")
		return
	}
	if len(analysis.Report) == 0 || string(analysis.Report) == "null" {
		analysis.Report = json.RawMessage(`{}`)
	}
	utils.RespondJSON(w, http.StatusOK, analysis)
}

func statusForScore(score float64) string {
	if score >= 80 {
		return "shortlisted"
	}
	if score >= 60 {
		return "interviewing"
	}
	return "rejected"
}

func sanitizeFilePart(value string) string {
	value = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			return r
		}
		return '-'
	}, value)
	return strings.Trim(value, "-")
}
