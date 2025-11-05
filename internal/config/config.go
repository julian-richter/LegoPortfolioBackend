package config

import (
	"LegoManagerAPI/internal/config/application"
	"LegoManagerAPI/internal/config/bricklink"
	"LegoManagerAPI/internal/config/cache"
	"LegoManagerAPI/internal/config/database"
)

// Config represents the top-level configuration structure containing database, cache, and application settings.
type Config struct {
	Database  database.DatabaseConfig
	Cache     cache.CacheConfig
	App       application.ApplicationConfig
	Bricklink bricklink.BricklinkConfig
}

// Load creates and populates Config from env vars
func Load() (*Config, error) {
	cfg := &Config{
		Database:  database.LoadDatabaseConfig(),
		Cache:     cache.LoadCacheConfig(),
		App:       application.LoadApplicationConfig(),
		Bricklink: bricklink.LoadBricklinkConifg(),
	}

	return cfg, nil
}
