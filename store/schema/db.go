package schema

// --------------------- common ------------------------------
const (
	BoltAllocSize = 8 * 1024 * 1024
	BoltDirPath   = "./bolt"
)

var (
	ConstantBucket = "constant-bucket" // store some constant value

	// key
	SeqNum = "sequence-number"
)

// --------------------- about rollup server --------------------
const (
	RollupDBFileName = "rollup.db"
)

var (
	// bucket name
	AllTokenTxBucket = "all-token-transaction-bucket"  // store all token tx
	PoolTxIndex      = "pool-transaction-index-bucket" // store pending token tx

	// key
	LastOnChainTokenTxHashKey = "LastOnChainTokenTxHashKey"
	LastArTxIdKey             = "LastArTxIdKey"
	LastAddPoolTokenTxIdKey   = "LastAddPoolTokenTxIdKey"
)

// --------------------- about tracker server -------------------
const (
	TrackerDBFileName = "tracker.db"
)

var (
	// bucket name
	AllSyncedTokenTxBucket = "all-synced-token-tx-bucket" // store all token tx

	// key
	LastProcessArTxIdKey = "LastProcessArTxId" // process ar tx id
)

var (
	AllBkt = []string{
		ConstantBucket,
		AllTokenTxBucket,
		PoolTxIndex,
		AllSyncedTokenTxBucket,
	}
)

type Config struct {
	Bkt []string
	// use s3 or 4ever
	UseS3     bool
	Use4EVER  bool
	AccKey    string
	SecretKey string
	Region    string
	BktPrefix string

	// bolt
	DbPath     string
	DbFileName string
}
