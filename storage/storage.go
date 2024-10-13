package storage

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/kv-base-hack/base-server-api/common"
	"github.com/kv-base-hack/base-server-api/lib/coingecko"
	"github.com/kv-base-hack/base-server-api/util"
	"go.uber.org/zap"
)

const bigVolumeInUsdt = 50_000

type StorageByRangeIndex struct {
	StartIndex   int       // point to first trade logs of this chain that in duration
	StartBlockTs time.Time // use to debug, blockTs of StartIndex block
	StartBlock   uint64    // use to debug, block number of StartIndex block
	EndBlockTs   time.Time // use to debug, end block ts
	EndBlock     uint64    // use to debug, end block number
}

type TradeStorageByRange struct {
	duration time.Duration

	UserProfit  map[string]float64
	TokenProfit map[string]float64

	TokenInFlowInUsdt map[string]float64
	TokenInFlow       map[string]float64

	TokenOutFlowInUsdt map[string]float64
	TokenOutFlow       map[string]float64
	StorageByRangeIndex
}

type TransferStorageByRange struct {
	duration time.Duration

	CexInFlow       map[string]float64
	CexInFlowInUsdt map[string]float64

	CexOutFlow       map[string]float64
	CexOutFlowInUsdt map[string]float64

	StorageByRangeIndex
}

// format date: dd/mm/yyyy
type TokenTransfer map[string]float64

type ChainData struct {
	network           common.Chain
	tradeLogs         []common.Tradelog
	transferLogs      []common.Transferlog
	addrToTokenInfo   map[string]common.Token
	tradeDataRange    []TradeStorageByRange
	transferDataRange []TransferStorageByRange
	tokens            map[string]bool
	bigTx             []common.BigTx
	tokenDeposit      map[string]TokenTransfer
	tokenWithdraw     map[string]TokenTransfer
}

func NewTradeStorageByRange(duration time.Duration) TradeStorageByRange {
	return TradeStorageByRange{
		duration: duration, // 1h

		UserProfit:  make(map[string]float64),
		TokenProfit: make(map[string]float64),

		TokenInFlowInUsdt: make(map[string]float64),
		TokenInFlow:       make(map[string]float64),

		TokenOutFlowInUsdt: make(map[string]float64),
		TokenOutFlow:       make(map[string]float64),

		StorageByRangeIndex: StorageByRangeIndex{
			StartIndex: -1,
		},
	}
}

func NewTransferStorageByRange(duration time.Duration) TransferStorageByRange {
	return TransferStorageByRange{
		duration:        duration,
		CexInFlow:       make(map[string]float64),
		CexInFlowInUsdt: make(map[string]float64),

		CexOutFlow:       make(map[string]float64),
		CexOutFlowInUsdt: make(map[string]float64),

		StorageByRangeIndex: StorageByRangeIndex{
			StartIndex: -1,
		},
	}
}

type Storage struct {
	log   *zap.SugaredLogger
	mutex sync.RWMutex
	// we lower case all token in this map
	tokenUsdtRate  map[string]float64
	trendingTokens coingecko.CoingeckoTrending
	symbolToInfo   map[string]common.CmcTokenInfo
	chains         map[common.Chain]*ChainData
}

