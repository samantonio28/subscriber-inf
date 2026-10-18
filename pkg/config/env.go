package config

import (
	"os"
	"strconv"
)

// applyEnvOverrides переопределяет поля Config из переменных окружения.
// Это позволяет одному и тому же бинарнику работать и на хосте (YAML: localhost:8001),
// и в контейнере (env: DB_HOST=postgres DB_PORT=5432).
func applyEnvOverrides(c *Config) {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASSWORD")
	name := os.Getenv("DB_NAME")

	if host != "" {
		c.Postgres.Host = host
	}
	if port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			c.Postgres.Port = p
		}
	}
	if user != "" {
		c.Postgres.User = user
	}
	if pass != "" {
		c.Postgres.Password = pass
	}
	if name != "" {
		c.Postgres.DBName = name
	}
}
