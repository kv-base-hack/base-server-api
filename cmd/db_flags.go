package main

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/jmoiron/sqlx"
)

const (
	postgresHostFlag     = "postgres-host"
	postgresPortFlag     = "postgres-port"
	postgresUserFlag     = "postgres-user"
	postgresPasswordFlag = "postgres-password"
	postgresDatabaseFlag = "postgres-database"
)

// NewPostgreSQLFlags creates new cli flags for PostgreSQL client.
func NewPostgreSQLFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:    postgresHostFlag,
			EnvVars: []string{"POSTGRES_HOST"},
		},
		&cli.StringFlag{
			Name:    postgresUserFlag,
			EnvVars: []string{"POSTGRES_USER"},
		},
		&cli.StringFlag{
			Name:    postgresPasswordFlag,
			EnvVars: []string{"POSTGRES_PASSWORD"},
		},
		&cli.StringFlag{
			Name:    postgresDatabaseFlag,
			EnvVars: []string{"POSTGRES_DB"},
		},
		&cli.Int64Flag{
			Name:    postgresPortFlag,
			EnvVars: []string{"POSTGRES_PORT"},
		},
	}
}

// NewDBFromContext creates a DB instance from cli flags configuration.
func NewDBFromContext(c *cli.Context) (*sqlx.DB, error) {
	const driverName = "postgres"
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		c.String(postgresHostFlag),
		c.Int(postgresPortFlag),
		c.String(postgresUserFlag),
		c.String(postgresPasswordFlag),
		c.String(postgresDatabaseFlag),
	)
	return sqlx.Connect(driverName, connStr)
}

// DatabaseNameFromContext return database name
func DatabaseNameFromContext(c *cli.Context) string {
	return c.String(postgresDatabaseFlag)
}
