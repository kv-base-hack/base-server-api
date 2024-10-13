package moralis

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type MoralisClient struct {
	url string
	key string
}

func NewMoralisClient(url string, key string) *MoralisClient {
	return &MoralisClient{
		url: url,
		key: key,
	}
}

func (c *MoralisClient) GetUserBalance(userAddress string) ([]MoralisToken, error) {
	url := c.url + "/" + userAddress + "/tokens"
	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Add("Accept", "application/json")
	req.Header.Add("X-API-Key", c.key)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	fmt.Println(string(body))

	var tokenBalances []MoralisToken
	if err := json.Unmarshal(body, &tokenBalances); err != nil {
		return nil, err
	}
	return tokenBalances, nil
}
