package main

import (
	"time"

	"github.com/urfave/cli/v2"
)

const (
	getDataFromDbDuration = "get-data-from-db-duration"
	getRateDuration       = "get-rate-duration"
	tokenInfoDuration     = "token-info-duration"
	solFromBlock          = "sol-from-block"
	maxRangeBlock         = "max-range-block"
)

// NewFlags creates new cli flags.
func NewFlags() []cli.Flag {
	return []cli.Flag{
		&cli.DurationFlag{
			Name:    getDataFromDbDuration,
			Value:   time.Second * 3,
			Usage:   "duration to get new log from database",
			EnvVars: []string{"GET_DATA_FROM_DB_DURATION"},
		},
		&cli.Int64Flag{
			Name:    solFromBlock,
			EnvVars: []string{"SOL_FROM_BLOCK"},
		},
		&cli.Int64Flag{
			Name:    maxRangeBlock,
			EnvVars: []string{"MAX_RANGE_BLOCK"},
		},
		&cli.DurationFlag{
			Name:    getRateDuration,
			Value:   time.Second * 10,
			Usage:   "duration to get rate from redis",
			EnvVars: []string{"GET_RATE_DURATION"},
		},
		&cli.DurationFlag{
			Name:    tokenInfoDuration,
			Value:   time.Second * 60,
			Usage:   "duration to get token info from redis",
			EnvVars: []string{"GET_TOKEN_INFO_DURATION"},
		},
	}
}
