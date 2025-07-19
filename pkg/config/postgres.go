package config

import (
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
	SSLMode  string `yaml:"sslmode"`
	Pool     struct {
		MaxConns        int    `yaml:"max_conns"`
		MinConns        int    `yaml:"min_conns"`
		MaxConnLifetime string `yaml:"max_conn_lifetime"`
		MaxConnIdleTime string `yaml:"max_conn_idle_time"`
	} `yaml:"pool"`
}

func (c *PostgresConfig) ToConnectionString() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.User,
		c.Password,
		c.Host,
		c.Port,
		c.DBName,
		c.SSLMode,
	)
}

func (c *PostgresConfig) ToPgxPoolConfig() (*pgxpool.Config, error) {
	connStr := c.ToConnectionString()
	poolConfig, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, err
	}

	poolConfig.MaxConns = int32(c.Pool.MaxConns)
	poolConfig.MinConns = int32(c.Pool.MinConns)

	if c.Pool.MaxConnLifetime != "" {
		duration, err := time.ParseDuration(c.Pool.MaxConnLifetime)
		if err != nil {
			return nil, err
		}
		poolConfig.MaxConnLifetime = duration
	}

	if c.Pool.MaxConnIdleTime != "" {
		duration, err := time.ParseDuration(c.Pool.MaxConnIdleTime)
		if err != nil {
			return nil, err
		}
		poolConfig.MaxConnIdleTime = duration
	}

	return poolConfig, nil
}
