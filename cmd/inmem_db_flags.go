package main

import (
	"github.com/urfave/cli/v2"
)

const (
	redisHostFlag     = "redis-host"
	redisPortFlag     = "redis-port"
	redisPasswordFlag = "redis-password"
	redisDBFlag       = "redis-db"
)

// NewRedisFlags creates new cli flags for PostgreSQL client.
func NewRedisFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:    redisHostFlag,
			EnvVars: []string{"REDIS_HOST"},
		},
		&cli.StringFlag{
			Name:    redisPortFlag,
			EnvVars: []string{"REDIS_PORT"},
		},
		&cli.StringFlag{
			Name:    redisPasswordFlag,
			Value:   "",
			EnvVars: []string{"REDIS_PASS"},
		},
		&cli.IntFlag{
			Name:    redisDBFlag,
			Value:   0,
			EnvVars: []string{"REDIS_DB"},
		},
	}
}
