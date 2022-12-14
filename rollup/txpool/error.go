package txpool

import "errors"

var (
	ErrNotFoundBlockCache = errors.New("not found block in TxGuard'Cache")
	ErrTimeBucketTime     = errors.New("block timestamp should be greater than bucket base time")
	ErrDifferentGenesis   = errors.New("found different genesis block")
	ErrInvalidTx          = errors.New("the transaction is broken")
	ErrTxIsExist          = errors.New("the transaction is exist in txPool")
	ErrTxPoolExtendFail   = errors.New("txPool extend fail")
	ErrInvalidBaseTime    = errors.New("invalid stable block time")
	ErrPushShard          = errors.New("can not push shard")
)
