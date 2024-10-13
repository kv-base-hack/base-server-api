package common

import (
	"time"
)

// enumer -type=Chain -linecomment -json=true -text=true -sql=true
type Chain uint64

const (
	ChainBase Chain = iota + 1 // base
)

// enumer -type=SourcePrice -linecomment -json=true -text=true -sql=true
type SourcePrice uint64

const (
	SourcePriceCex SourcePrice = iota + 1 // cex
	SourcePriceDex                        // dex
)

// enumer -type=SmartMoneyActivities -linecomment -json=true -text=true -sql=true
type SmartMoneyActivities uint64

const (
	SmartMoneyActivitiesAll      SmartMoneyActivities = iota + 1 // all
	SmartMoneyActivitiesDeposit                                  // deposit
	SmartMoneyActivitiesWithdraw                                 // withdraw
	SmartMoneyActivitiesBuying                                   // buying
	SmartMoneyActivitiesSelling                                  // selling
)

type Tradelog struct {
	BlockTimestamp time.Time `json:"timestamp"`
	BlockNumber    uint64    `json:"block_number"`
	TxIndex        uint      `json:"tx_index,omitempty"`
	TxHash         string    `json:"tx_hash"`
	Sender         string    `json:"sender"`
	LogIndex       uint      `json:"log_index"`

	TokenInAddress  string  `json:"token_in_address"`
	TokenInAmount   float64 `json:"token_in_amount"`
	TokenInUsdtRate float64 `json:"token_in_usdt_rate"`

	TokenOutAddress  string  `json:"token_out_address"`
	TokenOutAmount   float64 `json:"token_out_amount"`
	TokenOutUsdtRate float64 `json:"token_out_usdt_rate"`

	NativeTokenUsdtRate float64 `json:"native_token_usdt_rate"`

	CurrentTokenInUsdtRate  float64
	CurrentTokenOutUsdtRate float64
	Profit                  float64 `json:"profit"`
	GetCurrentRateFail      bool
}

type Transferlog struct {
	BlockTimestamp time.Time `json:"timestamp"`
	BlockNumber    uint64    `json:"block_number"`
	TxIndex        uint      `json:"tx_index,omitempty"`
	TxHash         string    `json:"tx_hash"`

	FromAddress string `json:"from_address"`
	ToAddress   string `json:"to_address"`

	TokenAddress string  `json:"token_in_address"`
	TokenAmount  float64 `json:"token_in_amount"`
	IsCexIn      bool    `json:"is_cex_in"`

	CurrentTokenUsdtRate float64
	GetCurrentRateFail   bool
}

type Token struct {
	UsdPrice    float64     `json:"usdPrice"`
	Address     string      `json:"tokenAddress"`
	Symbol      string      `json:"symbol"`
	ChainID     string      `json:"chainId"`
	SourcePrice SourcePrice `json:"sourcePrice"`
	ImageUrl    string      `json:"imageUrl"`
	DexID       string      `json:"dexId"`
	Url         string      `json:"url"`

	PriceChangeM5  float64 `json:"priceChangeM5"`
	PriceChangeH1  float64 `json:"priceChangeH1"`
	PriceChangeH6  float64 `json:"priceChangeH6"`
	PriceChangeH24 float64 `json:"priceChangeH24"`
}

type BigTx struct {
	Tx             string               `json:"tx"`
	TokenAddress   string               `json:"token_address"`
	BlockTimestamp time.Time            `json:"block_timestamp"`
	BlockNumber    uint64               `json:"block_number"`
	Sender         string               `json:"sender"`
	Time           time.Time            `json:"time"`
	ValueInToken   float64              `json:"value_in_token"`
	ValueInUsdt    float64              `json:"value_in_usdt"`
	Price          float64              `json:"price"`
	Movement       string               `json:"movement"`
	Action         SmartMoneyActivities `json:"action"`
}

type TokenBalance struct {
	Address string  `json:"address"`
	Amount  float64 `json:"amount"`
}

type CmcTokens struct {
	UpdatedTime int64          `json:"updated_time"`
	Tokens      []CmcTokenInfo `json:"tokens"`
}

type CmcTokenInfo struct {
	Name                  string   `json:"name"`
	Symbol                string   `json:"symbol"`
	CirculatingSupply     float64  `json:"circulating_supply"`
	TotalSupply           float64  `json:"total_supply"`
	MaxSupply             float64  `json:"max_supply"`
	UsdPrice              float64  `json:"usd_price"`
	MarketCap             float64  `json:"market_cap"`
	Tags                  []string `json:"tags"`
	Volume24H             float64  `json:"volume_24h"`
	FullyDilutedValuation float64  `json:"fully_diluted_valuation"`
	PercentChange1H       float64  `json:"percent_change_1h"`
	PercentChange24H      float64  `json:"percent_change_24h"`
	PercentChange7D       float64  `json:"percent_change_7d"`
}
