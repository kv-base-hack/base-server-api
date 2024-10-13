package coingecko

type CoingeckoTrending struct {
	Coins []CoingeckoCoin `json:"coins"`
}

type CoingeckoCoin struct {
	Item Item `json:"item"`
}

type Item struct {
	// CoinID uint          `json:"coin_id"`
	Name   string        `json:"name"`
	Symbol string        `json:"symbol"`
	Thumb  string        `json:"thumb"`
	Small  string        `json:"small"`
	Data   CoingeckoData `json:"data"`
}

type CoingeckoData struct {
	Price                    float64     `json:"price"`
	MarketCap                string      `json:"market_cap"`
	TotalVolume              string      `json:"total_volume"`
	PriceChangePercentage24h PriceChange `json:"price_change_percentage_24h"`
}

type PriceChange struct {
	Eur float64 `json:"eur"`
}
