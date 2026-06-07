package model

import (
	"encoding/json"
	"time"
)

// ─── Auth ────────────────────────────────────────────────────────────────────

type RegisterRequest struct {
	Name     string `json:"name"     validate:"required"`
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	Role     string `json:"role"     validate:"required,oneof=employee hr_recruiter senior_manager"`
	PhoneNo  string `json:"phone_no" validate:"required"`
}

type LoginRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type AuthResponse struct {
	Token  string `json:"token"`
	UserID string `json:"user_id"`
	Role   string `json:"role"`
}

// ─── User / Employee ─────────────────────────────────────────────────────────

type User struct {
	ID        string    `db:"id"         json:"id"`
	Name      string    `db:"name"       json:"name"`
	Email     string    `db:"email"      json:"email"`
	Role      string    `db:"role"       json:"role"`
	PhoneNo   string    `db:"phone_no"   json:"phone_no"`
	Status    string    `db:"status"     json:"status"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type CreateEmployeeRequest struct {
	Name     string `json:"name"     validate:"required"`
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	Role     string `json:"role"     validate:"required,oneof=employee hr_recruiter senior_manager admin"`
	PhoneNo  string `json:"phone_no" validate:"required"`
}

type UpdateEmployeeRequest struct {
	Name    string `json:"name"`
	PhoneNo string `json:"phone_no"`
	Role    string `json:"role"`
	Status  string `json:"status"`
}

// ─── Attendance ───────────────────────────────────────────────────────────────

type Attendance struct {
	ID         string     `db:"id"           json:"id"`
	EmployeeID string     `db:"employee_id"  json:"employee_id"`
	CheckIn    time.Time  `db:"check_in"     json:"check_in"`
	CheckOut   *time.Time `db:"check_out"    json:"check_out"`
	Date       time.Time  `db:"date"         json:"date"`
	CreatedAt  time.Time  `db:"created_at"   json:"created_at"`
}

// ─── Leave ────────────────────────────────────────────────────────────────────

type Leave struct {
	ID         string    `db:"id"          json:"id"`
	EmployeeID string    `db:"employee_id" json:"employee_id"`
	LeaveType  string    `db:"leave_type"  json:"leave_type"`
	StartDate  time.Time `db:"start_date"  json:"start_date"`
	EndDate    time.Time `db:"end_date"    json:"end_date"`
	Reason     string    `db:"reason"      json:"reason"`
	Status     string    `db:"status"      json:"status"`
	ApprovedBy *string   `db:"approved_by" json:"approved_by"`
	CreatedAt  time.Time `db:"created_at"  json:"created_at"`
}

type ApplyLeaveRequest struct {
	LeaveType string `json:"leave_type" validate:"required,oneof=sick casual earned"`
	StartDate string `json:"start_date" validate:"required"`
	EndDate   string `json:"end_date"   validate:"required"`
	Reason    string `json:"reason"     validate:"required"`
}

type ApproveLeaveRequest struct {
	Status string `json:"status" validate:"required,oneof=approved rejected"`
}

// ─── Dashboard ────────────────────────────────────────────────────────────────

type AdminDashboard struct {
	TotalEmployees int `json:"employees"`
	OpenJobs       int `json:"openJobs"`
	PendingLeaves  int `json:"pendingLeaves"`
	Candidates     int `json:"candidates"`
}

type HRDashboard struct {
	OpenJobs          int `json:"openJobs"`
	Candidates        int `json:"candidates"`
	PendingInterviews int `json:"pendingInterviews"`
}

type ManagerDashboard struct {
	TeamSize           int `json:"teamSize"`
	PendingLeaves      int `json:"pendingLeaves"`
	PerformanceReviews int `json:"performanceReviews"`
}

type EmployeeDashboard struct {
	AttendanceSummary int      `json:"attendanceSummary"`
	LeaveBalance      int      `json:"leaveBalance"`
	RecentPayroll     *Payroll `json:"recentPayroll"`
}

// Notifications

type Notification struct {
	ID        string    `db:"id" json:"id"`
	UserID    string    `db:"user_id" json:"user_id"`
	Title     string    `db:"title" json:"title"`
	Message   string    `db:"message" json:"message"`
	IsRead    bool      `db:"is_read" json:"is_read"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type NotificationsResponse struct {
	UnreadCount   int            `json:"unread_count"`
	Notifications []Notification `json:"notifications"`
}

// ─── Validation errors ────────────────────────────────────────────────────────

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ─── Payroll ──────────────────────────────────────────────────────────────────

type Payroll struct {
	ID           string    `db:"id"            json:"id"`
	EmployeeID   string    `db:"employee_id"   json:"employee_id"`
	EmployeeName string    `db:"employee_name" json:"employee_name"`
	Month        int       `db:"month"         json:"month"`
	Year         int       `db:"year"          json:"year"`
	BasicSalary  float64   `db:"basic_salary"  json:"basic_salary"`
	CreatedAt    time.Time `db:"created_at"    json:"created_at"`
}

type SetPayrollRequest struct {
	EmployeeID  string  `json:"employee_id"  validate:"required"`
	Month       int     `json:"month"        validate:"required"`
	Year        int     `json:"year"         validate:"required"`
	BasicSalary float64 `json:"basic_salary" validate:"required"`
}

type UpdatePayrollRequest struct {
	BasicSalary float64 `json:"basic_salary" validate:"required"`
}

// ─── Jobs ─────────────────────────────────────────────────────────────────────

type Job struct {
	ID             string    `db:"id"              json:"id"`
	Title          string    `db:"title"           json:"title"`
	Department     string    `db:"department"      json:"department"`
	RequiredSkills string    `db:"required_skills" json:"required_skills"`
	Description    string    `db:"description"     json:"description"`
	Location       string    `db:"location"        json:"location"`
	Status         string    `db:"status"          json:"status"`
	PostedBy       string    `db:"posted_by"       json:"posted_by"`
	CreatedAt      time.Time `db:"created_at"      json:"created_at"`
}

type CreateJobRequest struct {
	Title          string `json:"title"           validate:"required"`
	Department     string `json:"department"      validate:"required"`
	RequiredSkills string `json:"required_skills"`
	Description    string `json:"description"     validate:"required"`
	Location       string `json:"location"        validate:"required"`
}

type UpdateJobRequest struct {
	Title          string `json:"title"`
	Department     string `json:"department"`
	RequiredSkills string `json:"required_skills"`
	Description    string `json:"description"`
	Location       string `json:"location"`
	Status         string `json:"status"`
}

// ─── Candidates ───────────────────────────────────────────────────────────────

type Candidate struct {
	ID          string    `db:"id"           json:"id"`
	Name        string    `db:"name"         json:"name"`
	Email       string    `db:"email"        json:"email"`
	Phone       string    `db:"phone"        json:"phone"`
	JobID       string    `db:"job_id"       json:"job_id"`
	JobTitle    string    `db:"job_title"    json:"job_title"`
	ResumeURL   string    `db:"resume_url"   json:"resume_url"`
	ResumeText  string    `db:"resume_text"  json:"resume_text,omitempty"`
	AIScore     *float64  `db:"ai_score"     json:"ai_score,omitempty"`
	AIReport    JSONB     `db:"ai_report"    json:"ai_report,omitempty"`
	CoverLetter string    `db:"cover_letter" json:"cover_letter"`
	Status      string    `db:"status"       json:"status"`
	CreatedAt   time.Time `db:"created_at"   json:"created_at"`
}

type ApplyCandidateRequest struct {
	Name        string `json:"name"       validate:"required"`
	Email       string `json:"email"      validate:"required,email"`
	Phone       string `json:"phone"      validate:"required"`
	JobID       string `json:"job_id"     validate:"required"`
	ResumeURL   string `json:"resume_url" validate:"required"`
	CoverLetter string `json:"cover_letter"`
}

type UpdateCandidateStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=applied shortlisted interviewing interviewed offered rejected"`
}

