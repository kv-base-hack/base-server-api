package db

type DB interface {
	GetMaxBlockNumber(table string) (int64, error)
	GetSolTrades(fromBlock int64, limit uint64) ([]SolanaTradelogDB, error)
	GetSolTransfer(fromBlock int64, limit uint64) ([]SolanaTransferLogDb, error)
}