func NewStorage(log *zap.SugaredLogger) *Storage {
	baseTradeDataByRange := []TradeStorageByRange{
		NewTradeStorageByRange(time.Hour),
		NewTradeStorageByRange(time.Hour * 4),
		NewTradeStorageByRange(time.Hour * 24),      // 1 day
		NewTradeStorageByRange(time.Hour * 24 * 7),  // 1 week
		NewTradeStorageByRange(time.Hour * 24 * 30), // 1 month
	}

	baseTransferDataByRange := []TransferStorageByRange{
		NewTransferStorageByRange(time.Hour),
		NewTransferStorageByRange(time.Hour * 4),
		NewTransferStorageByRange(time.Hour * 24),      // 1 day
		NewTransferStorageByRange(time.Hour * 24 * 7),  // 1 week
		NewTransferStorageByRange(time.Hour * 24 * 30), // 1 month
	}

	return &Storage{
		log: log,
		chains: map[common.Chain]*ChainData{
			common.ChainBase: {
				network:           common.ChainBase,
				tradeLogs:         make([]common.Tradelog, 0),
				transferLogs:      make([]common.Transferlog, 0),
				addrToTokenInfo:   make(map[string]common.Token),
				tradeDataRange:    baseTradeDataByRange,
				transferDataRange: baseTransferDataByRange,
				tokens:            make(map[string]bool),
				bigTx:             make([]common.BigTx, 0),
				tokenDeposit:      make(map[string]TokenTransfer),
				tokenWithdraw:     make(map[string]TokenTransfer),
			},
		},
		tokenUsdtRate: make(map[string]float64),
		symbolToInfo:  make(map[string]common.CmcTokenInfo),
	}
}

func (s *Storage) AddTradeLogs(chain common.Chain, logs []common.Tradelog) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for _, log := range logs {
		tokenIn := strings.ToLower(log.TokenInAddress)
		tokenOut := strings.ToLower(log.TokenOutAddress)
		sender := strings.ToLower(log.Sender)

		s.chains[chain].tokens[tokenIn] = true
		s.chains[chain].tokens[tokenOut] = true
		if log.GetCurrentRateFail {
			continue
		}
		// we dont remove old trade, so append it too much can make memory leak
		s.chains[chain].tradeLogs = append(s.chains[chain].tradeLogs, log)

		// add big trade
		valueInUsdt := log.TokenOutAmount * log.TokenOutUsdtRate
		if valueInUsdt >= bigVolumeInUsdt {
			action := common.SmartMoneyActivitiesBuying
			// current token to quote token -> selling
			if util.IsQuote(log.TokenOutAddress) {
				action = common.SmartMoneyActivitiesSelling
			}

			s.chains[chain].bigTx = append(s.chains[chain].bigTx, common.BigTx{
				TokenAddress:   log.TokenOutAddress,
				Time:           log.BlockTimestamp,
				Sender:         log.Sender,
				ValueInToken:   log.TokenOutAmount,
				ValueInUsdt:    valueInUsdt,
				Price:          log.TokenOutUsdtRate,
				Movement:       action.String(),
				Action:         action,
				BlockTimestamp: log.BlockTimestamp,
				BlockNumber:    log.BlockNumber,
				Tx:             log.TxHash,
			})
		}

		// add big data range
		for i := range s.chains[chain].tradeDataRange {
			s.chains[chain].tradeDataRange[i].UserProfit[sender] += log.Profit
			s.chains[chain].tradeDataRange[i].TokenProfit[tokenOut] += log.Profit

			s.chains[chain].tradeDataRange[i].TokenInFlowInUsdt[tokenOut] += log.TokenOutAmount * log.TokenOutUsdtRate
			s.chains[chain].tradeDataRange[i].TokenInFlow[tokenOut] += log.TokenOutAmount

			s.chains[chain].tradeDataRange[i].TokenOutFlowInUsdt[tokenIn] += log.TokenInAmount * log.TokenInUsdtRate
			s.chains[chain].tradeDataRange[i].TokenOutFlow[tokenIn] += log.TokenInAmount

			s.chains[chain].tradeDataRange[i].EndBlockTs = log.BlockTimestamp
			s.chains[chain].tradeDataRange[i].EndBlock = log.BlockNumber

			if s.chains[chain].tradeDataRange[i].StartIndex == -1 {
				s.chains[chain].tradeDataRange[i].StartBlockTs = log.BlockTimestamp
				s.chains[chain].tradeDataRange[i].StartBlock = log.BlockNumber
				s.chains[chain].tradeDataRange[i].StartIndex = len(s.chains[chain].tradeLogs) - 1
			}
		}
	}
	s.log.Debugw("trade logs", "chain", chain, "len", len(s.chains[chain].tradeLogs))
}