type CandidateAnalysisResponse struct {
	CandidateID string          `json:"candidate_id"`
	Score       float64         `json:"score"`
	Report      json.RawMessage `json:"report"`
}

type InterviewQuestion struct {
	ID       int    `json:"id"`
	Question string `json:"question"`
}

type InterviewAnswer struct {
	QuestionID int    `json:"question_id"`
	Question   string `json:"question"`
	Answer     string `json:"answer"`
}

type Interview struct {
	ID                  string     `db:"id" json:"id"`
	CandidateID         string     `db:"candidate_id" json:"candidate_id"`
	CandidateName       string     `db:"candidate_name" json:"candidate_name"`
	JobID               string     `db:"job_id" json:"job_id"`
	JobTitle            string     `db:"job_title" json:"job_title"`
	Questions           JSONB      `db:"questions" json:"questions"`
	Answers             JSONB      `db:"answers" json:"answers"`
	TechnicalScore      *float64   `db:"technical_score" json:"technical_score,omitempty"`
	CommunicationScore  *float64   `db:"communication_score" json:"communication_score,omitempty"`
	ProblemSolvingScore *float64   `db:"problem_solving_score" json:"problem_solving_score,omitempty"`
	OverallScore        *float64   `db:"overall_score" json:"overall_score,omitempty"`
	Recommendation      string     `db:"recommendation" json:"recommendation"`
	Status              string     `db:"status" json:"status"`
	AIReport            JSONB      `db:"ai_report" json:"ai_report"`
	CreatedAt           time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt           *time.Time `db:"updated_at" json:"updated_at,omitempty"`
}

