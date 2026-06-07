package dbhelper

import (
	"encoding/json"
	"hrms/database"
	"hrms/model"
)

func ApplyCandidate(name, email, phone, jobID, resumeURL, coverLetter string) (*model.Candidate, error) {
	var c model.Candidate
	if !columnExists("candidates", "cover_letter") {
		err := database.DB.Get(&c, `
			INSERT INTO candidates (name, email, phone, job_id, resume_url, status)
			VALUES ($1, $2, $3, $4, $5, 'applied')
			RETURNING id::text, name, email, COALESCE(phone, '') AS phone, job_id::text,
			          (SELECT title FROM jobs WHERE id = $4) AS job_title,
			          COALESCE(resume_url, '') AS resume_url,
			          COALESCE(resume_text, '') AS resume_text,
			          ai_score,
			          COALESCE(ai_report, '{}'::jsonb) AS ai_report,
			          '' AS cover_letter,
			          COALESCE(status, 'applied') AS status, COALESCE(applied_at, NOW()) AS created_at
		`, name, email, phone, jobID, resumeURL)
		return &c, err
	}
	err := database.DB.Get(&c, `
		INSERT INTO candidates (name, email, phone, job_id, resume_url, cover_letter, status)
		VALUES ($1, $2, $3, $4, $5, $6, 'applied')
		RETURNING id::text, name, email, phone, job_id::text,
		          (SELECT title FROM jobs WHERE id = $4) AS job_title,
		          resume_url, COALESCE(resume_text, '') AS resume_text, ai_score,
		          COALESCE(ai_report, '{}'::jsonb) AS ai_report, cover_letter, status, created_at
	`, name, email, phone, jobID, resumeURL, coverLetter)
	return &c, err
}

func AlreadyApplied(email, jobID string) (bool, error) {
	var count int
	err := database.DB.Get(&count, `
		SELECT COUNT(*) FROM candidates WHERE email = $1 AND job_id::text = $2
	`, email, jobID)
	return count > 0, err
}

func GetAllCandidates() ([]model.Candidate, error) {
	var candidates []model.Candidate
	if !columnExists("candidates", "cover_letter") {
		err := database.DB.Select(&candidates, `
			SELECT c.id::text, c.name, c.email, COALESCE(c.phone, '') AS phone, c.job_id::text,
			       j.title AS job_title,
			       COALESCE(c.resume_url, '') AS resume_url,
			       COALESCE(c.resume_text, '') AS resume_text,
			       c.ai_score,
			       COALESCE(c.ai_report, '{}'::jsonb) AS ai_report,
			       '' AS cover_letter,
			       COALESCE(c.status, 'applied') AS status, COALESCE(c.applied_at, NOW()) AS created_at
			FROM candidates c
			LEFT JOIN jobs j ON j.id = c.job_id
			ORDER BY c.applied_at DESC
		`)
		return candidates, err
	}
	err := database.DB.Select(&candidates, `
		SELECT c.id::text, c.name, c.email, COALESCE(c.phone, '') AS phone, c.job_id::text,
		       j.title AS job_title,
		       COALESCE(c.resume_url, '') AS resume_url, COALESCE(c.resume_text, '') AS resume_text, c.ai_score,
		       COALESCE(c.ai_report, '{}'::jsonb) AS ai_report, COALESCE(c.cover_letter, '') AS cover_letter,
		       COALESCE(c.status, 'applied') AS status, c.created_at
		FROM candidates c
		LEFT JOIN jobs j ON j.id = c.job_id
		ORDER BY c.created_at DESC
	`)
	return candidates, err
}

