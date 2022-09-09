package store

// --------------------- common ------------------------------
const (
	boltAllocSize = 8 * 1024 * 1024
	StoreDirPath  = "./bolt"
)

var (
	ConstantBucket = []byte("constant-bucket") // store some constant value
)

// --------------------- about rollup server --------------------
const (
	RollupDBFileName = "rollup.db"
)

var (
	// bucket name
	AllTokenTxBucket = []byte("all-token-transaction-bucket")  // store all token tx
	PoolTxIndex      = []byte("pool-transaction-index-bucket") // store pending token tx

	// key
	LastOnChainTokenTxHashKey = []byte("LastOnChainTokenTxHashKey")
	LastArTxIdKey             = []byte("LastArTxIdKey")
	LastAddPoolTokenTxIdKey   = []byte("LastAddPoolTokenTxIdKey")
)

// --------------------- about tracker server -------------------
const (
	TrackerDBFileName = "tracker.db"
)

var (
	// bucket name
	AllSyncedTokenTxBucket = []byte("all-synced-token-tx-bucket") // store all token tx

	// key
	LastProcessArTxIdKey = []byte("LastProcessArTxId") // process ar tx id
)
