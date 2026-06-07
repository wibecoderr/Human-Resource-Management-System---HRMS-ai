package router

import (
	"hrms/handler"
	"hrms/middleware"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func New() *gin.Engine {
	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "HRMS API Running"})
	})
	r.Use(func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin == "http://127.0.0.1:5500" || origin == "http://localhost:5500" || origin == "http://localhost:8080" || origin == "http://127.0.0.1:8080" {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		}
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		if strings.HasPrefix(c.Request.URL.Path, "/frontend") {
			c.Header("Cache-Control", "no-store, max-age=0")
		}
		c.Next()
	})
	r.Static("/frontend", "./frontend")
	r.Static("/uploads", "./uploads")
	r.GET("/app", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/frontend/index.html")
	})

	// ── Auth ──────────────────────────────────────────────────────────────────
	auth := r.Group("/api/auth")
	{
		auth.GET("/login", func(c *gin.Context) {
			c.Redirect(http.StatusFound, "/frontend/login.html")
		})
		auth.POST("/login", wrapF(handler.LoginUser))
		auth.POST("/register", protectedWithRoles(handler.RegisterUser, "admin", "hr_recruiter"))
		auth.POST("/logout", protected(handler.LogoutUser))
		auth.GET("/me", protected(handler.GetMe))
	}

	// ── Employees ─────────────────────────────────────────────────────────────
	employees := r.Group("/api/employees")
	{
		employees.GET("", protectedWithRoles(handler.GetEmployees, "admin", "senior_manager", "hr_recruiter"))
		employees.GET("/:id", protected(handler.GetEmployee))
		employees.POST("", protectedWithRoles(handler.CreateEmployee, "admin"))
		employees.PUT("/:id", protected(handler.UpdateEmployee))
		employees.DELETE("/:id", protectedWithRoles(handler.DeleteEmployee, "admin"))
	}

	// ── Attendance ────────────────────────────────────────────────────────────
	attendance := r.Group("/api/attendance")
	{
		attendance.POST("/checkin", protected(handler.CheckIn))
		attendance.POST("/checkout", protected(handler.CheckOut))
		attendance.GET("/me", protected(handler.GetMyAttendance))
		attendance.GET("/:employeeId", protectedWithRoles(handler.GetEmployeeAttendance, "admin", "senior_manager", "hr_recruiter"))
	}

	// ── Leave ─────────────────────────────────────────────────────────────────
	leave := r.Group("/api/leave")
	{
		leave.POST("", protected(handler.ApplyLeave))
		leave.GET("/me", protected(handler.GetMyLeaves))
		leave.GET("", protectedWithRoles(handler.GetAllLeaves, "admin", "senior_manager", "hr_recruiter"))
		leave.PUT("/:id/approve", protectedWithRoles(handler.UpdateLeaveStatus, "admin", "senior_manager"))
		leave.PUT("/:id/reject", protectedWithRoles(handler.RejectLeave, "admin", "senior_manager"))
	}

	// ── Payroll ───────────────────────────────────────────────────────────────
	payroll := r.Group("/api/payroll")
	{
		payroll.POST("", protectedWithRoles(handler.SetPayroll, "admin"))
		payroll.GET("/me", protected(handler.GetMyPayroll))
		payroll.GET("", protectedWithRoles(handler.GetAllPayroll, "admin", "hr_recruiter"))
		payroll.GET("/:employeeId", protectedWithRoles(handler.GetEmployeePayroll, "admin", "hr_recruiter"))
		payroll.PUT("/:id", protectedWithRoles(handler.UpdatePayroll, "admin"))
	}

	// ── Jobs ──────────────────────────────────────────────────────────────────
	jobs := r.Group("/api/jobs")
	{
		jobs.GET("", wrapF(handler.GetJobs))    // public
		jobs.GET("/:id", wrapF(handler.GetJob)) // public
		jobs.POST("", protectedWithRoles(handler.CreateJob, "admin", "hr_recruiter", "senior_manager"))
		jobs.PUT("/:id", protectedWithRoles(handler.UpdateJob, "admin", "hr_recruiter", "senior_manager"))
		jobs.DELETE("/:id", protectedWithRoles(handler.DeleteJob, "admin", "hr_recruiter", "senior_manager"))
	}

	// ── Candidates ────────────────────────────────────────────────────────────
	candidates := r.Group("/api/candidates")
	{
		candidates.POST("/apply", wrapF(handler.ApplyCandidateWithResume)) // public applicant PDF upload
		candidates.POST("", wrapF(handler.ApplyCandidate))                 // public — anyone applies
		candidates.GET("", protectedWithRoles(handler.GetCandidates, "admin", "hr_recruiter", "senior_manager"))
		candidates.GET("/:id", protectedWithRoles(handler.GetCandidate, "admin", "hr_recruiter", "senior_manager"))
		candidates.PUT("/:id", protectedWithRoles(handler.UpdateCandidateStatus, "admin", "hr_recruiter"))
		candidates.DELETE("/:id", protectedWithRoles(handler.DeleteCandidate, "admin", "hr_recruiter"))
		candidates.POST("/:id/resume", protectedWithRoles(handler.UploadCandidateResume, "admin", "hr_recruiter"))
		candidates.POST("/:id/screen", protectedWithRoles(handler.ScreenCandidateResume, "admin", "hr_recruiter"))
		candidates.GET("/:id/analysis", protectedWithRoles(handler.GetCandidateAnalysis, "admin", "hr_recruiter"))
	}

	performance := r.Group("/api/performance")
	{
		performance.POST("", protectedWithRoles(handler.CreatePerformanceReview, "admin", "hr_recruiter", "senior_manager"))
		performance.GET("/me", protected(handler.GetMyPerformanceReviews))
		performance.GET("", protectedWithRoles(handler.GetAllPerformanceReviews, "admin", "hr_recruiter", "senior_manager"))
		performance.GET("/:employeeId", protectedWithRoles(handler.GetEmployeePerformanceReviews, "admin", "hr_recruiter", "senior_manager"))
		performance.PUT("/:id", protectedWithRoles(handler.UpdatePerformanceReview, "admin", "hr_recruiter", "senior_manager"))
		performance.DELETE("/:id", protectedWithRoles(handler.DeletePerformanceReview, "admin"))
	}

	// ── Dashboard ─────────────────────────────────────────────────────────────
	dashboard := r.Group("/api/dashboard")
	{
		dashboard.GET("/admin", protectedWithRoles(handler.GetAdminDashboard, "admin"))
		dashboard.GET("/hr", protectedWithRoles(handler.GetHRDashboard, "admin", "hr_recruiter"))
		dashboard.GET("/manager", protectedWithRoles(handler.GetManagerDashboard, "admin", "senior_manager"))
		dashboard.GET("/employee", protectedWithRoles(handler.GetEmployeeDashboard, "employee", "admin", "hr_recruiter", "senior_manager"))
	}

	notifications := r.Group("/api/notifications")
	{
		notifications.GET("", protected(handler.GetNotifications))
		notifications.PUT("/:id/read", protected(handler.MarkNotificationRead))
	}

	interviews := r.Group("/api/interviews")
	{
		interviews.POST("/start", protectedWithRoles(handler.StartInterview, "admin", "hr_recruiter"))
		interviews.GET("", protectedWithRoles(handler.GetInterviews, "admin", "hr_recruiter", "senior_manager"))
		interviews.GET("/:id", protectedWithRoles(handler.GetInterview, "admin", "hr_recruiter", "senior_manager"))
		interviews.PUT("/:id/submit", protectedWithRoles(handler.SubmitInterview, "admin", "hr_recruiter"))
	}

	return r
}

func protected(fn http.HandlerFunc) gin.HandlerFunc {
	return wrapH(middleware.Authenticate(fn))
}

func protectedWithRoles(fn http.HandlerFunc, roles ...string) gin.HandlerFunc {
	return wrapH(middleware.Authenticate(middleware.RequireRole(roles...)(fn)))
}

func wrapF(fn http.HandlerFunc) gin.HandlerFunc {
	return wrapH(fn)
}

func wrapH(h http.Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		for _, param := range c.Params {
			c.Request.SetPathValue(param.Key, param.Value)
		}
		h.ServeHTTP(c.Writer, c.Request)
	}
}