// heavy action
func (s *Storage) GetTradeLogsForUser(chain common.Chain, from time.Time, user string) []common.Tradelog {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	tradelogs := []common.Tradelog{}
	for _, t := range s.chains[chain].tradeLogs {
		if t.BlockTimestamp.Before(from) {
			continue
		}
		if strings.EqualFold(t.Sender, user) {
			tradelogs = append(tradelogs, t)
		}
	}
	return tradelogs
}

// heavy action
func (s *Storage) GetTradeLogsForToken(chain common.Chain, from time.Time, token string) []common.Tradelog {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	tradelogs := []common.Tradelog{}
	for _, t := range s.chains[chain].tradeLogs {
		if t.BlockTimestamp.Before(from) {
			continue
		}
		if strings.EqualFold(t.TokenInAddress, token) || strings.EqualFold(t.TokenOutAddress, token) {
			tradelogs = append(tradelogs, t)
		}
	}
	return tradelogs
}

func (s *Storage) GetTradeLogs(chain common.Chain, duration time.Duration) (TradeStorageByRange, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for _, t := range s.chains[chain].tradeDataRange {
		if t.duration == duration {
			return t, nil
		}
	}
	return TradeStorageByRange{}, fmt.Errorf("invalid duration to get sol trade logs")
}

func (s *Storage) AddTransferLogs(chain common.Chain, logs []common.Transferlog) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for _, log := range logs {
		token := strings.ToLower(log.TokenAddress)
		s.chains[chain].tokens[token] = true
		if log.GetCurrentRateFail {
			continue
		}
		// we dont remove old transfer, so append it too much can make memory leak
		s.chains[chain].transferLogs = append(s.chains[chain].transferLogs, log)

		// add big transfer
		valueInUsdt := log.TokenAmount * log.CurrentTokenUsdtRate
		if valueInUsdt >= bigVolumeInUsdt {
			sender := log.ToAddress
			if log.IsCexIn {
				sender = log.FromAddress
			}
			action := common.SmartMoneyActivitiesDeposit
			if log.IsCexIn {
				action = common.SmartMoneyActivitiesWithdraw
			}
			s.chains[chain].bigTx = append(s.chains[chain].bigTx, common.BigTx{
				TokenAddress:   log.TokenAddress,
				Sender:         sender,
				Time:           log.BlockTimestamp,
				ValueInToken:   log.TokenAmount,
				ValueInUsdt:    valueInUsdt,
				Price:          log.CurrentTokenUsdtRate,
				Movement:       action.String(),
				Action:         action,
				BlockTimestamp: log.BlockTimestamp,
				BlockNumber:    log.BlockNumber,
				Tx:             log.TxHash,
			})
		}

		// add transfer range data
		for i := range s.chains[chain].transferDataRange {
			if log.IsCexIn {
				s.chains[chain].transferDataRange[i].CexInFlow[token] += log.TokenAmount
				s.chains[chain].transferDataRange[i].CexInFlowInUsdt[token] += log.TokenAmount * log.CurrentTokenUsdtRate
			} else {
				s.chains[chain].transferDataRange[i].CexOutFlow[token] += log.TokenAmount
				s.chains[chain].transferDataRange[i].CexOutFlowInUsdt[token] += log.TokenAmount * log.CurrentTokenUsdtRate
			}
			s.chains[chain].transferDataRange[i].EndBlockTs = log.BlockTimestamp
			s.chains[chain].transferDataRange[i].EndBlock = log.BlockNumber

			if s.chains[chain].transferDataRange[i].StartIndex == -1 {
				s.chains[chain].transferDataRange[i].StartBlockTs = log.BlockTimestamp
				s.chains[chain].transferDataRange[i].StartBlock = log.BlockNumber
				s.chains[chain].transferDataRange[i].StartIndex = len(s.chains[chain].transferLogs) - 1
			}
		}

		formatDate := fmt.Sprintf("%d-%d-%d", log.BlockTimestamp.Day(), int(log.BlockTimestamp.Month()), log.BlockTimestamp.Year())
		if log.IsCexIn {
			if _, exist := s.chains[chain].tokenDeposit[token]; !exist {
				s.chains[chain].tokenDeposit[token] = make(TokenTransfer)
			}
			s.chains[chain].tokenDeposit[token][formatDate] += log.TokenAmount
		} else {
			if _, exist := s.chains[chain].tokenWithdraw[token]; !exist {
				s.chains[chain].tokenWithdraw[token] = make(TokenTransfer)
			}
			s.chains[chain].tokenWithdraw[token][formatDate] += log.TokenAmount
		}
	}

	s.log.Debugw("transfer logs", "chain", chain, "len", len(s.chains[chain].transferLogs))
}

