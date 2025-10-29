package config

import (
	"fmt"
	"os"
)

type Config struct {
	DBHost string
	DBPort string
	DBUser string
	DBPass string
	DBName string
}

func Load() *Config {
	return &Config{
		DBHost: getenv("DB_HOST", "db"),
		DBPort: getenv("DB_PORT", "5432"),
		DBUser: getenv("DB_USER", "postgres"),
		DBPass: getenv("DB_PASSWORD", "postgres"),
		DBName: getenv("DB_NAME", "colaboradores_db"),
	}
}

func (c *Config) DatabaseDSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		c.DBHost, c.DBPort, c.DBUser, c.DBPass, c.DBName)
}

func getenv(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}
