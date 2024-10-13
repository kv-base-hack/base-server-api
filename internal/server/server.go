package server

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/kv-base-hack/base-server-api/common"
	"github.com/kv-base-hack/base-server-api/storage"
	"github.com/kv-base-hack/base-server-api/util"
	inmem "github.com/kv-base-hack/common/inmem_db"
	"github.com/kv-base-hack/common/utils"
	"go.uber.org/zap"
)

const defaultLength = 100

// Server to serve the service.
type Server struct {
	s        *gin.Engine
	bindAddr string
	log      *zap.SugaredLogger
	storage  *storage.Storage
	inMemDB  inmem.Inmem
}

// New returns a new server.
func NewServer(bindAddr string, storage *storage.Storage, inMemDB inmem.Inmem) *Server {
	engine := gin.New()

	engine.Use(gin.Recovery())

	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"*"}
	config.AddAllowHeaders("Digest", "Authorization", "Signature", "Nonce")

	engine.Use(cors.New(config))
	engine.Use(gin.Recovery())

	s := &Server{
		s:        engine,
		log:      zap.S(),
		bindAddr: bindAddr,
		storage:  storage,
		inMemDB:  inMemDB,
	}

	gin.SetMode(gin.DebugMode)
	s.register()

	return s
}

// Run runs server.
func (s *Server) Run() error {
	s.log.Debugw("run in ", "s.bindAddr", s.bindAddr)
	if err := s.s.Run(s.bindAddr); err != nil {
		return fmt.Errorf("run server: %w", err)
	}
	return nil
}

func (s *Server) register() {
	s.s.GET("/debug/pprof/*all", gin.WrapH(http.DefaultServeMux))
	v1 := s.s.Group("/v1")

	v1.GET("/token_cex_in", s.getTopCexIn)
	v1.GET("/token_cex_out", s.getTopCexOut)
	v1.GET("/activities", s.getActivities)
	v1.GET("/leaderboard", s.getLeaderboard)

	token := v1.Group("token")
	token.GET("/profit", s.getTokenProfit)
	token.GET("/inspect/depositwithdraw", s.tokenInspectDepositWithdraw)
	token.GET("/inspect/buysell", s.tokenInspectBuySell)
	token.GET("/inspect/activities", s.tokenInspectActivities)
	token.GET("/list", s.listToken)
	token.GET("/trending", s.getTokenTrending)
	token.GET("/info", s.getTokenInfo)
	token.GET("/price_with_transfer", s.getPriceWithTransfer)

	user := v1.Group("user")
	user.GET("/profit", s.getUserProfit)
	user.GET("/inspect", s.userInspect)
	user.GET("/inspect/activities", s.userInspectActivities)
	user.GET("/balances", s.getUserBalances)
	user.GET("/portfolio", s.getUserPortfolio)
}

type AddressResponse struct {
	Addr  string  `json:"address"`
	Value float64 `json:"value"`
	Chain string  `json:"network,omitempty"`
}

type TokenAddressResponse struct {
	AddressResponse
	Symbol       string  `json:"symbol"`
	CurrentPrice float64 `json:"current_price"`
	ImageUrl     string  `json:"image_url"`
}

type UserAddressResponse struct {
	AddressResponse
	Name string `json:"name,omitempty"`
}

type TopCexInRequest struct {
	Duration time.Duration `form:"duration" binding:"required"`
	Start    int           `form:"start" binding:"required,numeric,min=1"`
	Limit    int           `form:"limit" binding:"required,numeric,min=1"`
	Chain    string        `form:"chain" binding:"required"`
}

type Data struct {
	key   string
	value float64
}

func (s *Server) getTopToken(chain common.Chain, data map[string]float64,
	addrToTokenInfo map[string]common.Token, start, limit int) []TokenAddressResponse {
	arrData := []Data{}

	for k, v := range data {
		arrData = append(arrData, Data{
			key:   k,
			value: v,
		})
	}

	sort.Slice(arrData, func(i, j int) bool {
		return arrData[i].value > arrData[j].value
	})

	top := make([]TokenAddressResponse, 0)
	st := (start - 1) * limit
	ed := st + limit - 1
	for i := st; i <= ed; i++ {
		if i >= len(arrData) {
			break
		}
		t := arrData[i]
		info := addrToTokenInfo[strings.ToLower(t.key)]
		top = append(top, TokenAddressResponse{
			AddressResponse: AddressResponse{
				Addr:  t.key,
				Value: t.value,
				Chain: chain.String(),
			},
			Symbol:       info.Symbol,
			CurrentPrice: info.UsdPrice,
			ImageUrl:     info.ImageUrl,
		})
	}
	return top
}

