package bricklink

import (
	"LegoManagerAPI/internal/config/configUtilities"
)

// BricklinkConfig hold the Bricklink API credentials
type BricklinkConfig struct {
	SignatureMethod   string
	ConsumerKey       string
	ConsumerSecret    string
	AccessToken       string
	AccessTokenSecret string
}

// LoadBricklinkConifg initializes and returns a BricklinkConfig struct populated with values from env vars.
func LoadBricklinkConifg() BricklinkConfig {
	return BricklinkConfig{
		SignatureMethod:   "HMAC-SHA1",
		ConsumerSecret:    configUtilities.GetEnvAsString("BRICKLINK_CONSUMER_SECRET", "consumer_secret"),
		ConsumerKey:       configUtilities.GetEnvAsString("BRICKLINK_CONSUMER_KEY", "consumer_key"),
		AccessToken:       configUtilities.GetEnvAsString("BRICKLINK_ACCESS_TOKEN", "access_token"),
		AccessTokenSecret: configUtilities.GetEnvAsString("BRICKLINK_ACCESS_TOKEN_SECRET", "access_token_secret"),
	}
}
