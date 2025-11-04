package database

import (
	"LegoManagerAPI/internal/config/configUtilities"
)

// DatabaseConfig type for database configuration
type DatabaseConfig struct {
	Host     string
	User     string
	Password string
	DBName   string
	SSLMode  string
	Port     int
	MaxConns int
	MinConns int
}

// LoadDatabaseConfig initializes and returns a DatabaseConfig struct populated with values from environment variables.
func LoadDatabaseConfig() DatabaseConfig {
	return DatabaseConfig{
		Host:     configUtilities.GetEnvAsString("DB_HOST", "localhost"),
		Port:     configUtilities.GetEnvAsInt("DB_PORT", 5432),
		User:     configUtilities.GetEnvAsString("DB_USER", "legouser"),
		Password: configUtilities.GetEnvAsString("DB_PASSWORD", "legopas"),
		DBName:   configUtilities.GetEnvAsString("DB_NAME", "lego_collection"),
		SSLMode:  configUtilities.GetEnvAsString("DB_SSL_MODE", "disable"),
		MaxConns: configUtilities.GetEnvAsInt("DB_MAX_CONNS", 100),
		MinConns: configUtilities.GetEnvAsInt("DB_MIN_CONNS", 1),
	}
}
