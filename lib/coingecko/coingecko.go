package coingecko

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const providerName = "coingecko"

const (
	timeLayout         = "02-01-2006"
	currentEndpoint    = "%s/coins/%s"
	historicalEndpoint = "%s/coins/%s/history"
)

// CoinGecko is the CoinGecko implementation of Provider. The
// precision of CoinGecko provider is up to day.
type CoinGecko struct {
	client  *http.Client
	baseURL string
}

var treding = "search/trending"

// New creates a new CoinGecko instance.
func NewCoinGecko() *CoinGecko {
	const (
		defaultTimeout = time.Second * 10
		baseURL        = "https://api.coingecko.com/api/v3"
	)
	client := &http.Client{
		Timeout: defaultTimeout,
	}
	return &CoinGecko{
		client:  client,
		baseURL: baseURL,
	}
}

func (cg *CoinGecko) GetTrending() (CoingeckoTrending, error) {
	url := cg.baseURL + "/" + treding

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return CoingeckoTrending{}, err
	}
	req.Header.Add("Accept", "application/json")
	// q := req.URL.Query()
	// q.Add("date", timestamp.UTC().Format(timeLayout))
	// req.URL.RawQuery = q.Encode()
	rsp, err := cg.client.Do(req)
	if err != nil {
		return CoingeckoTrending{}, err
	}
	defer rsp.Body.Close()

	if rsp.StatusCode != http.StatusOK {
		return CoingeckoTrending{}, fmt.Errorf("unexpected status code: %s", rsp.Status)
	}
	respBody, err := io.ReadAll(rsp.Body)
	if err != nil {
		return CoingeckoTrending{}, err
	}

	var coins CoingeckoTrending
	if err := json.Unmarshal(respBody, &coins); err != nil {
		return CoingeckoTrending{}, err
	}
	return coins, nil
}
