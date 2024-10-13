package db

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // sql driver name: "postgres"
)

const SolanaTradeTable = "solana_trade_logs"
const SolanaTransferTable = "solana_transfer_logs"

type Postgres struct {
	db *sqlx.DB
}

func NewPostgres(db *sqlx.DB) *Postgres {
	return &Postgres{
		db: db,
	}
}

func (pg *Postgres) GetMaxBlockNumber(table string) (int64, error) {
	query := sq.Select("max(block_number)").From(table)

	var maxBlock int64
	sql, args, err := query.ToSql()
	if err != nil {
		return 0, err
	}
	err = pg.db.Get(&maxBlock, sql, args...)
	if err != nil {
		return 0, err
	}

	return maxBlock, nil
}

func (pg *Postgres) GetSolTrades(fromBlock int64, limit uint64) ([]SolanaTradelogDB, error) {
	query := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
		Select("block_timestamp", "block_number", "tx_hash", "sender",
			"token_in_address", "token_in_amount", "token_in_usdt_rate",
			"token_out_address", "token_out_amount", "token_out_usdt_rate",
			"sol_usdt_rate",
		).
		From(SolanaTradeTable).OrderBy("block_number").Limit(limit).Where(sq.GtOrEq{"block_number": fromBlock})

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}
	var logs []SolanaTradelogDB

	err = pg.db.Select(&logs, sql, args...)
	if err != nil {
		return nil, err
	}

	return logs, nil
}

func (pg *Postgres) GetSolTransfer(fromBlock int64, limit uint64) ([]SolanaTransferLogDb, error) {
	query := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
		Select("block_timestamp", "block_number", "tx_hash",
			"from_address", "to_address",
			"token_address", "token_amount",
			"is_cex_in",
		).From(SolanaTransferTable).OrderBy("block_number").Limit(limit).Where(sq.GtOrEq{"block_number": fromBlock})

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	var logs []SolanaTransferLogDb
	err = pg.db.Select(&logs, sql, args...)
	if err != nil {
		return nil, err
	}

	return logs, nil
}
