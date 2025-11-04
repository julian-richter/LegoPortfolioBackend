package application

import (
	"LegoManagerAPI/internal/config/configUtilities"
)

// ApplicationConfig holds the Application configuration options
type ApplicationConfig struct {
	Port            int
	ApplicationName string
}

// LoadApplicationConfig initializes and returns an ApplicationConfig struct populated with values from environment variables.
func LoadApplicationConfig() ApplicationConfig {
	return ApplicationConfig{
		Port:            configUtilities.GetEnvAsInt("PORT", 8080),
		ApplicationName: configUtilities.GetEnvAsString("APP_NAME", "Lego Manager API"),
	}
}
