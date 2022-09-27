package tracker

import (
	arseeding "github.com/everFinance/arseeding/sdk"
	"github.com/everFinance/turing/store/schema"
	"sync"
	"time"

	"github.com/everFinance/goar"

	"github.com/everFinance/turing/common"
	"github.com/everFinance/turing/store"

	"github.com/everFinance/goar/types"
	"github.com/go-co-op/gocron"
)

const (
	DefaultNodeUrl = "https://arweave.net"
)

// Tracker transactions from Arweave
type Tracker struct {
	arClient     *goar.Client
	arSeedingCli *arseeding.ArSeedCli
	store        *store.Store
	scheduler    *gocron.Scheduler
	subscribeTx  chan common.SubscribeTransaction

	qryTags  []types.Tag // query filter params
	arOwner  string
	isSynced bool

	once sync.Once
}

func initDbConfig(cfg *schema.Config) {
	if len(cfg.DbPath) == 0 {
		cfg.DbPath = schema.BoltDirPath
	}
	if len(cfg.DbFileName) == 0 {
		cfg.DbFileName = schema.TrackerDBFileName
	}
	cfg.Bkt = []string{
		schema.AllSyncedTokenTxBucket,
		schema.ConstantBucket,
	}
}

func New(tags []types.Tag, arNode, arSeedingUrl string, arOwner string, dbConfig schema.Config) *Tracker {

	initDbConfig(&dbConfig)
	var err error
	kv := &store.Store{}
	if dbConfig.UseS3 {
		kv, err = store.NewS3Store(dbConfig)
	} else {
		kv, err = store.NewBoltStore(dbConfig)
	}
	if err != nil {
		panic(err)
	}

	tr := &Tracker{
		arClient:    goar.NewClient(arNode),
		store:       kv,
		scheduler:   gocron.NewScheduler(time.UTC),
		subscribeTx: make(chan common.SubscribeTransaction),
		qryTags:     tags,
		arOwner:     arOwner,
		isSynced:    false,
	}
	if arSeedingUrl != "" {
		tr.arSeedingCli = arseeding.New(arSeedingUrl)
	}
	return tr
}

// Run Tracker, auto load txs from arweave
func (t *Tracker) Run(cursor uint64) {
	log.Debug("start run tracker...", "cursor", cursor)
	go t.once.Do(func() {
		// 1. load store token tx
		now := time.Now()
		err := t.store.LoadSubscribeTxsToStream(cursor, t.subscribeTx)
		if err != nil {
			log.Error("load store subscribe transaction error", "err", err)
			panic(err)
		}
		log.Debug("Tracker t.store.LoadSubscribeTxsToStream() success", "used time(s)", time.Since(now).Seconds())
		// 2. timing get token tx from ar chain
		t.runJobs()
	})
}

func (t *Tracker) Close() {
	t.scheduler.Stop()
}

func (t *Tracker) SubscribeTx() <-chan common.SubscribeTransaction {
	return t.subscribeTx
}

func (t *Tracker) ProcessedArTxId() (string, error) {
	return t.store.GetConstant(schema.LastProcessArTxIdKey)
}

func (t *Tracker) Addr() string {
	return t.arOwner
}

func (t *Tracker) IsSynced() bool {
	return t.isSynced
}
