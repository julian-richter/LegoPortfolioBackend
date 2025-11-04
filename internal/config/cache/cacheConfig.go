package cache

import (
	"os"

	"LegoManagerAPI/internal/config/configUtilities"
)

// CacheConfig holds the configuration settings for connecting to a caching service like Redis.
type CacheConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

// LoadCacheConfig initializes and returns a CacheConfig struct populated with values from environment variables.
func LoadCacheConfig() CacheConfig {
	return CacheConfig{
		Host:     configUtilities.GetEnvAsString(os.Getenv("REDIS_HOST"), "localhost"),
		Port:     configUtilities.GetEnvAsInt(os.Getenv("REDIS_PORT"), 6379),
		Password: configUtilities.GetEnvAsString(os.Getenv("REDIS_PASSWORD"), "password"),
		DB:       configUtilities.GetEnvAsInt(os.Getenv("REDIS_DB"), 1),
	}
}
