package main

import (
	"os"
	"sort"

	"github.com/joho/godotenv"
	"github.com/kv-base-hack/base-server-api/internal/httputil"
	"github.com/kv-base-hack/base-server-api/internal/server"
	"github.com/kv-base-hack/base-server-api/lib/coingecko"
	"github.com/kv-base-hack/base-server-api/storage"
	"github.com/kv-base-hack/base-server-api/storage/db"
	"github.com/kv-base-hack/base-server-api/worker"
	inmem "github.com/kv-base-hack/common/inmem_db"
	"github.com/kv-base-hack/common/logger"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

func main() {
	_ = godotenv.Load()
	app := cli.NewApp()
	app.Action = run
	app.Flags = append(app.Flags, logger.NewSentryFlags()...)
	app.Flags = append(app.Flags, NewPostgreSQLFlags()...)
	app.Flags = append(app.Flags, NewRedisFlags()...)
	app.Flags = append(app.Flags, NewFlags()...)
	app.Flags = append(app.Flags, httputil.NewHTTPCliFlags(httputil.Port)...)

	sort.Sort(cli.FlagsByName(app.Flags))

	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
}

func run(c *cli.Context) error {
	logger, flusher, err := logger.NewLogger(c)
	if err != nil {
		return err
	}
	defer flusher()

	zap.ReplaceGlobals(logger)
	log := logger.Sugar()
	log.Debugw("Starting application...")

	store := storage.NewStorage(log)

	database, err := NewDBFromContext(c)
	if err != nil {
		log.Errorw("error when connect to database", "err", err)
		return err
	}

	pg := db.NewPostgres(database)

	redisHost := c.String(redisHostFlag)
	redisPort := c.String(redisPortFlag)
	redisPassword := c.String(redisPasswordFlag)
	redisDB := c.Int(redisDBFlag)
	redisAddr := redisHost + ":" + redisPort
	redis := inmem.NewRedisClient(redisAddr, redisPassword, redisDB)

	getRate := worker.NewGetRate(log, redis, c.Duration(getRateDuration), store)
	getRate.Init()
	go getRate.Run()

	tokenInfo := worker.NewTokenInfoWorker(log, c.Duration(tokenInfoDuration), redis, store)
	tokenInfo.Init()
	go tokenInfo.Run()

	solLogs := worker.NewSolanaLogs(log, c.Duration(getDataFromDbDuration),
		pg, store, c.Int64(solFromBlock), c.Int64(maxRangeBlock))
	go solLogs.Run()

	coingecko := coingecko.NewCoinGecko()
	getTrendingWorker := worker.NewGetTrendingWorker(log, coingecko, store)
	go getTrendingWorker.Run()

	host := httputil.NewHTTPAddressFromContext(c)
	server := server.NewServer(host, store, redis)
	return server.Run()
}
