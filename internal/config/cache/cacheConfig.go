package cache

import (
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
		Host:     configUtilities.GetEnvAsString("REDIS_HOST", "localhost"),
		Port:     configUtilities.GetEnvAsInt("REDIS_PORT", 6379),
		Password: configUtilities.GetEnvAsString("REDIS_PASSWORD", "password"),
		DB:       configUtilities.GetEnvAsInt("REDIS_DB", 1),
	}
}