func (s *Server) getTopCexIn(c *gin.Context) {
	log := s.log.With("ID", utils.RandomString(29))
	now := time.Now()
	defer func() {
		log.Debugw("Execution time", "getTopCexIn", time.Since(now))
	}()

	var request TopCexInRequest
	if err := c.ShouldBindQuery(&request); err != nil {
		log.Errorw("invalid request when get top cex in", "err", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": ErrInvalidTopCexInRequest.Error()})
		return
	}

	chain, err := common.ChainString(request.Chain)
	if err != nil {
		log.Errorw("invalid request when get top cex in", "err", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": ErrInvalidListToken.Error()})
		return
	}

	transferLogs, err := s.storage.GetTransferLogs(chain, request.Duration)
	if err != nil {
		log.Errorw("invalid duration when get top cex in", "duration", request.Duration, "err", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Infow("get top cex in",
		"request", request,
		"transferBlockTs", transferLogs.StorageByRangeIndex)

	addrToTokenInfo := s.storage.GetTokenInfo(chain)
	topCexIn := s.getTopToken(chain, transferLogs.CexInFlowInUsdt, addrToTokenInfo, request.Start, request.Limit)

	c.JSON(http.StatusOK, gin.H{
		"top_cex_in": topCexIn,
		"total":      len(transferLogs.CexInFlowInUsdt),
	})
}

type TopCexOutRequest struct {
	Duration time.Duration `form:"duration" binding:"required"`
	Start    int           `form:"start" binding:"required,numeric,min=1"`
	Limit    int           `form:"limit" binding:"required,numeric,min=1"`
	Chain    string        `form:"chain" binding:"required"`
}

func (s *Server) getTopCexOut(c *gin.Context) {
	log := s.log.With("ID", utils.RandomString(29))
	now := time.Now()
	defer func() {
		log.Debugw("Execution time", "getTopCexOut", time.Since(now))
	}()

	var request TopCexOutRequest
	if err := c.ShouldBindQuery(&request); err != nil {
		log.Errorw("invalid request when get top cex out", "err", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": ErrInvalidTopCexOutRequest.Error()})
		return
	}

	chain, err := common.ChainString(request.Chain)
	if err != nil {
		log.Errorw("invalid request when get top cex in", "err", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": ErrInvalidListToken.Error()})
		return
	}

	transferLogs, err := s.storage.GetTransferLogs(chain, request.Duration)
	if err != nil {
		log.Errorw("invalid duration when get top cex out", "duration", request.Duration, "err", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Infow("get top cex out",
		"request", request,
		"transferBlockTs", transferLogs.StorageByRangeIndex)

	addrToTokenInfo := s.storage.GetTokenInfo(chain)
	topCexOut := s.getTopToken(chain, transferLogs.CexOutFlowInUsdt, addrToTokenInfo, request.Start, request.Limit)

	c.JSON(http.StatusOK, gin.H{
		"top_cex_out": topCexOut,
		"total":       len(transferLogs.CexOutFlowInUsdt),
	})
}

type GetActivitiesRequest struct {
	Action string `form:"action" binding:"required"`
	Start  int    `form:"start" binding:"required,numeric,min=1"`
	Limit  int    `form:"limit" binding:"required,numeric,min=1"`
	Chain  string `form:"chain" binding:"required"`
}

type GetActivitiesResponse struct {
	common.BigTx
	TokenSymbol   string `json:"symbol"`
	TokenImageUrl string `json:"token_image_url"`
	ChainID       string `json:"chain_id"`
}

func (s *Server) getActivities(c *gin.Context) {
	log := s.log.With("ID", utils.RandomString(29))
	now := time.Now()
	defer func() {
		log.Debugw("Execution time", "getActivities", time.Since(now))
	}()

	var request GetActivitiesRequest
	if err := c.ShouldBindQuery(&request); err != nil {
		log.Errorw("invalid request when get activities", "err", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": ErrInvalidGetActivities.Error()})
		return
	}

	chain, err := common.ChainString(request.Chain)
	if err != nil {
		log.Errorw("invalid request when get list user", "err", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": ErrInvalidListToken.Error()})
		return
	}

	action, err := common.SmartMoneyActivitiesString(request.Action)
	if err != nil {
		log.Errorw("invalid request when get list user", "err", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": ErrInvalidListToken.Error()})
		return
	}

	addrToTokenInfo := s.storage.GetTokenInfo(chain)
	activities := s.storage.GetLastBigTx(chain, action, defaultLength)
	if err != nil {
		log.Errorw("invalid request when get list user", "err", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": ErrInvalidListToken.Error()})
		return
	}
	act := []GetActivitiesResponse{}
	st := (request.Start - 1) * request.Limit
	ed := st + request.Limit - 1

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
		"total":      len(activities),
	})
}

type GetLeaderboardRequest struct {
	Start int    `form:"start" binding:"required,numeric,min=1"`
	Limit int    `form:"limit" binding:"required,numeric,min=1"`
	Chain string `form:"chain" binding:"required"`
}

type GetLeaderboardResponse struct {
	UserAddress string  `json:"user_address"`
	NetProfit   float64 `json:"net_profit"`
	// MostProfitableTradeToken common.Token `json:"most_profitable_trade_token"`
	CurrentLargestPosition common.Token `json:"current_largest_position"`
	MostTokenBuy           common.Token `json:"most_token_buy"`
	MostTokenSell          common.Token `json:"most_token_sell"`
	LastTrade              time.Time    `json:"last_trade"`
}

func (s *Server) getLeaderboard(c *gin.Context) {
	log := s.log.With("ID", utils.RandomString(29))
	now := time.Now()
	defer func() {
		log.Debugw("Execution time", "getLeaderboard", time.Since(now))
	}()

	var request GetLeaderboardRequest
	if err := c.ShouldBindQuery(&request); err != nil {
		log.Errorw("invalid request when get activities", "err", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": ErrInvalidGetActivities.Error()})
		return
	}

	chain, err := common.ChainString(request.Chain)
	if err != nil {
		log.Errorw("invalid request when get list user", "err", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": ErrInvalidListToken.Error()})
		return
	}

	tradeLogs, err := s.storage.GetTradeLogs(chain, time.Hour*24)
	if err != nil {
		log.Errorw("invalid duration when get user profit", "err", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

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

	var res []GetLeaderboardResponse
	from := time.Now().Add(-time.Hour * 24)
	addrToTokenInfo := s.storage.GetTokenInfo(chain)

	for _, t := range topUserProfit {
		trade := s.storage.GetTradeLogsForUser(chain, from, t.Addr)
		lastTrade := from
		if len(trade) > 0 {
			lastTrade = trade[len(trade)-1].BlockTimestamp
		}

		mostTokenIn := map[string]float64{}
		mostTokenOut := map[string]float64{}
		var mostTokenInValue float64
		var mostTokenInAddress string
		var mostTokenOutValue float64
		var mostTokenOutAddress string

		var largestPositionAddress string
		var largestPositionInUsdtValue float64

		for _, tra := range trade {
			tokenIn := strings.ToLower(tra.TokenInAddress)
			tokenOut := strings.ToLower(tra.TokenOutAddress)
			valuePosition := tra.CurrentTokenOutUsdtRate * tra.TokenOutAmount
			mostTokenIn[tokenOut] += valuePosition
			mostTokenOut[tokenIn] += tra.CurrentTokenInUsdtRate * tra.TokenInAmount

			if mostTokenInValue < mostTokenIn[tokenOut] {
				mostTokenInAddress = tra.TokenOutAddress
				mostTokenInValue = mostTokenIn[tokenOut]
			}
			if mostTokenOutValue < mostTokenOut[tokenIn] {
				mostTokenOutAddress = tra.TokenInAddress
				mostTokenOutValue = mostTokenOut[tokenIn]
			}
			if !util.IsQuote(tokenOut) && largestPositionInUsdtValue < valuePosition {
				largestPositionInUsdtValue = valuePosition
				largestPositionAddress = tra.TokenOutAddress
			}
		}

		mostTokenInInfo := addrToTokenInfo[strings.ToLower(mostTokenInAddress)]
		mostTokenOutInfo := addrToTokenInfo[strings.ToLower(mostTokenOutAddress)]
		largestPositionTokenInfo := addrToTokenInfo[strings.ToLower(largestPositionAddress)]

		res = append(res, GetLeaderboardResponse{
			UserAddress:            t.Addr,
			NetProfit:              t.Value,
			LastTrade:              lastTrade,
			MostTokenBuy:           mostTokenInInfo,
			MostTokenSell:          mostTokenOutInfo,
			CurrentLargestPosition: largestPositionTokenInfo,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"leaderboard": res,
	})
}
