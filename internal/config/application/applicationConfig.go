package application

import (
	"strings"

	"github.com/charmbracelet/log"

	"LegoManagerAPI/internal/config/configUtilities"
)

// ApplicationConfig holds the Application configuration options
type ApplicationConfig struct {
	Port            int
	ApplicationName string
	LogLVL          string
	Environment     string
}

// LoadApplicationConfig initializes and returns an ApplicationConfig struct populated with values from environment variables.
func LoadApplicationConfig() ApplicationConfig {
	return ApplicationConfig{
		Port:            configUtilities.GetEnvAsInt("PORT", 8080),
		ApplicationName: configUtilities.GetEnvAsString("APP_NAME", "Lego Manager API"),
		LogLVL:          configUtilities.GetEnvAsString("LOG_LEVEL", "info"),
		Environment:     configUtilities.GetEnvAsString("APP_ENV", "development"),
	}
}

// SetupLogger sets the global log level according to the application's configuration.
func SetupLogger(levelString string) {
	var level log.Level

	switch strings.ToLower(levelString) {
	case "debug":
		level = log.DebugLevel
	case "info":
		level = log.InfoLevel
	case "warn":
		level = log.WarnLevel
	case "error":
		level = log.ErrorLevel
	default:
	}

	log.SetLevel(level)
	log.Infof("Log level set to %s", level)
}
