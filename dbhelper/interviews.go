package dbhelper

import (
	"encoding/json"
	"hrms/database"
	"hrms/model"
)

func CreateInterview(candidateID, jobID string, questions json.RawMessage) (*model.Interview, error) {
	var id string
	if err := database.DB.Get(&id, `
		INSERT INTO interviews (candidate_id, job_id, questions, status)
		VALUES ($1, $2, $3::jsonb, 'in_progress')
		RETURNING id::text
	`, candidateID, jobID, string(questions)); err != nil {
		return nil, err
	}
	return GetInterviewByID(id)
}

func GetAllInterviews() ([]model.Interview, error) {
	var interviews []model.Interview
	err := database.DB.Select(&interviews, interviewSelect()+`
		ORDER BY i.created_at DESC
	`)
	return interviews, err
}

func GetInterviewByID(id string) (*model.Interview, error) {
	var interview model.Interview
	err := database.DB.Get(&interview, interviewSelect()+`
		WHERE i.id::text = $1
	`, id)
	if err != nil {
		return nil, err
	}
	return &interview, nil
}

func GetInterviewsByCandidate(candidateID string) ([]model.Interview, error) {
	var interviews []model.Interview
	err := database.DB.Select(&interviews, interviewSelect()+`
		WHERE i.candidate_id = $1
		ORDER BY i.created_at DESC
	`, candidateID)
	return interviews, err
}

func CompleteInterview(id string, answers json.RawMessage, evaluation *model.InterviewEvaluation, report json.RawMessage) (*model.Interview, error) {
	_, err := database.DB.Exec(`
		UPDATE interviews
		SET answers = $2::jsonb,
			technical_score = $3,
			communication_score = $4,
			problem_solving_score = $5,
			overall_score = $6,
			recommendation = $7,
			status = 'completed',
			ai_report = $8::jsonb,
			updated_at = NOW()
		WHERE id::text = $1
	`, id, string(answers), evaluation.TechnicalScore, evaluation.CommunicationScore, evaluation.ProblemSolvingScore, evaluation.OverallScore, evaluation.Recommendation, string(report))
	if err != nil {
		return nil, err
	}
	return GetInterviewByID(id)
}

func interviewSelect() string {
	return interviewSelectFrom("interviews i")
}

func interviewSelectFrom(from string) string {
	return `
		SELECT
			i.id::text,
			i.candidate_id,
			COALESCE(c.name, '') AS candidate_name,
			i.job_id,
			COALESCE(j.title, '') AS job_title,
			COALESCE(i.questions, '[]'::jsonb) AS questions,
			COALESCE(i.answers, '[]'::jsonb) AS answers,
			i.technical_score,
			i.communication_score,
			i.problem_solving_score,
			i.overall_score,
			COALESCE(i.recommendation, '') AS recommendation,
			COALESCE(i.status, 'in_progress') AS status,
			COALESCE(i.ai_report, '{}'::jsonb) AS ai_report,
			i.created_at,
			i.updated_at
		FROM ` + from + `
		LEFT JOIN candidates c ON c.id::text = i.candidate_id::text
		LEFT JOIN jobs j ON j.id::text = i.job_id::text
	`
}
