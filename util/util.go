package util

import "strings"

var quotes = []string{
	"0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee", // eth
	"0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2", // weth
	"0xdac17f958d2ee523a2206206994597c13d831ec7", // usdt
	"0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48", // usdc
	"0x6b175474e89094c44da98b954eedeac495271d0f", // dai
	"0x853d955acef822db058eb8505911ed77f175b99e", // fxs
}

func IsQuote(tokenAddress string) bool {
	for _, q := range quotes {
		if strings.EqualFold(q, tokenAddress) {
			return true
		}
	}
	return false
}