func (s *Storage) GetTransferLogsForToken(chain common.Chain, from time.Time, token string) []common.Transferlog {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	transferlogs := []common.Transferlog{}
	for _, t := range s.chains[chain].transferLogs {
		if strings.EqualFold(t.TokenAddress, token) {
			transferlogs = append(transferlogs, t)
		}
	}
	return transferlogs
}

func (s *Storage) GetTransferLogs(chain common.Chain, duration time.Duration) (TransferStorageByRange, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	for _, t := range s.chains[chain].transferDataRange {
		if t.duration == duration {
			return t, nil
		}
	}

	return TransferStorageByRange{}, fmt.Errorf("invalid duration to get transfer logs")
}

// we lowercase all key
func (s *Storage) SetTokenUsdtRate(rates []common.Token) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	for _, rate := range rates {
		s.tokenUsdtRate[strings.ToLower(rate.Address)] = rate.UsdPrice
	}
}

func (s *Storage) GetTokenUsdtRate() map[string]float64 {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	tokenUsdtRate := map[string]float64{}
	for k, v := range s.tokenUsdtRate {
		tokenUsdtRate[k] = v
	}

	return tokenUsdtRate
}

// set token info from dexscreener
func (s *Storage) SetAddrToTokenInfo(tokens []common.Token) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	for _, t := range tokens {
		if t.ChainID == common.ChainBase.String() {
			s.chains[common.ChainBase].addrToTokenInfo[strings.ToLower(t.Address)] = t
		}
	}
}

// get token info from dexscreener
func (s *Storage) GetTokenInfo(chain common.Chain) map[string]common.Token {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	tokens := map[string]common.Token{}
	for k, v := range s.chains[chain].addrToTokenInfo {
		tokens[k] = v
	}

	return tokens
}

// set token info from dexscreener
func (s *Storage) SetSymbolToTokenInfoFromCmc(tokens common.CmcTokens) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	for _, t := range tokens.Tokens {
		_, exist := s.symbolToInfo[t.Symbol]
		if !exist {
			s.symbolToInfo[t.Symbol] = t
		}
	}
}

// get token info from dexscreener
func (s *Storage) GetTokenInfoFromSymbol(symbol string) common.CmcTokenInfo {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.symbolToInfo[symbol]
}

