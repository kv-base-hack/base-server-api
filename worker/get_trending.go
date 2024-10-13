package worker

import (
	"time"

	"github.com/kv-base-hack/base-server-api/lib/coingecko"
	"github.com/kv-base-hack/base-server-api/storage"
	"go.uber.org/zap"
)

type GetTrendingWorker struct {
	log       *zap.SugaredLogger
	coingecko *coingecko.CoinGecko
	storage   *storage.Storage
}

func NewGetTrendingWorker(log *zap.SugaredLogger, coingecko *coingecko.CoinGecko, storage *storage.Storage) *GetTrendingWorker {
	return &GetTrendingWorker{
		log:       log,
		coingecko: coingecko,
		storage:   storage,
	}
}

func (g *GetTrendingWorker) Run() {
	t := time.NewTicker(time.Hour * 6)
	for ; ; <-t.C {
		g.Do()
	}
}

func (g *GetTrendingWorker) Do() {
	trendingToken, err := g.coingecko.GetTrending()
	if err != nil {
		g.log.Errorw("error when get trending worker", "err", err)
		return
	}
	g.log.Debugw("set trending token", "trendingToken", trendingToken)
	g.storage.SetTrendingToken(trendingToken)
}
