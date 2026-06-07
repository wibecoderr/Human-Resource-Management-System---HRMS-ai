package handler

import (
	"database/sql"
	"hrms/database"
	"hrms/dbhelper"
	"hrms/model"
	"hrms/utils"
	"net/http"
	"strings"

	"github.com/jmoiron/sqlx"
)

// POST /api/auth/register
func RegisterUser(w http.ResponseWriter, r *http.Request) {
	var req model.RegisterRequest
	if err := utils.ParseBody(r.Body, &req); err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "Failed to parse request body")
		return
	}

	if errs := utils.ValidateStruct(req); errs != nil {
		utils.RespondValidationError(w, errs)
		return
	}

	// Only admin can be created via seed — block self-registration as admin
	if req.Role == "admin" {
		utils.RespondError(w, http.StatusForbidden, nil, "Cannot self-register as admin")
		return
	}

	exists, err := dbhelper.UserExist(req.Email)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to check user")
		return
	}
	if exists {
		utils.RespondError(w, http.StatusConflict, nil, "User with this email already exists")
		return
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to hash password")
		return
	}

	var (
		userID    string
		sessionID string
		jwtToken  string
	)

	err = database.Tx(func(tx *sqlx.Tx) error {
		var txErr error
		userID, txErr = dbhelper.AddEmployee(tx, req.Name, req.Email, req.Role, req.PhoneNo, hashedPassword)
		if txErr != nil {
			return txErr
		}
		sessionID, txErr = dbhelper.CreateSession(tx, userID)
		return txErr
	})
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to register user")
		return
	}

	jwtToken, err = utils.GenerateJWT(userID, sessionID, req.Role)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to generate token")
		return
	}

	// Save token to session row for logout-by-token support
	_ = dbhelper.AttachTokenToSession(sessionID, jwtToken)

	utils.RespondJSON(w, http.StatusCreated, model.AuthResponse{
		Token:  jwtToken,
		UserID: userID,
		Role:   req.Role,
	})
}

// POST /api/auth/login
func LoginUser(w http.ResponseWriter, r *http.Request) {
	var req model.LoginRequest
	if err := utils.ParseBody(r.Body, &req); err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "Failed to parse request body")
		return
	}

	if errs := utils.ValidateStruct(req); errs != nil {
		utils.RespondValidationError(w, errs)
		return
	}

	user, err := dbhelper.GetUserByEmail(req.Email)
	if err == sql.ErrNoRows {
		utils.RespondError(w, http.StatusUnauthorized, nil, "Invalid email or password")
		return
	}
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to fetch user")
		return
	}

	if !utils.CheckPassword(req.Password, user.PasswordHash) {
		utils.RespondError(w, http.StatusUnauthorized, nil, "Invalid email or password")
		return
	}
	if user.Status != "active" {
		utils.RespondError(w, http.StatusForbidden, nil, "User account is inactive")
		return
	}

	var (
		sessionID string
		jwtToken  string
	)

	err = database.Tx(func(tx *sqlx.Tx) error {
		var txErr error
		sessionID, txErr = dbhelper.CreateSession(tx, user.ID)
		return txErr
	})
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to create session")
		return
	}

	jwtToken, err = utils.GenerateJWT(user.ID, sessionID, user.Role)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to generate token")
		return
	}

	_ = dbhelper.AttachTokenToSession(sessionID, jwtToken)

	utils.RespondJSON(w, http.StatusOK, model.AuthResponse{
		Token:  jwtToken,
		UserID: user.ID,
		Role:   user.Role,
	})
}

// POST /api/auth/logout
func LogoutUser(w http.ResponseWriter, r *http.Request) {
	sessionID := r.Header.Get("X-Session-ID")
	if sessionID == "" {

		authHeader := r.Header.Get("Authorization")
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenStr != "" {
			_ = dbhelper.InvalidateSessionByToken(tokenStr)
		}
	} else {
		_ = dbhelper.InvalidateSession(sessionID)
	}

	utils.RespondJSON(w, http.StatusOK, map[string]string{"message": "Logged out successfully"})
}

// GET /api/auth/me
func GetMe(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")

	user, err := dbhelper.GetUserByID(userID)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to fetch user")
		return
	}

	utils.RespondJSON(w, http.StatusOK, user)
}
