package worker

import (
	"strings"
	"time"

	"github.com/kv-base-hack/base-server-api/common"
	"github.com/kv-base-hack/base-server-api/storage"
	"github.com/kv-base-hack/base-server-api/storage/db"
	"go.uber.org/zap"
)

const limitLogs = 200000

type SolanaLogs struct {
	log               *zap.SugaredLogger
	duration          time.Duration
	storage           *storage.Storage
	db                db.DB
	lastTradeBlock    int64
	lastTransferBlock int64
	maxRangeBlock     int64
}

func NewSolanaLogs(log *zap.SugaredLogger, duration time.Duration,
	db db.DB, storage *storage.Storage, lastBlock int64, maxRangeBlock int64) *SolanaLogs {
	return &SolanaLogs{
		log:               log.With("worker", "getSolanaLogs"),
		duration:          duration,
		db:                db,
		storage:           storage,
		lastTradeBlock:    lastBlock,
		lastTransferBlock: lastBlock,
		maxRangeBlock:     maxRangeBlock,
	}
}

func (g *SolanaLogs) Run() {
	now := time.Now()
	g.init()
	g.log.Debugw("Execution time", "init", time.Since(now))
	ticker := time.NewTicker(g.duration)
	for ; ; <-ticker.C {
		g.process()
	}
}

func (g *SolanaLogs) handleTrades(trades []db.SolanaTradelogDB) []common.Tradelog {
	ratesMap := g.storage.GetTokenUsdtRate()
	logs := []common.Tradelog{}

	for _, t := range trades {
		solTradeLog := t.Convert()
		currentRateOfTokenIn, exist := ratesMap[strings.ToLower(t.TokenInAddress)]
		if !exist {
			solTradeLog.GetCurrentRateFail = true
			solTradeLog.Profit = 0
			continue
		}
		currentRateOfTokenOut, exist := ratesMap[strings.ToLower(t.TokenOutAddress)]
		if !exist {
			solTradeLog.GetCurrentRateFail = true
			solTradeLog.Profit = 0
			continue
		}
		profitOfTokenIn := (currentRateOfTokenIn - t.TokenInUsdtRate) * t.TokenInAmount
		profitOfTokenOut := (currentRateOfTokenOut - t.TokenOutUsdtRate) * t.TokenOutAmount

		solTradeLog.CurrentTokenInUsdtRate = currentRateOfTokenIn
		solTradeLog.CurrentTokenOutUsdtRate = currentRateOfTokenOut
		solTradeLog.GetCurrentRateFail = false
		solTradeLog.Profit = profitOfTokenOut - profitOfTokenIn

		logs = append(logs, solTradeLog)
	}
	return logs
}

func (g *SolanaLogs) initSolanaTrade() {
	currentBlock, err := g.db.GetMaxBlockNumber(db.SolanaTradeTable)
	lastTradeBlock := g.lastTradeBlock
	var lastTradeBlockTs time.Time
	if err == nil && currentBlock-g.maxRangeBlock > lastTradeBlock {
		lastTradeBlock = currentBlock - g.maxRangeBlock
	}
	trades := []db.SolanaTradelogDB{}
	for {
		oldTrades, err := g.db.GetSolTrades(lastTradeBlock, limitLogs)
		g.log.Infow("initSolanaTrade",
			"currentBlock", currentBlock,
			"lastTradeBlock", lastTradeBlock,
			"lastTradeBlockTs", lastTradeBlockTs,
			"oldTrades", len(oldTrades))
		if err != nil {
			g.log.Errorw("error when init old trades", "lastTradeBlock", lastTradeBlock, "err", err)
			return
		}
		if len(oldTrades) == 0 {
			break
		}
		lastTradeBlock = int64(oldTrades[len(oldTrades)-1].BlockNumber)
		lastTradeBlockTs = oldTrades[len(oldTrades)-1].BlockTimestamp
		// remove old tx added to trades that get again from oldTrades
		for len(trades) > 0 && trades[len(trades)-1].BlockNumber == oldTrades[0].BlockNumber {
			trades = trades[:len(trades)-1]
		}
		trades = append(trades, oldTrades...)
		if lastTradeBlock >= currentBlock {
			break
		}
	}

	logs := g.handleTrades(trades)
	g.storage.AddTradeLogs(common.ChainBase, logs)

	g.lastTradeBlock = lastTradeBlock
}

