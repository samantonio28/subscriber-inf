package config

import (
	"os"
	"strconv"
)

// applyEnvOverrides переопределяет поля DatabaseConfig из переменных окружения.
// Это позволяет одному и тому же бинарнику работать и на хосте (YAML: localhost:8001),
// и в контейнере (env: DB_HOST=postgres DB_PORT=5432).
func applyEnvOverrides(c *DatabaseConfig) {
	if v := os.Getenv("DB_TYPE"); v != "" {
		c.Type = DatabaseType(v)
	}

	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASSWORD")
	name := os.Getenv("DB_NAME")

	if c.Postgres != nil {
		if host != "" {
			c.Postgres.Host = host
		}
		if p, err := strconv.Atoi(port); port != "" && err == nil {
			c.Postgres.Port = p
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

	if c.MySQL != nil {
		if host != "" {
			c.MySQL.Host = host
		}
		if p, err := strconv.Atoi(port); port != "" && err == nil {
			c.MySQL.Port = p
		}
		if user != "" {
			c.MySQL.User = user
		}
		if pass != "" {
			c.MySQL.Password = pass
		}
		if name != "" {
			c.MySQL.DBName = name
		}
	}
}
