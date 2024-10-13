package server

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kv-base-hack/base-server-api/common"
	"github.com/kv-base-hack/common/utils"
)

type GetTokenProfitRequest struct {
	Duration time.Duration `form:"duration" binding:"required"`
	Start    int           `form:"start" binding:"required,numeric,min=1"`
	Limit    int           `form:"limit" binding:"required,numeric,min=1"`
	Chain    string        `form:"chain" binding:"required"`
}

type GetTokenProfitRes struct {
	TokenAddressResponse
	Gains   float64 `json:"gains"`
	AvgCost float64 `json:"avg_cost"`
	NetFlow float64 `json:"net_flow"`
}

func (s *Server) getTokenProfit(c *gin.Context) {
	log := s.log.With("ID", utils.RandomString(29))
	now := time.Now()
	defer func() {
		log.Debugw("Execution time", "getTokenProfit", time.Since(now))
	}()

	var request GetTokenProfitRequest
	if err := c.ShouldBindQuery(&request); err != nil {
		log.Errorw("invalid request when get token profit", "err", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": ErrInvalidTopTokenProfitRequest.Error()})
		return
	}
	chain, err := common.ChainString(request.Chain)
	if err != nil {
		log.Errorw("invalid request when get token profit", "err", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": ErrInvalidListToken.Error()})
		return
	}

	tradeLogs, err := s.storage.GetTradeLogs(chain, request.Duration)
	if err != nil {
		log.Errorw("invalid duration when get token profit", "duration", request.Duration, "err", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Infow("get top token profit",
		"request", request,
		"tradeBlockTs", tradeLogs.StorageByRangeIndex)

	addrToTokenInfo := s.storage.GetTokenInfo(chain)

	topTokenProfit := s.getTopToken(chain, tradeLogs.TokenProfit, addrToTokenInfo, request.Start, request.Limit)

	tokenInFlowInUsdt, err := s.storage.GetTokenInFlowInUsdt(chain, request.Duration)
	if err != nil {
		log.Errorw("invalid request when get token profit", "err", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": ErrInvalidTopTokenProfitRequest.Error()})
		return
	}
	tokenInFlow, err := s.storage.GetTokenInFlow(chain, request.Duration)
	if err != nil {
		log.Errorw("invalid request when get token profit", "err", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": ErrInvalidTopTokenProfitRequest.Error()})
		return
	}
	tokenOutFlow, err := s.storage.GetTokenOutFlow(chain, request.Duration)
	if err != nil {
		log.Errorw("invalid request when get token profit", "err", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": ErrInvalidTopTokenProfitRequest.Error()})
		return
	}
	res := []GetTokenProfitRes{}
	for _, t := range topTokenProfit {
		addr := strings.ToLower(t.Addr)
		inUdst := tokenInFlowInUsdt[addr]
		in := tokenInFlow[addr]
		out := tokenOutFlow[addr]
		res = append(res, GetTokenProfitRes{
			TokenAddressResponse: t,
			AvgCost:              inUdst / in,
			Gains:                t.Value,
			NetFlow:              in - out,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"top_token_profit": res,
	})
}

type GetTokenInspectSellBuy struct {
	Chain    string        `form:"chain" binding:"required"`
	Address  string        `form:"address" binding:"required"`
	Duration time.Duration `form:"duration" binding:"required"`
}

func (s *Server) tokenInspectBuySell(c *gin.Context) {
	log := s.log.With("ID", utils.RandomString(29))

	var request GetTokenInspectSellBuy
	if err := c.ShouldBindQuery(&request); err != nil {
		log.Errorw("invalid request when get token inspect sell buy", "err", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": ErrInvalidTokenInspect.Error()})
		return
	}
	chain, err := common.ChainString(request.Chain)
	if err != nil {
		log.Errorw("invalid request when get token profit", "err", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": ErrInvalidListToken.Error()})
		return
	}

	tradeLogs, err := s.storage.GetTradeLogs(chain, request.Duration)
	if err != nil {
		log.Errorw("invalid request when get token inspect sell buy", "err", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": ErrInvalidTokenInspect.Error()})
		return
	}
	addr := strings.ToLower(request.Address)
	inFlowIntoken := tradeLogs.TokenInFlow[addr]
	inFlowInUsdt := tradeLogs.TokenInFlowInUsdt[addr]

	outFlowIntoken := tradeLogs.TokenOutFlow[addr]
	outFlowInUsdt := tradeLogs.TokenOutFlowInUsdt[addr]

	c.JSON(http.StatusOK, gin.H{
		"in_flow_in_token":  inFlowIntoken,
		"in_flow_in_usdt":   inFlowInUsdt,
		"out_flow_in_token": outFlowIntoken,
		"out_flow_in_usdt":  outFlowInUsdt,
	})
}

type GetTokenInspectDepositWithdraw struct {
	Chain    string        `form:"chain" binding:"required"`
	Address  string        `form:"address" binding:"required"`
	Duration time.Duration `form:"duration" binding:"required"`
}

func (s *Server) tokenInspectDepositWithdraw(c *gin.Context) {
	log := s.log.With("ID", utils.RandomString(29))

	var request GetTokenInspectDepositWithdraw
	if err := c.ShouldBindQuery(&request); err != nil {
		log.Errorw("invalid request when get token inspect deposit withdraw", "err", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": ErrInvalidTokenInspect.Error()})
		return
	}

	chain, err := common.ChainString(request.Chain)
	if err != nil {
		log.Errorw("invalid request when get token profit", "err", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": ErrInvalidListToken.Error()})
		return
	}

	transfer, err := s.storage.GetTransferLogs(chain, request.Duration)
	if err != nil {
		log.Errorw("invalid request when get token inspect deposit withdraw", "err", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": ErrInvalidTokenInspect.Error()})
		return
	}

	addr := strings.ToLower(request.Address)
	cexInFlowInUsdt := transfer.CexInFlowInUsdt[addr]
	cexInFlow := transfer.CexInFlow[addr]

	cexOutFlowInUsdt := transfer.CexOutFlowInUsdt[addr]
	cexOutFlow := transfer.CexOutFlow[addr]

	c.JSON(http.StatusOK, gin.H{
		"cex_in_flow":          cexInFlow,
		"cex_in_flow_in_usdt":  cexInFlowInUsdt,
		"cex_out_flow_in_usdt": cexOutFlowInUsdt,
		"cex_out_flow":         cexOutFlow,
	})
}

type GetTokenInspectActivitiesRequest struct {
	Action       string `form:"action" binding:"required"`
	Chain        string `form:"chain" binding:"required"`
	TokenAddress string `form:"address" binding:"required"`
	Start        int    `form:"start" binding:"required,numeric,min=1"`
	Limit        int    `form:"limit" binding:"required,numeric,min=1"`
}

type GetTokenInspectActivitiesResponse struct {
	common.BigTx
	TokenSymbol   string `json:"symbol"`
	TokenImageUrl string `json:"token_image_url"`
	ChainID       string `json:"chainId"`
}

func (s *Server) tokenInspectActivities(c *gin.Context) {
	log := s.log.With("ID", utils.RandomString(29))

	var request GetTokenInspectActivitiesRequest
	if err := c.ShouldBindQuery(&request); err != nil {
		log.Errorw("invalid request when get token inspect activities", "err", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": ErrInvalidTokenInspect.Error()})
		return
	}

	chain, err := common.ChainString(request.Chain)
	if err != nil {
		log.Errorw("invalid request when get token inspect activities", "err", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": ErrInvalidListToken.Error()})
		return
	}

	action, err := common.SmartMoneyActivitiesString(request.Action)
	if err != nil {
		log.Errorw("invalid request when get token inspect activities", "err", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": ErrInvalidListToken.Error()})
		return
	}

	activities := s.storage.GetLastBigTxForToken(chain, action, defaultLength, request.TokenAddress)
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

type ListTokenRequest struct {
	Chain        string `form:"chain" binding:"required"`
	SymbolSearch string `form:"symbol_search"`
}

type ListTokenResponse struct {
	Symbol   string  `json:"symbol"`
	UsdPrice float64 `json:"usdPrice"`
	Address  string  `json:"tokenAddress"`
	ChainID  string  `json:"chainId"`
	ImageUrl string  `json:"imageUrl"`
}

func (s *Server) listToken(c *gin.Context) {
	log := s.log.With("ID", utils.RandomString(29))

	var request ListTokenRequest
	if err := c.ShouldBindQuery(&request); err != nil {
		log.Errorw("invalid request when get list token", "err", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": ErrInvalidListToken.Error()})
		return
	}
	log.Infow("get list tokens", "chain", request.Chain)
	chain, err := common.ChainString(request.Chain)
	if err != nil {
		log.Errorw("invalid request when get list token", "err", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": ErrInvalidListToken.Error()})
		return
	}
	tokens := s.storage.GetTokens(chain)
	addrToTokenInfo := s.storage.GetTokenInfo(chain)
	limit := 10
	res := []ListTokenResponse{}
	search := strings.ToLower(request.SymbolSearch)
	for _, t := range tokens {
		info := addrToTokenInfo[strings.ToLower(t)]

		pass := search == "" || strings.Contains(strings.ToLower(t), search) ||
			strings.Contains(strings.ToLower(info.Symbol), search)

		if pass && len(res) < limit {
			res = append(res, ListTokenResponse{
				Symbol:   info.Symbol,
				UsdPrice: info.UsdPrice,
				Address:  info.Address,
				ChainID:  info.ChainID,
				ImageUrl: info.ImageUrl,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"tokens": res,
	})
}

type TokenTrendingReponse struct {
	Name                     string  `json:"name"`
	Symbol                   string  `json:"symbol"`
	Thumb                    string  `json:"thumb"`
	Small                    string  `json:"small"`
	Price                    float64 `json:"price"`
	MarketCap                string  `json:"market_cap"`
	TotalVolume              string  `json:"total_volume"`
	PriceChangePercentage24h float64 `json:"price_change_percentage_24h"`
	Address                  string  `json:"address"`
	ChainID                  string  `json:"chain_id"`
}

func (s *Server) getTokenTrending(c *gin.Context) {
	log := s.log.With("ID", utils.RandomString(29))
	log.Infow("get trending tokens")

	addrToTokenInfo := s.storage.GetTokenInfo(common.ChainBase)
	mappingSymbolToTokenInfo := map[string]common.Token{}
	for _, t := range addrToTokenInfo {
		mappingSymbolToTokenInfo[t.Symbol] = t
	}

	t := s.storage.GetTrendingToken()
	res := []TokenTrendingReponse{}
	for _, t := range t.Coins {
		info := mappingSymbolToTokenInfo[t.Item.Symbol]
		var addr string
		var chainID string
		if info.ChainID == common.ChainBase.String() {
			addr = info.Address
			chainID = info.ChainID
		}
		res = append(res, TokenTrendingReponse{
			Name:                     t.Item.Name,
			Symbol:                   t.Item.Symbol,
			Thumb:                    t.Item.Thumb,
			Small:                    t.Item.Small,
			Price:                    t.Item.Data.Price,
			MarketCap:                t.Item.Data.MarketCap,
			TotalVolume:              t.Item.Data.TotalVolume,
			PriceChangePercentage24h: t.Item.Data.PriceChangePercentage24h.Eur,
			Address:                  addr,
			ChainID:                  chainID,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"trending_tokens": res,
	})
}

type TokenInfoRequest struct {
	Chain   string `form:"chain" binding:"required"`
	Address string `form:"address" binding:"required"`
}

func (s *Server) getTokenInfo(c *gin.Context) {
	log := s.log.With("ID", utils.RandomString(29))
	log.Infow("get token info")

	var request TokenInfoRequest
	if err := c.ShouldBindQuery(&request); err != nil {
		log.Errorw("invalid request when get token info", "err", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": ErrInvalidGetTokenInfo.Error()})
		return
	}

	chain, err := common.ChainString(request.Chain)
	if err != nil {
		log.Errorw("invalid request when get token info", "err", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": ErrInvalidGetTokenInfo.Error()})
		return
	}

	addrToTokenInfo := s.storage.GetTokenInfo(chain)
	token := addrToTokenInfo[strings.ToLower(request.Address)]
	info := s.storage.GetTokenInfoFromSymbol(token.Symbol)

	c.JSON(http.StatusOK, gin.H{
		"info": info,
	})
}

type PriceWithTransferRequest struct {
	Chain   string `form:"chain" binding:"required"`
	Address string `form:"address" binding:"required"`
}

type PriceWithTransferResponse struct {
	Date     string  `json:"date"`
	Deposit  float64 `json:"deposit"`
	Withdraw float64 `json:"withdraw"`
	Price    float64 `json:"price"`
}

func (s *Server) getPriceWithTransfer(c *gin.Context) {
	log := s.log.With("ID", utils.RandomString(29))
	log.Infow("get price with transfer")

	var request PriceWithTransferRequest
	if err := c.ShouldBindQuery(&request); err != nil {
		log.Errorw("invalid request when get price with transfer", "err", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": ErrInvalidGetTokenInfo.Error()})
		return
	}

	chain, err := common.ChainString(request.Chain)
	if err != nil {
		log.Errorw("invalid request when get price with transfer", "err", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": ErrInvalidListToken.Error()})
		return
	}

	deposit, withdraw := s.storage.GetPriceWithTransferByRange(chain, strings.ToLower(request.Address))
	res := map[string]PriceWithTransferResponse{}
	for date, value := range deposit {
		res[date] = PriceWithTransferResponse{
			Date:     date,
			Deposit:  value,
			Withdraw: withdraw[date],
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"price_with_transfer": res,
	})
}
