package utils

import (
	"bytes"
	"encoding/json"
	"hrms/model"
	"net/http"
)

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Errors  interface{} `json:"errors,omitempty"`
}

func RespondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(APIResponse{Success: true, Data: data}); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(APIResponse{Success: false, Message: "Failed to encode response", Errors: err.Error()})
		return
	}
	w.WriteHeader(status)
	_, _ = w.Write(buf.Bytes())
}

func RespondError(w http.ResponseWriter, status int, err error, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}
	json.NewEncoder(w).Encode(APIResponse{
		Success: false,
		Message: message,
		Errors:  errMsg,
	})
}

func RespondValidationError(w http.ResponseWriter, errs []model.ValidationError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnprocessableEntity)
	json.NewEncoder(w).Encode(APIResponse{
		Success: false,
		Message: "Validation failed",
		Errors:  errs,
	})
}
