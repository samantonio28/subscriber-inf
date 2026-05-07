package config

import (
	"database/sql"
	"fmt"
	"path/filepath"

	// _ "github.com/mattn/go-sqlite3"
)

type DatabaseType string

const (
	DBTypePostgres DatabaseType = "postgres"
	DBTypeSQLite   DatabaseType = "sqlite"
)

type DatabaseConfig struct {
	Type     DatabaseType   `yaml:"type"`
	Postgres *PostgresConfig `yaml:"postgres,omitempty"`
	SQLite   *SQLiteConfig   `yaml:"sqlite,omitempty"`
}

type SQLiteConfig struct {
	Path string `yaml:"path"`
}

func (c *DatabaseConfig) Open() (*sql.DB, error) {
	switch c.Type {
	case DBTypePostgres:
		return c.openPostgres()
	case DBTypeSQLite:
		return c.openSQLite()
	default:
		return nil, fmt.Errorf("unsupported database type: %s", c.Type)
	}
}

func (c *DatabaseConfig) openPostgres() (*sql.DB, error) {
	if c.Postgres == nil {
		return nil, fmt.Errorf("postgres config is missing")
	}
	connStr := c.Postgres.ToConnectionString()
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	// Configure connection pool
	db.SetMaxOpenConns(c.Postgres.Pool.MaxConns)
	db.SetMaxIdleConns(c.Postgres.Pool.MinConns)
	// TODO: parse MaxConnLifetime and MaxConnIdleTime
	return db, nil
}

func (c *DatabaseConfig) openSQLite() (*sql.DB, error) {
	if c.SQLite == nil {
		return nil, fmt.Errorf("sqlite config is missing")
	}
	path := c.SQLite.Path
	// Ensure directory exists
	dir := filepath.Dir(path)
	if dir != "." && dir != "" {
		// TODO: create directory if needed
	}
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	// SQLite doesn't need connection pooling in the same way, but we can set limits
	db.SetMaxOpenConns(1) // SQLite recommends single connection for write safety
	db.SetMaxIdleConns(1)
	return db, nil
}