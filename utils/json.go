package utils

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-playground/validator/v10"
)

var Validator *validator.Validate

func init() {
	log.Println("[INFO] Initializing utils package")
	Validator = NewValidator()
}

func NewValidator() *validator.Validate {
	return validator.New(validator.WithRequiredStructEnabled())
}

func WriteJsonResponse(w http.ResponseWriter, status int, data any) error {
	// Set the content type to application/json
	w.Header().Set("Content-Type", "application/json")

	// Set the HTTP status code
	w.WriteHeader(status)

	// Encode the data as JSON and write it to the response
	return json.NewEncoder(w).Encode(data)
}

func WriteJsonSuccessResponse(w http.ResponseWriter, status int, message string, data any) error {
	response := map[string]any{}

	response["status"] = "success"
	response["message"] = message
	response["data"] = data
	return WriteJsonResponse(w, status, response)
}

func WriteJsonErrorResponse(w http.ResponseWriter, status int, message string, err error) error {
	response := map[string]any{}

	response["status"] = "error"
	response["message"] = message
	response["error"] = err.Error()

	return WriteJsonResponse(w, status, response)
}

func ReadJsonBody(r *http.Request, result any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields() // Prevent unkown fields from begin include in the JSON body
	return decoder.Decode(result)
}
