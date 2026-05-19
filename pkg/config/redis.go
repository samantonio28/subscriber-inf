package config

import "fmt"

type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

func (c *RedisConfig) Addr() string {
	if c.Host == "" {
		c.Host = "localhost"
	}
	if c.Port == 0 {
		c.Port = 6379
	}
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

func (c *RedisConfig) WithDefaults() *RedisConfig {
	if c.Host == "" {
		c.Host = "localhost"
	}
	if c.Port == 0 {
		c.Port = 6379
	}
	if c.DB == 0 {
		c.DB = 0
	}
	return c
}
