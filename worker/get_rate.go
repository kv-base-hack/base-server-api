package worker

import (
	"encoding/json"
	"time"

	"github.com/kv-base-hack/base-server-api/common"
	"github.com/kv-base-hack/base-server-api/storage"
	inmem "github.com/kv-base-hack/common/inmem_db"
	"go.uber.org/zap"
)

const ratePricesKey = "dex_screener_prices"

type GetRate struct {
	log      *zap.SugaredLogger
	inMemDB  inmem.Inmem
	duration time.Duration
	storage  *storage.Storage
}

func NewGetRate(log *zap.SugaredLogger, inMemDB inmem.Inmem, duration time.Duration, storage *storage.Storage) *GetRate {
	return &GetRate{
		log:      log.With("worker", "getRate"),
		inMemDB:  inMemDB,
		duration: duration,
		storage:  storage,
	}
}

func (r *GetRate) Init() {
	r.process()
}

func (r *GetRate) Run() {
	ticker := time.NewTicker(r.duration)
	for ; ; <-ticker.C {
		r.process()
	}
}

func (r *GetRate) process() {
	rates, err := r.inMemDB.Get(ratePricesKey)
	if err != nil {
		r.log.Errorw("error when get rate", "err", err)
		return
	}
	var ratesList []common.Token
	if err = json.Unmarshal([]byte(rates), &ratesList); err != nil {
		r.log.Errorw("error when parse rate", "rates", rates, "err", err)
		return
	}
	r.storage.SetTokenUsdtRate(ratesList)
	r.storage.SetAddrToTokenInfo(ratesList)
}