func (s *Storage) RemoveTrades(sugar *zap.SugaredLogger, chain common.Chain) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for i := range s.chains[chain].tradeDataRange {
		duration := s.chains[chain].tradeDataRange[i].duration
		currentIndex := s.chains[chain].tradeDataRange[i].StartIndex
		if currentIndex == -1 {
			// it means all of the data in transferLog is out of data for this duration -> dont need to remove anything
			continue
		}
		for {
			if currentIndex >= len(s.chains[chain].tradeLogs) ||
				time.Since(s.chains[chain].tradeLogs[currentIndex].BlockTimestamp) <= duration {
				break
			}
			log := s.chains[chain].tradeLogs[currentIndex]
			tokenIn := strings.ToLower(log.TokenInAddress)
			tokenOut := strings.ToLower(log.TokenOutAddress)
			sender := strings.ToLower(log.Sender)

			// old trade, remove it
			s.chains[chain].tradeDataRange[i].UserProfit[sender] -= log.Profit
			s.chains[chain].tradeDataRange[i].TokenProfit[tokenOut] -= log.Profit

			s.chains[chain].tradeDataRange[i].TokenInFlowInUsdt[tokenOut] -= log.TokenOutAmount * log.TokenOutUsdtRate
			s.chains[chain].tradeDataRange[i].TokenInFlow[tokenOut] -= log.TokenOutAmount

			s.chains[chain].tradeDataRange[i].TokenOutFlowInUsdt[tokenIn] -= log.TokenInAmount * log.TokenInUsdtRate
			s.chains[chain].tradeDataRange[i].TokenOutFlow[tokenIn] -= log.TokenInAmount
			currentIndex++
		}
		if currentIndex > s.chains[chain].tradeDataRange[i].StartIndex {
			sugar.Debugw("remove old trade and set new start index",
				"chain", chain,
				"current_index", currentIndex,
				"old_index", s.chains[chain].tradeDataRange[i].StartIndex,
				"tradeLogs", len(s.chains[chain].tradeLogs))
			s.chains[chain].tradeDataRange[i].StartIndex = currentIndex
			if currentIndex < len(s.chains[chain].tradeLogs) {
				s.chains[chain].tradeDataRange[i].StartBlockTs = s.chains[chain].tradeLogs[currentIndex].BlockTimestamp
				s.chains[chain].tradeDataRange[i].StartBlock = s.chains[chain].tradeLogs[currentIndex].BlockNumber
			}
		}
	}
}

func (s *Storage) RemoveTransfer(sugar *zap.SugaredLogger, chain common.Chain) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for i := range s.chains[chain].transferDataRange {
		duration := s.chains[chain].transferDataRange[i].duration
		currentIndex := s.chains[chain].transferDataRange[i].StartIndex
		if currentIndex == -1 {
			// it means all of the data in transferLog is out of data for this duration -> dont need to remove anything
			continue
		}
		for {
			if currentIndex >= len(s.chains[chain].transferLogs) ||
				time.Since(s.chains[chain].transferLogs[currentIndex].BlockTimestamp) <= duration {
				break
			}
			log := s.chains[chain].transferLogs[currentIndex]
			token := strings.ToLower(log.TokenAddress)

			if log.IsCexIn {
				s.chains[chain].transferDataRange[i].CexInFlow[token] -= log.TokenAmount
				s.chains[chain].transferDataRange[i].CexInFlowInUsdt[token] -= log.TokenAmount * log.CurrentTokenUsdtRate
			} else {
				s.chains[chain].transferDataRange[i].CexOutFlow[token] -= log.TokenAmount
				s.chains[chain].transferDataRange[i].CexOutFlowInUsdt[token] -= log.TokenAmount * log.CurrentTokenUsdtRate
			}

			currentIndex++
		}
		if currentIndex > s.chains[chain].transferDataRange[i].StartIndex {
			sugar.Debugw("remove old transfer and set new start index",
				"chain", chain,
				"current_index", currentIndex,
				"old_index", s.chains[chain].transferDataRange[i].StartIndex,
				"transferLogs", len(s.chains[chain].transferLogs))

			s.chains[chain].transferDataRange[i].StartIndex = currentIndex
			if currentIndex < len(s.chains[chain].transferLogs) {
				s.chains[chain].transferDataRange[i].StartBlockTs = s.chains[chain].transferLogs[currentIndex].BlockTimestamp
				s.chains[chain].transferDataRange[i].StartBlock = s.chains[chain].transferLogs[currentIndex].BlockNumber
			}
		}
	}
}

