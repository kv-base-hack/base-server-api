package server

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kv-base-hack/base-server-api/common"
	"github.com/kv-base-hack/common/utils"
)

type GetUserProfitRequest struct {
	Duration time.Duration `form:"duration" binding:"required"`
	Start    int           `form:"start" binding:"required,numeric,min=1"`
	Limit    int           `form:"limit" binding:"required,numeric,min=1"`
	Chain    string        `form:"chain" binding:"required"`
}

func (s *Server) getUserProfit(c *gin.Context) {
	log := s.log.With("ID", utils.RandomString(29))
	now := time.Now()
	defer func() {
		log.Debugw("Execution time", "getUserProfit", time.Since(now))
	}()

	var request GetUserProfitRequest
	if err := c.ShouldBindQuery(&request); err != nil {
		log.Errorw("invalid request when get user profit", "err", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": ErrInvalidTopUserProfitRequest.Error()})
		return
	}

	chain, err := common.ChainString(request.Chain)
	if err != nil {
		log.Errorw("invalid request when get top user profit", "err", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": ErrInvalidListToken.Error()})
		return
	}

	tradeLogs, err := s.storage.GetTradeLogs(chain, request.Duration)
	if err != nil {
		log.Errorw("invalid duration when get user profit", "duration", request.Duration, "err", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Infow("get top user profit",
		"chain", chain,
		"request", request,
		"tradeBlockTs", tradeLogs.StorageByRangeIndex)

	arrData := []Data{}
	for k, v := range tradeLogs.UserProfit {
		arrData = append(arrData, Data{
			key:   k,
			value: v,
		})
	}

	sort.Slice(arrData, func(i, j int) bool {
		return arrData[i].value > arrData[j].value
	})

	var topUserProfit []UserAddressResponse
	st := (request.Start - 1) * request.Limit
	ed := st + request.Limit - 1

	for i := st; i <= ed; i++ {
		if i >= len(arrData) {
			break
		}
		u := arrData[i]
		topUserProfit = append(topUserProfit, UserAddressResponse{
			AddressResponse: AddressResponse{
				Addr:  u.key,
				Value: u.value,
				Chain: request.Chain,
			},
			// TODO: add new of address
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"top_user_profit": topUserProfit,
	})
}

type UserInspect struct {
	Chain    string        `form:"chain" binding:"required"`
	Address  string        `form:"address" binding:"required"`
	Duration time.Duration `form:"duration" binding:"required"`
}

func (s *Server) userInspect(c *gin.Context) {
	log := s.log.With("ID", utils.RandomString(29))

	var request UserInspect
	if err := c.ShouldBindQuery(&request); err != nil {
		log.Errorw("invalid request when get user inspect", "err", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": ErrInvalidUserInspect.Error()})
		return
	}

	chain, err := common.ChainString(request.Chain)
	if err != nil {
		log.Errorw("invalid request when get top user profit", "err", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": ErrInvalidListToken.Error()})
		return
	}

	fromTime := time.Now().Add(-request.Duration)
	tradeLogs := s.storage.GetTradeLogsForUser(chain, fromTime, request.Address)

	txProfit := make(map[string]float64)

	for _, t := range tradeLogs {
		if t.GetCurrentRateFail {
			continue
		}

		txProfit[t.TxHash] = txProfit[t.TxHash] + t.Profit
	}

	c.JSON(http.StatusOK, gin.H{
		"tx_profit": txProfit,
	})
}

type GetUserInspectActivitiesRequest struct {
	Action      string `form:"action" binding:"required"`
	Chain       string `form:"chain" binding:"required"`
	UserAddress string `form:"address" binding:"required"`
	Start       int    `form:"start" binding:"required,numeric,min=1"`
	Limit       int    `form:"limit" binding:"required,numeric,min=1"`
}

type GetUserInspectActivitiesResponse struct {
	common.BigTx
	TokenSymbol   string `json:"symbol"`
	TokenImageUrl string `json:"token_image_url"`
	ChainID       string `json:"chainId"`
}

func (s *Server) userInspectActivities(c *gin.Context) {
	log := s.log.With("ID", utils.RandomString(29))

	var request GetUserInspectActivitiesRequest
	if err := c.ShouldBindQuery(&request); err != nil {
		log.Errorw("invalid request when get user inspect activities", "err", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": ErrInvalidTokenInspect.Error()})
		return
	}

	chain, err := common.ChainString(request.Chain)
	if err != nil {
		log.Errorw("invalid request when get user inspect activities", "err", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": ErrInvalidListToken.Error()})
		return
	}

	action, err := common.SmartMoneyActivitiesString(request.Action)
	if err != nil {
		log.Errorw("invalid request when get user inspect activities", "err", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": ErrInvalidListToken.Error()})
		return
	}

	activities := s.storage.GetLastBigTxForUser(chain, action, defaultLength, request.UserAddress)
	addrToTokenInfo := s.storage.GetTokenInfo(chain)
	st := (request.Start - 1) * request.Limit
	ed := st + request.Limit - 1

	act := []GetActivitiesResponse{}
	for i := st; i <= ed; i++ {
		if i >= len(activities) {
			break
		}
		a := activities[i]
		info := addrToTokenInfo[strings.ToLower(a.TokenAddress)]
		act = append(act, GetActivitiesResponse{
			BigTx:         a,
			TokenSymbol:   info.Symbol,
			TokenImageUrl: info.ImageUrl,
			ChainID:       info.ChainID,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"activities": act,
	})
}

type GetUserBalanceRequest struct {
	Chain   string `form:"chain" binding:"required"`
	Address string `form:"address" binding:"required"`
}

type GetUserBalanceResponse struct {
	Address       string                 `json:"address,omitempty"`
	TotalBalance  float64                `json:"total_balance,omitempty"`
	Profit        float64                `json:"profit,omitempty"`
	PnlPercent    float64                `json:"pnl_percent,omitempty"`
	TokenBalances []TokenBalanceResponse `json:"token_balances,omitempty"`
}

type TokenBalanceResponse struct {
	Symbol     string  `json:"symbol,omitempty"`
	ImageUrl   string  `json:"imageUrl,omitempty"`
	Amount     float64 `json:"amount,omitempty"`
	TotalSpent float64 `json:"total_spent,omitempty"`
	Pnl        float64 `json:"pnl,omitempty"`
}

func (s *Server) getUserBalances(c *gin.Context) {
	log := s.log.With("ID", utils.RandomString(29))

	var request GetUserBalanceRequest
	if err := c.ShouldBindQuery(&request); err != nil {
		log.Errorw("invalid request when get user balances", "err", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": ErrInvalidGetUserBalances.Error()})
		return
	}

	chain, err := common.ChainString(request.Chain)
	if err != nil {
		log.Errorw("invalid request when get user inspect activities", "err", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": ErrInvalidListToken.Error()})
		return
	}

	trades, err := s.storage.GetTradeLogs(chain, time.Hour*24)
	if err != nil {
		log.Errorw("invalid request when get user balances", "err", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": ErrInvalidGetUserBalances.Error()})
		return
	}
	profit := trades.UserProfit[request.Address]
	totalBalance := 0.0

	balancesStr, err := s.inMemDB.Get(chain.String() + "_" + strings.ToLower(request.Address))
	if err != nil {
		log.Errorw("couldn't get user balance", "err", err)
	}

	var balances []common.TokenBalance
	if err := json.Unmarshal([]byte(balancesStr), &balances); err != nil {
		log.Errorw("couldn't parse user balance user balance", "balancesStr", balancesStr, "err", err)
	}

	userBalances := []TokenBalanceResponse{}

	addrToTokenInfo := s.storage.GetTokenInfo(chain)
	for _, balance := range balances {
		info := addrToTokenInfo[strings.ToLower(balance.Address)]
		userBalances = append(userBalances, TokenBalanceResponse{
			Symbol:   info.Symbol,
			ImageUrl: info.ImageUrl,
			Amount:   balance.Amount,
		})
		totalBalance += info.UsdPrice * balance.Amount
	}

	var pnlPercent float64
	if totalBalance > 0 {
		pnlPercent = profit / totalBalance * 100
	}

	res := GetUserBalanceResponse{
		Address:       request.Address,
		TotalBalance:  totalBalance,
		Profit:        profit,
		PnlPercent:    pnlPercent,
		TokenBalances: userBalances,
	}

	c.JSON(http.StatusOK, gin.H{
		"balances": res,
	})
}

type GetUserPortfolioRequest struct {
	Chain   string `form:"chain" binding:"required"`
	Address string `form:"address" binding:"required"`
	Start   int    `form:"start" binding:"required,numeric,min=1"`
	Limit   int    `form:"limit" binding:"required,numeric,min=1"`
}

func (s *Server) getUserPortfolio(c *gin.Context) {
	log := s.log.With("ID", utils.RandomString(29))

	var request GetUserPortfolioRequest
	if err := c.ShouldBindQuery(&request); err != nil {
		log.Errorw("invalid request when get user balances", "err", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": ErrInvalidGetUserBalances.Error()})
		return
	}

	chain, err := common.ChainString(request.Chain)
	if err != nil {
		log.Errorw("invalid request when get user inspect activities", "err", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": ErrInvalidListToken.Error()})
		return
	}

	balancesStr, err := s.inMemDB.Get(chain.String() + "_" + strings.ToLower(request.Address))
	if err != nil {
		log.Errorw("couldn't get user balance", "err", err)
	}

	var balances []common.TokenBalance
	if err := json.Unmarshal([]byte(balancesStr), &balances); err != nil {
		log.Errorw("couldn't parse user balance user balance", "balancesStr", balancesStr, "err", err)
	}

	tokens := []TokenBalanceResponse{}
	addrToTokenInfo := s.storage.GetTokenInfo(chain)

	st := (request.Start - 1) * request.Limit
	ed := st + request.Limit - 1
	for i := st; i <= ed; i++ {
		if i >= len(balances) {
			break
		}
		info := addrToTokenInfo[strings.ToLower(balances[i].Address)]
		tokens = append(tokens, TokenBalanceResponse{
			Symbol:   info.Symbol,
			ImageUrl: info.ImageUrl,
			Amount:   balances[i].Amount,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"tokens": tokens,
		"total":  len(balances),
	})
}
