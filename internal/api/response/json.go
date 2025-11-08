package response

import (
	"encoding/json"
	"net/http"

	"github.com/charmbracelet/log"
)

// JSON writes a JSON response
func JSON(res http.ResponseWriter, status int, data interface{}) {
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(status)

	if err := json.NewEncoder(res).Encode(data); err != nil {
		log.Error("Failed to encode JSON response", "error", err)
	}
}

// Error writes an error JSON response
func Error(res http.ResponseWriter, status int, message string) {
	JSON(res, status, map[string]string{
		"error": message,
	})
}
