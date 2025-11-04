package config

import (
	"LegoManagerAPI/internal/config/application"
	"LegoManagerAPI/internal/config/cache"
	"LegoManagerAPI/internal/config/database"
)

// Config represents the top-level configuration structure containing database, cache, and application settings.
type Config struct {
	Database database.DatabaseConfig
	Cache    cache.CacheConfig
	App      application.ApplicationConfig
}

// Load creates and populates Config from env vars
func Load() (*Config, error) {
	cfg := &Config{
		Database: database.LoadDatabaseConfig(),
		Cache:    cache.LoadCacheConfig(),
		App:      application.LoadApplicationConfig(),
	}

	return cfg, nil
}
