package handler

import (
	"hrms/dbhelper"
	"hrms/model"
	"hrms/utils"
	"net/http"
)

// GET /api/dashboard/admin
func GetAdminDashboard(w http.ResponseWriter, r *http.Request) {
	employees, err := dbhelper.CountActiveEmployees()
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to fetch employee count")
		return
	}

	openJobs, _ := dbhelper.CountOpenJobs()
	pendingLeaves, _ := dbhelper.CountPendingLeaves()
	candidates, _ := dbhelper.CountCandidates()

	utils.RespondJSON(w, http.StatusOK, model.AdminDashboard{
		TotalEmployees: employees,
		OpenJobs:       openJobs,
		PendingLeaves:  pendingLeaves,
		Candidates:     candidates,
	})
}

// GET /api/dashboard/hr
func GetHRDashboard(w http.ResponseWriter, r *http.Request) {
	openJobs, _ := dbhelper.CountOpenJobs()
	candidates, _ := dbhelper.CountCandidates()
	pendingInterviews, _ := dbhelper.CountPendingInterviews()

	utils.RespondJSON(w, http.StatusOK, model.HRDashboard{
		OpenJobs:          openJobs,
		Candidates:        candidates,
		PendingInterviews: pendingInterviews,
	})
}

// GET /api/dashboard/manager
func GetManagerDashboard(w http.ResponseWriter, r *http.Request) {
	teamSize, err := dbhelper.CountActiveEmployees()
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to fetch team size")
		return
	}
	pendingLeaves, _ := dbhelper.CountPendingLeaves()
	reviews, _ := dbhelper.CountPerformanceReviews()

	utils.RespondJSON(w, http.StatusOK, model.ManagerDashboard{
		TeamSize:           teamSize,
		PendingLeaves:      pendingLeaves,
		PerformanceReviews: reviews,
	})
}

// GET /api/dashboard/employee
func GetEmployeeDashboard(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")

	attendanceSummary, err := dbhelper.CountEmployeeAttendance(userID)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to fetch attendance summary")
		return
	}
	leaveBalance, err := dbhelper.GetLeaveBalance(userID)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to fetch leave balance")
		return
	}
	recentPayroll, err := dbhelper.GetRecentPayroll(userID)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to fetch recent payroll")
		return
	}

	utils.RespondJSON(w, http.StatusOK, model.EmployeeDashboard{
		AttendanceSummary: attendanceSummary,
		LeaveBalance:      leaveBalance,
		RecentPayroll:     recentPayroll,
	})
}