type StartInterviewRequest struct {
	CandidateID string `json:"candidate_id" validate:"required"`
}

type SubmitInterviewRequest struct {
	Answers []InterviewAnswer `json:"answers" validate:"required"`
}

type InterviewEvaluation struct {
	TechnicalScore      float64  `json:"technical_score"`
	CommunicationScore  float64  `json:"communication_score"`
	ProblemSolvingScore float64  `json:"problem_solving_score"`
	OverallScore        float64  `json:"overall_score"`
	Recommendation      string   `json:"recommendation"`
	Strengths           []string `json:"strengths"`
	Risks               []string `json:"risks"`
	Summary             string   `json:"summary"`
}

type JSONB json.RawMessage

func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = JSONB([]byte("{}"))
		return nil
	}
	switch v := value.(type) {
	case []byte:
		*j = JSONB(v)
	case string:
		*j = JSONB([]byte(v))
	default:
		*j = JSONB([]byte("{}"))
	}
	return nil
}

func (j JSONB) MarshalJSON() ([]byte, error) {
	if len(j) == 0 {
		return []byte("{}"), nil
	}
	raw := json.RawMessage(j)
	if !json.Valid(raw) {
		return []byte("{}"), nil
	}
	return raw.MarshalJSON()
}

// Performance

type PerformanceReview struct {
	ID           string    `db:"id" json:"id"`
	EmployeeID   string    `db:"employee_id" json:"employee_id"`
	EmployeeName string    `db:"employee_name" json:"employee_name"`
	ReviewerID   string    `db:"reviewer_id" json:"reviewer_id"`
	ReviewerName string    `db:"reviewer_name" json:"reviewer_name"`
	ReviewPeriod string    `db:"review_period" json:"review_period"`
	Rating       int       `db:"rating" json:"rating"`
	Feedback     string    `db:"feedback" json:"feedback"`
	Goals        string    `db:"goals" json:"goals"`
	Status       string    `db:"status" json:"status"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
}

type CreatePerformanceReviewRequest struct {
	EmployeeID   string `json:"employee_id" validate:"required"`
	ReviewPeriod string `json:"review_period" validate:"required"`
	Rating       int    `json:"rating"`
	Feedback     string `json:"feedback" validate:"required"`
	Goals        string `json:"goals"`
	Status       string `json:"status"`
}

type UpdatePerformanceReviewRequest struct {
	ReviewPeriod string `json:"review_period"`
	Rating       int    `json:"rating"`
	Feedback     string `json:"feedback"`
	Goals        string `json:"goals"`
	Status       string `json:"status"`
}
