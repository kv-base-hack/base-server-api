package db

import (
	"time"

	"github.com/kv-base-hack/base-server-api/common"
)

type SolanaTradelogDB struct {
	BlockTimestamp time.Time `db:"block_timestamp"`
	BlockNumber    uint64    `db:"block_number"`
	TxHash         string    `db:"tx_hash"`
	Sender         string    `db:"sender"`

	TokenInAddress  string  `db:"token_in_address"`
	TokenInAmount   float64 `db:"token_in_amount"`
	TokenInUsdtRate float64 `db:"token_in_usdt_rate"`

	TokenOutAddress  string  `db:"token_out_address"`
	TokenOutAmount   float64 `db:"token_out_amount"`
	TokenOutUsdtRate float64 `db:"token_out_usdt_rate"`

	SolUsdtRate float64 `db:"sol_usdt_rate"`

	Created time.Time `db:"created"`
}

func (t SolanaTradelogDB) Convert() common.Tradelog {
	return common.Tradelog{
		BlockTimestamp: t.BlockTimestamp,
		BlockNumber:    t.BlockNumber,
		TxHash:         t.TxHash,
		Sender:         t.Sender,

		TokenInAddress:  t.TokenInAddress,
		TokenInAmount:   t.TokenInAmount,
		TokenInUsdtRate: t.TokenInUsdtRate,

		TokenOutAddress:     t.TokenOutAddress,
		TokenOutAmount:      t.TokenOutAmount,
		TokenOutUsdtRate:    t.TokenOutUsdtRate,
		NativeTokenUsdtRate: t.SolUsdtRate,
	}
}

type SolanaTransferLogDb struct {
	BlockTimestamp time.Time `db:"block_timestamp"`
	BlockNumber    uint64    `db:"block_number"`
	TxHash         string    `db:"tx_hash"`
	FromAddress    string    `db:"from_address"`
	ToAddress      string    `db:"to_address"`

	TokenAddress string  `db:"token_address"`
	TokenAmount  float64 `db:"token_amount"`

	IsCexIn bool      `db:"is_cex_in"`
	Created time.Time `db:"created"`
}

func (e SolanaTransferLogDb) Convert() common.Transferlog {
	return common.Transferlog{
		BlockTimestamp: e.BlockTimestamp,
		BlockNumber:    e.BlockNumber,
		TxHash:         e.TxHash,
		FromAddress:    e.FromAddress,
		ToAddress:      e.ToAddress,
		TokenAddress:   e.TokenAddress,
		TokenAmount:    e.TokenAmount,
		IsCexIn:        e.IsCexIn,
	}
}
