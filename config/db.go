package config

import "fmt"

func DatabaseURL() string {
	return fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s sslmode=%s",
		cfg.DatabaseHost,
		cfg.DatabasePort,
		cfg.DatabaseName,
		cfg.DatabaseUser,
		cfg.DatabasePassword,
		cfg.DatabaseSSLMode,
	)
}
