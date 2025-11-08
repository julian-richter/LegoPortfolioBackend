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
		Host:     configUtilities.GetEnvAsString("POSTGRES_HOST", "localhost"),
		Port:     configUtilities.GetEnvAsInt("POSTGRES_PORT", 5432),
		User:     configUtilities.GetEnvAsString("POSTGRES_USER", "legouser"),
		Password: configUtilities.GetEnvAsString("POSTGRES_PASSWORD", "legopas"),
		DBName:   configUtilities.GetEnvAsString("POSTGRES_DB", "lego_collection"),
		SSLMode:  configUtilities.GetEnvAsString("POSTGRES_SSL_MODE", "disable"),
		MaxConns: configUtilities.GetEnvAsInt("POSTGRES_MAX_CONNS", 100),
		MinConns: configUtilities.GetEnvAsInt("POSTGRES_MIN_CONNS", 1),
	}
}
