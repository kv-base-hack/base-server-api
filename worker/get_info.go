package worker

import (
	"encoding/json"
	"time"

	"github.com/kv-base-hack/base-server-api/common"
	"github.com/kv-base-hack/base-server-api/storage"
	inmem "github.com/kv-base-hack/common/inmem_db"
	"go.uber.org/zap"
)

const cmcTokenInfoKey = "cmc_token_info"

type TokenInfoWorker struct {
	log      *zap.SugaredLogger
	duration time.Duration
	inMemDB  inmem.Inmem
	storage  *storage.Storage
}

func NewTokenInfoWorker(log *zap.SugaredLogger, duration time.Duration, inMemDB inmem.Inmem, storage *storage.Storage) *TokenInfoWorker {
	return &TokenInfoWorker{
		log:      log.With("worker", "token_info"),
		duration: duration,
		inMemDB:  inMemDB,
		storage:  storage,
	}
}

func (t *TokenInfoWorker) Init() {
	t.process()
}

func (t *TokenInfoWorker) Run() {
	ticker := time.NewTicker(t.duration)
	for ; ; <-ticker.C {
		t.process()
	}
}

func (t *TokenInfoWorker) process() {
	infoBytes, err := t.inMemDB.Get(cmcTokenInfoKey)
	if err != nil {
		t.log.Errorw("error when get token info", "err", err)
		return
	}

	var info common.CmcTokens
	if err = json.Unmarshal([]byte(infoBytes), &info); err != nil {
		t.log.Errorw("error when parse token info", "info", infoBytes, "err", err)
		return
	}
	t.storage.SetSymbolToTokenInfoFromCmc(info)
}
