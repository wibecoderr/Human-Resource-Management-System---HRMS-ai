package handler

import (
	"database/sql"
	"hrms/dbhelper"
	"hrms/model"
	"hrms/utils"
	"net/http"
)

// GET /api/notifications
func GetNotifications(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")

	notifications, err := dbhelper.GetNotifications(userID)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to fetch notifications")
		return
	}

	unreadCount, err := dbhelper.CountUnreadNotifications(userID)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to fetch unread count")
		return
	}

	utils.RespondJSON(w, http.StatusOK, model.NotificationsResponse{
		UnreadCount:   unreadCount,
		Notifications: notifications,
	})
}

// PUT /api/notifications/{id}/read
func MarkNotificationRead(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	notificationID := r.PathValue("id")

	notification, err := dbhelper.MarkNotificationRead(notificationID, userID)
	if err == sql.ErrNoRows {
		utils.RespondError(w, http.StatusNotFound, nil, "Notification not found")
		return
	}
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to mark notification as read")
		return
	}

	utils.RespondJSON(w, http.StatusOK, notification)
}
