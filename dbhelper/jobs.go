package dbhelper

import (
	"hrms/database"
	"hrms/model"
)

func CreateJob(title, department, requiredSkills, description, location, postedBy string) (*model.Job, error) {
	var j model.Job
	if !columnExists("jobs", "location") {
		err := database.DB.Get(&j, `
			INSERT INTO jobs (title, department, required_skills, description, status, posted_by)
			VALUES ($1, $2, $3, $4, 'open', $5)
			RETURNING id::text, title, COALESCE(department, '') AS department,
			          COALESCE(required_skills, '') AS required_skills,
			          COALESCE(description, '') AS description, '' AS location,
			          COALESCE(status, 'open') AS status, posted_by::text, COALESCE(created_at, NOW()) AS created_at
		`, title, department, requiredSkills, description, postedBy)
		return &j, err
	}
	err := database.DB.Get(&j, `
		INSERT INTO jobs (title, department, required_skills, description, location, status, posted_by)
		VALUES ($1, $2, $3, $4, $5, 'open', $6)
		RETURNING id::text, title, department, COALESCE(required_skills, '') AS required_skills, description, location, status, posted_by::text, created_at
	`, title, department, requiredSkills, description, location, postedBy)
	return &j, err
}

func GetAllJobs() ([]model.Job, error) {
	var jobs []model.Job
	if !columnExists("jobs", "location") {
		err := database.DB.Select(&jobs, `
			SELECT id::text, title, COALESCE(department, '') AS department,
			       COALESCE(required_skills, '') AS required_skills,
			       COALESCE(description, '') AS description, '' AS location,
			       COALESCE(status, 'open') AS status, posted_by::text, COALESCE(created_at, NOW()) AS created_at
			FROM jobs ORDER BY created_at DESC
		`)
		return jobs, err
	}
	err := database.DB.Select(&jobs, `
		SELECT id::text, title, department, COALESCE(required_skills, '') AS required_skills, description, location, status, posted_by::text, created_at
		FROM jobs ORDER BY created_at DESC
	`)
	return jobs, err
}

func GetOpenJobs() ([]model.Job, error) {
	var jobs []model.Job
	if !columnExists("jobs", "location") {
		err := database.DB.Select(&jobs, `
			SELECT id::text, title, COALESCE(department, '') AS department,
			       COALESCE(required_skills, '') AS required_skills,
			       COALESCE(description, '') AS description, '' AS location,
			       COALESCE(status, 'open') AS status, posted_by::text, COALESCE(created_at, NOW()) AS created_at
			FROM jobs WHERE COALESCE(status, 'open') = 'open' ORDER BY created_at DESC
		`)
		return jobs, err
	}
	err := database.DB.Select(&jobs, `
		SELECT id::text, title, department, COALESCE(required_skills, '') AS required_skills, description, location, status, posted_by::text, created_at
		FROM jobs WHERE status = 'open' ORDER BY created_at DESC
	`)
	return jobs, err
}

func GetJobByID(id string) (*model.Job, error) {
	var j model.Job
	if !columnExists("jobs", "location") {
		err := database.DB.Get(&j, `
			SELECT id::text, title, COALESCE(department, '') AS department,
			       COALESCE(required_skills, '') AS required_skills,
			       COALESCE(description, '') AS description, '' AS location,
			       COALESCE(status, 'open') AS status, posted_by::text, COALESCE(created_at, NOW()) AS created_at
			FROM jobs WHERE id::text = $1
		`, id)
		if err != nil {
			return nil, err
		}
		return &j, nil
	}
	err := database.DB.Get(&j, `
		SELECT id::text, title, department, COALESCE(required_skills, '') AS required_skills, description, location, status, posted_by::text, created_at
		FROM jobs WHERE id::text = $1
	`, id)
	if err != nil {
		return nil, err
	}
	return &j, nil
}

func UpdateJob(id, title, department, requiredSkills, description, location, status string) (*model.Job, error) {
	var j model.Job
	if !columnExists("jobs", "location") {
		err := database.DB.Get(&j, `
			UPDATE jobs SET
				title       = COALESCE(NULLIF($2, ''), title),
				department  = COALESCE(NULLIF($3, ''), department),
				required_skills = COALESCE(NULLIF($4, ''), required_skills),
				description = COALESCE(NULLIF($5, ''), description),
				status      = COALESCE(NULLIF($6, ''), status)
			WHERE id::text = $1
			RETURNING id::text, title, COALESCE(department, '') AS department,
			          COALESCE(required_skills, '') AS required_skills,
			          COALESCE(description, '') AS description, '' AS location,
			          COALESCE(status, 'open') AS status, posted_by::text, COALESCE(created_at, NOW()) AS created_at
		`, id, title, department, requiredSkills, description, status)
		return &j, err
	}
	err := database.DB.Get(&j, `
		UPDATE jobs SET
			title       = COALESCE(NULLIF($2, ''), title),
			department  = COALESCE(NULLIF($3, ''), department),
			required_skills = COALESCE(NULLIF($4, ''), required_skills),
			description = COALESCE(NULLIF($5, ''), description),
			location    = COALESCE(NULLIF($6, ''), location),
			status      = COALESCE(NULLIF($7, ''), status)
		WHERE id::text = $1
		RETURNING id::text, title, department, COALESCE(required_skills, '') AS required_skills, description, location, status, posted_by::text, created_at
	`, id, title, department, requiredSkills, description, location, status)
	return &j, err
}

func DeleteJob(id string) error {
	_, err := database.DB.Exec(`UPDATE jobs SET status = 'closed' WHERE id::text = $1`, id)
	return err
}