func GetCandidatesByJob(jobID string) ([]model.Candidate, error) {
	var candidates []model.Candidate
	if !columnExists("candidates", "cover_letter") {
		err := database.DB.Select(&candidates, `
			SELECT c.id::text, c.name, c.email, COALESCE(c.phone, '') AS phone, c.job_id::text,
			       j.title AS job_title,
			       COALESCE(c.resume_url, '') AS resume_url,
			       COALESCE(c.resume_text, '') AS resume_text,
			       c.ai_score,
			       COALESCE(c.ai_report, '{}'::jsonb) AS ai_report,
			       '' AS cover_letter,
			       COALESCE(c.status, 'applied') AS status, COALESCE(c.applied_at, NOW()) AS created_at
			FROM candidates c
			LEFT JOIN jobs j ON j.id = c.job_id
			WHERE c.job_id::text = $1
			ORDER BY c.applied_at DESC
		`, jobID)
		return candidates, err
	}
	err := database.DB.Select(&candidates, `
		SELECT c.id::text, c.name, c.email, COALESCE(c.phone, '') AS phone, c.job_id::text,
		       j.title AS job_title,
		       COALESCE(c.resume_url, '') AS resume_url, COALESCE(c.resume_text, '') AS resume_text, c.ai_score,
		       COALESCE(c.ai_report, '{}'::jsonb) AS ai_report, COALESCE(c.cover_letter, '') AS cover_letter,
		       COALESCE(c.status, 'applied') AS status, c.created_at
		FROM candidates c
		LEFT JOIN jobs j ON j.id = c.job_id
		WHERE c.job_id::text = $1
		ORDER BY c.created_at DESC
	`, jobID)
	return candidates, err
}

func GetCandidateByID(id string) (*model.Candidate, error) {
	var c model.Candidate
	if !columnExists("candidates", "cover_letter") {
		err := database.DB.Get(&c, `
			SELECT c.id::text, c.name, c.email, COALESCE(c.phone, '') AS phone, c.job_id::text,
			       j.title AS job_title,
			       COALESCE(c.resume_url, '') AS resume_url,
			       COALESCE(c.resume_text, '') AS resume_text,
			       c.ai_score,
			       COALESCE(c.ai_report, '{}'::jsonb) AS ai_report,
			       '' AS cover_letter,
			       COALESCE(c.status, 'applied') AS status, COALESCE(c.applied_at, NOW()) AS created_at
			FROM candidates c
			LEFT JOIN jobs j ON j.id = c.job_id
			WHERE c.id::text = $1
		`, id)
		if err != nil {
			return nil, err
		}
		return &c, nil
	}
	err := database.DB.Get(&c, `
		SELECT c.id::text, c.name, c.email, COALESCE(c.phone, '') AS phone, c.job_id::text,
		       j.title AS job_title,
		       COALESCE(c.resume_url, '') AS resume_url, COALESCE(c.resume_text, '') AS resume_text, c.ai_score,
		       COALESCE(c.ai_report, '{}'::jsonb) AS ai_report, COALESCE(c.cover_letter, '') AS cover_letter,
		       COALESCE(c.status, 'applied') AS status, c.created_at
		FROM candidates c
		LEFT JOIN jobs j ON j.id = c.job_id
		WHERE c.id::text = $1
	`, id)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func UpdateCandidateStatus(id, status string) (*model.Candidate, error) {
	_, err := database.DB.Exec(`UPDATE candidates SET status = $2 WHERE id::text = $1`, id, status)
	if err != nil {
		return nil, err
	}
	return GetCandidateByID(id)
}

func DeleteCandidate(id string) error {
	_, err := database.DB.Exec(`DELETE FROM candidates WHERE id::text = $1`, id)
	return err
}

func SaveCandidateResume(id, resumeURL, resumeText string) (*model.Candidate, error) {
	_, err := database.DB.Exec(`
		UPDATE candidates
		SET resume_url = $2, resume_text = $3
		WHERE id::text = $1
	`, id, resumeURL, resumeText)
	if err != nil {
		return nil, err
	}
	return GetCandidateByID(id)
}

func SaveCandidateAnalysis(id string, score float64, report json.RawMessage, status string) (*model.Candidate, error) {
	_, err := database.DB.Exec(`
		UPDATE candidates
		SET ai_score = $2, ai_report = $3::jsonb, status = $4
		WHERE id::text = $1
	`, id, score, string(report), status)
	if err != nil {
		return nil, err
	}
	return GetCandidateByID(id)
}

func GetCandidateAnalysis(id string) (*model.CandidateAnalysisResponse, error) {
	var row struct {
		CandidateID string      `db:"candidate_id"`
		Score       float64     `db:"score"`
		Report      model.JSONB `db:"report"`
	}
	err := database.DB.Get(&row, `
		SELECT id::text AS candidate_id,
		       COALESCE(ai_score, 0) AS score,
		       COALESCE(ai_report, '{}'::jsonb) AS report
		FROM candidates
		WHERE id::text = $1
	`, id)
	if err != nil {
		return nil, err
	}
	return &model.CandidateAnalysisResponse{
		CandidateID: row.CandidateID,
		Score:       row.Score,
		Report:      json.RawMessage(row.Report),
	}, nil
}