func (g *SolanaLogs) handleTransfer(transfers []db.SolanaTransferLogDb) []common.Transferlog {
	ratesMap := g.storage.GetTokenUsdtRate()
	logs := []common.Transferlog{}
	for _, t := range transfers {
		transfer := t.Convert()
		currentRate, exist := ratesMap[strings.ToLower(t.TokenAddress)]
		if !exist {
			transfer.GetCurrentRateFail = true
			continue
		}
		transfer.GetCurrentRateFail = false
		transfer.CurrentTokenUsdtRate = currentRate
		logs = append(logs, transfer)
	}
	return logs
}

func (g *SolanaLogs) initSolanaTransfer() {
	currentBlock, err := g.db.GetMaxBlockNumber(db.SolanaTransferTable)
	lastTransferBlock := g.lastTransferBlock
	var lastTransferBlockTs time.Time
	if err == nil && currentBlock-g.maxRangeBlock > lastTransferBlock {
		lastTransferBlock = currentBlock - g.maxRangeBlock
	}

	transfer := []db.SolanaTransferLogDb{}
	for {
		oldTransfers, err := g.db.GetSolTransfer(lastTransferBlock, limitLogs)
		g.log.Infow("initSolanaTransfer",
			"currentBlock", currentBlock,
			"lastTransferBlock", lastTransferBlock,
			"lastTransferBlockTs", lastTransferBlockTs,
			"oldTransfers", len(oldTransfers))
		if err != nil {
			g.log.Errorw("error when init old oldTransfers", "lastTransferBlock", lastTransferBlock, "err", err)
			return
		}
		if len(oldTransfers) == 0 {
			break
		}
		lastTransferBlock = int64(oldTransfers[len(oldTransfers)-1].BlockNumber)
		lastTransferBlockTs = oldTransfers[len(oldTransfers)-1].BlockTimestamp
		// remove old tx added to transfer that get again from oldTransfers
		for len(transfer) > 0 && transfer[len(transfer)-1].BlockNumber == oldTransfers[0].BlockNumber {
			transfer = transfer[:len(transfer)-1]
		}
		transfer = append(transfer, oldTransfers...)
		if lastTransferBlock >= currentBlock {
			break
		}
	}

	logs := g.handleTransfer(transfer)
	g.storage.AddTransferLogs(common.ChainBase, logs)

	g.lastTransferBlock = lastTransferBlock
}

func (g *SolanaLogs) init() {
	now := time.Now()
	g.initSolanaTrade()
	g.initSolanaTransfer()
	g.log.Infow("Execution time", "init duration(s)", time.Since(now).Seconds())
}

func (g *SolanaLogs) processNewTrade() {
	newTrades, err := g.db.GetSolTrades(g.lastTradeBlock+1, limitLogs)
	if err != nil {
		g.log.Errorw("error when init new trades", "block", g.lastTradeBlock+1, "err", err)
		return
	}
	lenNewTrades := len(newTrades)
	g.log.Debugw("add new trade", "block", g.lastTradeBlock+1, "len", lenNewTrades)
	if lenNewTrades == 0 {
		return
	}

	logs := g.handleTrades(newTrades)
	g.storage.AddTradeLogs(common.ChainBase, logs)

	if lenNewTrades > 0 {
		g.lastTradeBlock = int64(newTrades[lenNewTrades-1].BlockNumber)
	}
}

func (g *SolanaLogs) processNewTransfer() {
	newTransfer, err := g.db.GetSolTransfer(g.lastTransferBlock+1, limitLogs)
	if err != nil {
		g.log.Errorw("error when init new transfer", "block", g.lastTransferBlock+1, "err", err)
		return
	}
	lenNewTransfer := len(newTransfer)
	g.log.Debugw("add new transfer", "block", g.lastTransferBlock+1, "len", lenNewTransfer)
	if lenNewTransfer == 0 {
		return
	}
	logs := g.handleTransfer(newTransfer)
	g.storage.AddTransferLogs(common.ChainBase, logs)

	if lenNewTransfer > 0 {
		g.lastTransferBlock = int64(newTransfer[lenNewTransfer-1].BlockNumber)
	}
}

func (g *SolanaLogs) removeStaleTrade() {
	g.storage.RemoveTrades(g.log, common.ChainBase)
}

func (g *SolanaLogs) removeStaleTransfer() {
	g.storage.RemoveTransfer(g.log, common.ChainBase)
}

func (g *SolanaLogs) process() {
	now := time.Now()
	g.processNewTrade()
	g.processNewTransfer()
	g.removeStaleTrade()
	g.removeStaleTransfer()
	g.log.Infow("Execution time", "process duration(s)", time.Since(now).Seconds())
}
