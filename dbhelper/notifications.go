package dbhelper

import (
	"hrms/database"
	"hrms/model"
)

func GetNotifications(userID string) ([]model.Notification, error) {
	var notifications []model.Notification
	err := database.DB.Select(&notifications, `
		SELECT id::text, user_id::text, COALESCE(title, '') AS title, COALESCE(message, '') AS message,
		       COALESCE(is_read, FALSE) AS is_read, COALESCE(created_at, NOW()) AS created_at
		FROM notifications
		WHERE user_id::text = $1
		ORDER BY created_at DESC
	`, userID)
	return notifications, err
}

func CountUnreadNotifications(userID string) (int, error) {
	var count int
	err := database.DB.Get(&count, `
		SELECT COUNT(*)
		FROM notifications
		WHERE user_id::text = $1 AND COALESCE(is_read, FALSE) = FALSE
	`, userID)
	return count, err
}

func MarkNotificationRead(notificationID, userID string) (*model.Notification, error) {
	var notification model.Notification
	err := database.DB.Get(&notification, `
		UPDATE notifications
		SET is_read = TRUE
		WHERE id::text = $1 AND user_id::text = $2
		RETURNING id::text, user_id::text, COALESCE(title, '') AS title, COALESCE(message, '') AS message,
		          COALESCE(is_read, FALSE) AS is_read, COALESCE(created_at, NOW()) AS created_at
	`, notificationID, userID)
	if err != nil {
		return nil, err
	}
	return &notification, nil
}

func CreateNotification(userID, title, message string) (*model.Notification, error) {
	var notification model.Notification
	err := database.DB.Get(&notification, `
		INSERT INTO notifications (user_id, title, message)
		VALUES ($1, $2, $3)
		RETURNING id::text, user_id::text, COALESCE(title, '') AS title, COALESCE(message, '') AS message,
		          COALESCE(is_read, FALSE) AS is_read, COALESCE(created_at, NOW()) AS created_at
	`, userID, title, message)
	if err != nil {
		return nil, err
	}
	return &notification, nil
}
