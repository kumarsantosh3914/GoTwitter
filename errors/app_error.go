package apperrors

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// AppError defines a structured error for the application
type AppError struct {
	Message    string `json:"message"`
	StatusCode int    `json:"status_code"`
	Err        error  `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func NewAppError(message string, statusCode int, err error) *AppError {
	return &AppError{
		Message:    message,
		StatusCode: statusCode,
		Err:        err,
	}
}

// WriteError writes the error response in JSON format
func WriteError(w http.ResponseWriter, appErr *AppError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(appErr.StatusCode)

	response := map[string]any{
		"status":  "error",
		"message": appErr.Message,
	}

	// In a real application, you might want to hide internal errors in production
	// but for now, we'll include it if it exists.
	if appErr.Err != nil {
		response["error"] = appErr.Err.Error()
	}

	_ = json.NewEncoder(w).Encode(response)
}
