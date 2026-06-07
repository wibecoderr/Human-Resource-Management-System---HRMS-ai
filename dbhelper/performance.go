package dbhelper

import (
	"hrms/database"
	"hrms/model"
)

func CreatePerformanceReview(employeeID, reviewerID, reviewPeriod string, rating int, feedback, goals, status string) (*model.PerformanceReview, error) {
	if status == "" {
		status = "draft"
	}
	dbEmployeeID, err := employeeRecordID(employeeID)
	if err != nil {
		return nil, err
	}
	dbReviewerID, err := employeeRecordID(reviewerID)
	if err != nil {
		return nil, err
	}

	var review model.PerformanceReview
	err = database.DB.Get(&review, `
		INSERT INTO performance_reviews
			(employee_id, reviewer_id, review_period, rating, feedback, goals, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING
			id::text,
			COALESCE((SELECT e.user_id::text FROM employees e WHERE e.id = employee_id), employee_id::text) AS employee_id,
			COALESCE((SELECT e.name FROM employees e WHERE e.id = employee_id), employee_id::text) AS employee_name,
			COALESCE((SELECT e.user_id::text FROM employees e WHERE e.id = reviewer_id), reviewer_id::text) AS reviewer_id,
			COALESCE((SELECT e.name FROM employees e WHERE e.id = reviewer_id), reviewer_id::text) AS reviewer_name,
			review_period,
			COALESCE(rating, 0)::int AS rating,
			feedback,
			COALESCE(goals, '') AS goals,
			COALESCE(status, 'draft') AS status,
			COALESCE(created_at, NOW()) AS created_at,
			COALESCE(updated_at, created_at, NOW()) AS updated_at
	`, dbEmployeeID, dbReviewerID, reviewPeriod, rating, feedback, goals, status)
	return &review, err
}

func GetAllPerformanceReviews() ([]model.PerformanceReview, error) {
	var reviews []model.PerformanceReview
	err := database.DB.Select(&reviews, `
		SELECT
			pr.id::text,
			COALESCE(e.user_id::text, pr.employee_id::text) AS employee_id,
			COALESCE(e.name, pr.employee_id::text) AS employee_name,
			COALESCE(r.user_id::text, pr.reviewer_id::text) AS reviewer_id,
			COALESCE(r.name, pr.reviewer_id::text) AS reviewer_name,
			pr.review_period,
			COALESCE(pr.rating, 0)::int AS rating,
			pr.feedback,
			COALESCE(pr.goals, '') AS goals,
			COALESCE(pr.status, 'draft') AS status,
			COALESCE(pr.created_at, NOW()) AS created_at,
			COALESCE(pr.updated_at, pr.created_at, NOW()) AS updated_at
		FROM performance_reviews pr
		LEFT JOIN employees e ON e.id = pr.employee_id
		LEFT JOIN employees r ON r.id = pr.reviewer_id
		ORDER BY pr.created_at DESC
	`)
	return reviews, err
}

func GetPerformanceReviewsByEmployee(employeeID string) ([]model.PerformanceReview, error) {
	var reviews []model.PerformanceReview
	err := database.DB.Select(&reviews, `
		SELECT
			pr.id::text,
			COALESCE(e.user_id::text, pr.employee_id::text) AS employee_id,
			COALESCE(e.name, pr.employee_id::text) AS employee_name,
			COALESCE(r.user_id::text, pr.reviewer_id::text) AS reviewer_id,
			COALESCE(r.name, pr.reviewer_id::text) AS reviewer_name,
			pr.review_period,
			COALESCE(pr.rating, 0)::int AS rating,
			pr.feedback,
			COALESCE(pr.goals, '') AS goals,
			COALESCE(pr.status, 'draft') AS status,
			COALESCE(pr.created_at, NOW()) AS created_at,
			COALESCE(pr.updated_at, pr.created_at, NOW()) AS updated_at
		FROM performance_reviews pr
		LEFT JOIN employees e ON e.id = pr.employee_id
		LEFT JOIN employees r ON r.id = pr.reviewer_id
		WHERE pr.employee_id::text = $1 OR e.user_id::text = $1
		ORDER BY pr.created_at DESC
	`, employeeID)
	return reviews, err
}

func UpdatePerformanceReview(id, reviewPeriod string, rating int, feedback, goals, status string) (*model.PerformanceReview, error) {
	var review model.PerformanceReview
	err := database.DB.Get(&review, `
		UPDATE performance_reviews
		SET
			review_period = COALESCE(NULLIF($2, ''), review_period),
			rating = CASE WHEN $3 > 0 THEN $3 ELSE rating END,
			feedback = COALESCE(NULLIF($4, ''), feedback),
			goals = COALESCE(NULLIF($5, ''), goals),
			status = COALESCE(NULLIF($6, ''), status),
			updated_at = NOW()
		WHERE id::text = $1
		RETURNING
			id::text,
			COALESCE((SELECT e.user_id::text FROM employees e WHERE e.id = employee_id), employee_id::text) AS employee_id,
			COALESCE((SELECT e.name FROM employees e WHERE e.id = employee_id), employee_id::text) AS employee_name,
			COALESCE((SELECT e.user_id::text FROM employees e WHERE e.id = reviewer_id), reviewer_id::text) AS reviewer_id,
			COALESCE((SELECT e.name FROM employees e WHERE e.id = reviewer_id), reviewer_id::text) AS reviewer_name,
			review_period,
			COALESCE(rating, 0)::int AS rating,
			feedback,
			COALESCE(goals, '') AS goals,
			COALESCE(status, 'draft') AS status,
			COALESCE(created_at, NOW()) AS created_at,
			COALESCE(updated_at, created_at, NOW()) AS updated_at
	`, id, reviewPeriod, rating, feedback, goals, status)
	return &review, err
}

func DeletePerformanceReview(id string) error {
	_, err := database.DB.Exec(`DELETE FROM performance_reviews WHERE id::text = $1`, id)
	return err
}
