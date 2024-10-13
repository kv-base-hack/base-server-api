package moralis

type Token struct {
	TokenAddress string `json:"token_address"`
}

type Tokens struct {
	Tokens []Token `json:"tokens"`
}

type MoralisToken struct {
	TokenAddress string `json:"associatedTokenAddress"`
	Mint         string `json:"mint"`
	AmountRaw    string `json:"amountRaw"`
	Amount       string `json:"amount"`
	Decimals     string `json:"decimals"`
	Name         string `json:"name"`
	Symbol       string `json:"symbol"`
}
