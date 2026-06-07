# HRMS AI

A Go + PostgreSQL Human Resource Management System with JWT authentication, employee management, attendance, leave, payroll, jobs, candidates, performance reviews, dashboards, notifications, AI resume screening, and AI interview assessment.

## Tech Stack

- Go
- Gin / `net/http` handler style
- PostgreSQL
- JWT authentication
- Google Gemini API
- HTML
- CSS
- JavaScript

## Main Features

- Authentication with JWT and session validation
- Role based access for admin, HR recruiter, senior manager, and employee
- Employee CRUD
- Attendance check-in/check-out
- Leave apply, approve, and reject
- Payroll create and update
- Jobs create, update, close, and public listing
- Public candidate job application with PDF resume upload
- Candidate management and status tracking
- AI resume screening with extracted PDF text
- AI interview question generation and assessment
- Interview history
- Admin, HR, manager, and employee dashboards
- Notifications
- Performance reviews

## Project Structure

```text
HRMS/
  main.go                  Application entry point
  database/                Database connection and additive schema preparation
  dbhelper/                Database helper functions
  handler/                 HTTP handlers
  middleware/              JWT and role middleware
  model/                   Request/response/database models
  router/                  Route registration
  services/                Gemini and PDF extraction services
  utils/                   JWT, response, parsing, hashing, validation utilities
  frontend/                HTML, CSS, and JavaScript frontend
```

## Environment Setup

Create a `.env` file from `.env.example`:

```env
DB_HOST=localhost
DB_PORT=5432
DB_NAME=hrms
DB_USER=postgres
DB_PASSWORD=change-me
DB_SSLMODE=disable
PORT=8080
GEMINI_API_KEY=your-gemini-api-key
```

Do not commit `.env`. It is ignored by git.

## Run Locally

```bash
go mod download
go run main.go
```

Open:

```text
http://localhost:8080/frontend/login.html
```

Default seeded admin:

```text
Email: admin@hrms.com
Password: admin123
```

## Build And Test

```bash
go test ./...
go build ./...
```

## Application Flow

### 1. Authentication Flow

1. User opens `frontend/login.html`.
2. Login form calls `POST /api/auth/login`.
3. Backend validates email and password.
4. Backend creates a session row and returns a JWT.
5. Frontend stores token in `localStorage`.
6. Every protected page calls `GET /api/auth/me` with `Authorization: Bearer <token>`.
7. Middleware validates JWT and active session before allowing protected routes.

### 2. Employee Flow

1. Admin opens Employees page.
2. Frontend calls `GET /api/employees`.
3. Admin creates employee with `POST /api/employees`.
4. Backend creates user and employee-compatible data.
5. Table refreshes after create/update/delete.

### 3. Jobs And Candidate Application Flow

1. HR/admin creates a job from Jobs page using `POST /api/jobs`.
2. Open job appears in the Jobs table.
3. Public applicant clicks Apply or opens `frontend/apply.html`.
4. Applicant selects job, enters details, and uploads PDF resume.
5. Frontend sends multipart request to `POST /api/candidates/apply`.
6. Backend validates PDF, stores it under `uploads/resumes`, extracts text, and creates candidate.
7. HR can view candidate in Candidates page.
8. Candidate appears with AI Score as `Pending` until screening is run.

### 4. AI Resume Screening Flow

1. HR opens Resume Screening page.
2. Frontend calls `GET /api/candidates`.
3. HR selects candidate.
4. HR clicks Screen Candidate.
5. Frontend calls `POST /api/candidates/{id}/screen`.
6. Backend fetches candidate, resume text, and linked job description.
7. Backend sends job title, required skills, job description, and resume text to Gemini.
8. Gemini returns structured JSON with score, skills, strengths, weaknesses, and recommendation.
9. Backend saves `ai_score`, `ai_report`, and updates candidate status:
   - `score >= 80`: shortlisted
   - `score >= 60`: interviewing
   - otherwise: rejected
10. Candidates page displays score badge.

### 5. AI Interview Flow

1. HR opens AI Interview page.
2. Frontend loads candidates using `GET /api/candidates`.
3. HR selects candidate and starts interview.
4. Frontend calls `POST /api/interviews/start`.
5. Backend fetches candidate, job, and resume analysis.
6. Gemini generates interview questions.
7. If Gemini quota/API fails, backend creates fallback job-specific interview questions so the workflow still works.
8. Candidate answers questions in the browser.
9. Optional voice mode can speak questions and capture spoken answers when supported by the browser.
10. HR submits answers using `PUT /api/interviews/{id}/submit`.
11. Gemini evaluates answers and returns technical, communication, problem-solving, overall score, and recommendation.
12. If Gemini evaluation fails, backend creates a fallback assessment and marks it clearly in the report.
13. Results are saved in the `interviews` table.
14. Interview History page displays saved results.

### 6. Leave Flow

1. Employee applies for leave using `POST /api/leave`.
2. Manager/admin views pending leaves using `GET /api/leave`.
3. Manager/admin approves using `PUT /api/leave/{id}/approve`.
4. Manager/admin rejects using `PUT /api/leave/{id}/reject`.
5. UI refreshes after status update.

### 7. Payroll Flow

1. Admin creates payroll using `POST /api/payroll`.
2. Admin updates salary using `PUT /api/payroll/{id}`.
3. Payroll table refreshes after update.

## Important API Routes

```text
POST   /api/auth/login
GET    /api/auth/me
POST   /api/auth/logout

GET    /api/employees
POST   /api/employees
PUT    /api/employees/{id}
DELETE /api/employees/{id}

GET    /api/jobs
POST   /api/jobs
PUT    /api/jobs/{id}
DELETE /api/jobs/{id}

POST   /api/candidates/apply
GET    /api/candidates
GET    /api/candidates/{id}
PUT    /api/candidates/{id}
DELETE /api/candidates/{id}
POST   /api/candidates/{id}/resume
POST   /api/candidates/{id}/screen
GET    /api/candidates/{id}/analysis

POST   /api/interviews/start
GET    /api/interviews
GET    /api/interviews/{id}
PUT    /api/interviews/{id}/submit

GET    /api/dashboard/admin
GET    /api/dashboard/hr
GET    /api/dashboard/manager
GET    /api/dashboard/employee
```

## Security Notes

- `.env`, uploaded resumes, PDFs, executables, and local build caches are ignored.
- JWT is required for protected APIs.
- Role middleware protects admin/HR/manager/employee routes.
- Public candidate application only accepts PDF files.
- Gemini API key must be configured through environment variables.

## GitHub Publishing Notes

Before pushing, verify:

```bash
git status
git ls-files | grep -E "\.env$|uploads/|\.pdf$|\.exe$"
```

The second command should return nothing.