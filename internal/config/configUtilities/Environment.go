package configUtilities

import (
	"os"
	"strconv"

	"github.com/charmbracelet/log"
)

// GetEnvAsInt retrieves the environment variable value by key and converts it to an int, returning the defaultValue if unset or invalid.
func GetEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)

	if valueStr == "" {
		log.Warn("Environment variable " + key + " is not set. Using default value: " + strconv.Itoa(defaultValue))
		return defaultValue
	}
	if valueStr != "" {
		if value, err := strconv.Atoi(valueStr); err == nil {
			return value
		}

	}
	return defaultValue
}

// GetEnvAsString retrieves the environment variable value by key and returns it only if it is a string, returning the defaultValue if unset or invalid.
func GetEnvAsString(key string, defaultValue string) string {
	if valueStr := os.Getenv(key); valueStr == "" {
		log.Warn("Environment variable " + key + " is not set. Using default value: " + defaultValue)
		return defaultValue
	}

	if valueStr := os.Getenv(key); valueStr != "" {
		return valueStr
	}
	return defaultValue
}