func (s *Storage) GetTokens(chain common.Chain) []string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	res := []string{}
	for k := range s.chains[chain].tokens {
		res = append(res, k)
	}
	return res
}

func (s *Storage) GetLastBigTx(chain common.Chain, action common.SmartMoneyActivities, last int) []common.BigTx {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	res := []common.BigTx{}
	for i := len(s.chains[chain].bigTx) - 1; i >= 0; i-- {
		if action == common.SmartMoneyActivitiesAll || action == s.chains[chain].bigTx[i].Action {
			res = append(res, s.chains[chain].bigTx[i])
		}
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].BlockNumber > res[j].BlockNumber
	})
	return res
}

func (s *Storage) GetLastBigTxForToken(chain common.Chain, action common.SmartMoneyActivities, last int, tokenAddress string) []common.BigTx {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	res := []common.BigTx{}
	for i := len(s.chains[chain].bigTx) - 1; i >= 0; i-- {
		if !strings.EqualFold(s.chains[chain].bigTx[i].TokenAddress, tokenAddress) {
			continue
		}
		if action == common.SmartMoneyActivitiesAll || action == s.chains[chain].bigTx[i].Action {
			res = append(res, s.chains[chain].bigTx[i])
		}
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].BlockNumber > res[j].BlockNumber
	})
	return res
}

func (s *Storage) GetLastBigTxForUser(chain common.Chain, action common.SmartMoneyActivities, last int, userAddress string) []common.BigTx {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	res := []common.BigTx{}
	for i := len(s.chains[chain].bigTx) - 1; i >= 0; i-- {
		if !strings.EqualFold(s.chains[chain].bigTx[i].Sender, userAddress) {
			continue
		}
		if action == common.SmartMoneyActivitiesAll || action == s.chains[chain].bigTx[i].Action {
			res = append(res, s.chains[chain].bigTx[i])
		}
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].BlockNumber > res[j].BlockNumber
	})
	return res
}

func (s *Storage) SetTrendingToken(t coingecko.CoingeckoTrending) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.trendingTokens = t
}

func (s *Storage) GetTrendingToken() coingecko.CoingeckoTrending {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.trendingTokens
}

func (s *Storage) GetTokenInFlowInUsdt(chain common.Chain, duration time.Duration) (map[string]float64, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for _, t := range s.chains[chain].tradeDataRange {
		if t.duration == duration {
			return t.TokenInFlowInUsdt, nil
		}
	}

	return map[string]float64{}, fmt.Errorf("invalid duration to get trade logs")
}

func (s *Storage) GetTokenInFlow(chain common.Chain, duration time.Duration) (map[string]float64, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for _, t := range s.chains[chain].tradeDataRange {
		if t.duration == duration {
			return t.TokenInFlow, nil
		}
	}

	return map[string]float64{}, fmt.Errorf("invalid duration to get trade logs")
}

func (s *Storage) GetTokenOutFlowInUsdt(chain common.Chain, duration time.Duration) (map[string]float64, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for _, t := range s.chains[chain].tradeDataRange {
		if t.duration == duration {
			return t.TokenOutFlowInUsdt, nil
		}
	}

	return map[string]float64{}, fmt.Errorf("invalid duration to get sol transfer logs")
}

func (s *Storage) GetTokenOutFlow(chain common.Chain, duration time.Duration) (map[string]float64, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for _, t := range s.chains[chain].tradeDataRange {
		if t.duration == duration {
			return t.TokenOutFlow, nil
		}
	}

	return map[string]float64{}, fmt.Errorf("invalid duration to get sol transfer logs")
}

func (s *Storage) GetPriceWithTransferByRange(chain common.Chain, token string) (TokenTransfer, TokenTransfer) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.chains[chain].tokenDeposit[token], s.chains[chain].tokenWithdraw[token]
}
