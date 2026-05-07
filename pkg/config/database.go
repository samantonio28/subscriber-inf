package config

import (
	"database/sql"
	"fmt"
	"path/filepath"

	_ "github.com/go-sql-driver/mysql"
)

type DatabaseType string

const (
	DBTypePostgres DatabaseType = "postgres"
	DBTypeSQLite   DatabaseType = "sqlite"
	DBTypeMySQL    DatabaseType = "mysql"
)

type DatabaseConfig struct {
	Type     DatabaseType    `yaml:"type"`
	Postgres *PostgresConfig `yaml:"postgres,omitempty"`
	SQLite   *SQLiteConfig   `yaml:"sqlite,omitempty"`
	MySQL    *MySQLConfig    `yaml:"mysql,omitempty"`
}

type SQLiteConfig struct {
	Path string `yaml:"path"`
}

type MySQLConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
	Params   string `yaml:"params,omitempty"`
}

func (c *DatabaseConfig) Open() (*sql.DB, error) {
	switch c.Type {
	case DBTypePostgres:
		return c.openPostgres()
	case DBTypeSQLite:
		return c.openSQLite()
	case DBTypeMySQL:
		return c.openMySQL()
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

func (c *DatabaseConfig) openMySQL() (*sql.DB, error) {
	if c.MySQL == nil {
		return nil, fmt.Errorf("mysql config is missing")
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?%s",
		c.MySQL.User,
		c.MySQL.Password,
		c.MySQL.Host,
		c.MySQL.Port,
		c.MySQL.DBName,
		c.MySQL.Params,
	)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	// Configure connection pool with defaults (could be extended with config)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	return db, nil
}
