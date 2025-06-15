package config

import "fmt"

func DatabaseURL() string {
	// Use DATABASE_URL if available (includes SSL configuration)
	if cfg.DatabaseURL != "" {
		return cfg.DatabaseURL
	}

	// Fallback to constructing URL with SSL
	return fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s sslmode=%s",
		cfg.DatabaseHost,
		cfg.DatabasePort,
		cfg.DatabaseName,
		cfg.DatabaseUser,
		cfg.DatabasePassword,
		cfg.DatabaseSSLMode,
	)
}
