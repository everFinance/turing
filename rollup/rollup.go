package rollup

import (
	"errors"
	"fmt"
	types "github.com/everFinance/turing/common"
	"github.com/everFinance/turing/tracker"
	"github.com/getsentry/sentry-go"
	"sync"

	"time"

	"github.com/everFinance/goar"
	arTypes "github.com/everFinance/goar/types"
	"github.com/everFinance/goar/utils"

	arseeding "github.com/everFinance/arseeding/sdk"
	"github.com/everFinance/turing/rollup/txpool"
	"github.com/everFinance/turing/store"
)

var log = types.NewLog(types.RollupLogMode)

type Rollup struct {
	store  *store.Store
	txPool *txpool.TxPool
	txChan chan []byte

	lastAddPoolTokenTxHash string
	lastArTxId             string // last ar tx hash
	watcherArTxId          string // onchain watcher arId; use to monitor rollup status
	curRollupTxNum         int    // current onChain token txs num; use to monitor rollup status
	locker                 sync.Mutex

	arWallet     *goar.Wallet
	arSeedingCli *arseeding.ArSeedCli
	tags         []arTypes.Tag
	txSpeed      int64 // rollup ar tx speed, default 50; means speed 50%

	stop chan struct{}
}

func initTxPool(kv *store.Store, lastPendTokenTxHash string) (*txpool.TxPool, error) {
	// init tx pool
	pool := txpool.NewTxPool(lastPendTokenTxHash)
	// load pool tx from db
	poolTxs, err := kv.LoadPoolTxFromDb()
	if err != nil {
		return nil, err
	}

	// sort poolTxs
	resTxs, err := types.TxsSort(poolTxs, lastPendTokenTxHash)
	if err != nil {
		return nil, err
	}
	// add to pool
	for _, ttx := range resTxs {
		if err := pool.AddTx(ttx); err != nil {
			return nil, err
		}
	}
	return pool, nil
}

// New suggestLastArTxId use rollup last on chain tx id, if server bolt.db is clear, restart server need config this args
func New(suggestLastArTxId string, arNode, arSeedingUrl string, arWalletKeyPath string, owner string, tags []arTypes.Tag, dbDirPath string) *Rollup {
	if len(dbDirPath) == 0 {
		dbDirPath = store.StoreDirPath
	}
	kv, err := store.NewKvStore(
		dbDirPath, store.RollupDBFileName,
		store.AllTokenTxBucket, store.PoolTxIndex, store.ConstantBucket,
	)
	if err != nil {
		panic(err)
	}

	// get constant from kv
	lastArTxId, err := kv.GetConstant(store.LastArTxIdKey)
	if err != nil {
		panic(err)
	}
	// is lastArTxId is "", use suggestLastArTxId
	if lastArTxId == "" {
		lastArTxId = suggestLastArTxId
		// verify last ar tx id is correct
		if err := verifyLastArTxId(lastArTxId, arNode, owner, tags); err != nil {
			panic(err)
		}
	}

	lastOnChainTokenTxHash, err := kv.GetConstant(store.LastOnChainTokenTxHashKey)
	if err != nil {
		panic(err)
	}
	lastAddPoolTokenTxHash, err := kv.GetConstant(store.LastAddPoolTokenTxIdKey)
	if err != nil {
		panic(err)
	}
	log.Info("store constant load", "lastOnChainTokenTxHash", lastOnChainTokenTxHash,
		"lastAddPoolTokenTxHash", lastAddPoolTokenTxHash, "lastArTxId", lastArTxId)
	// init tx pool
	pool, err := initTxPool(kv, lastOnChainTokenTxHash)
	if err != nil {
		panic(err)
	}

	wlt, err := goar.NewWalletFromPath(arWalletKeyPath, arNode)
	if err != nil {
		panic(err)
	}
	rp := &Rollup{
		store:                  kv,
		txChan:                 make(chan []byte),
		lastAddPoolTokenTxHash: lastAddPoolTokenTxHash,
		txPool:                 pool,
		lastArTxId:             lastArTxId,
		arWallet:               wlt,
		tags:                   tags,
		txSpeed:                50,
		stop:                   make(chan struct{}),
		locker:                 sync.Mutex{},
	}
	if arSeedingUrl != "" {
		rp.arSeedingCli = arseeding.New(arSeedingUrl)
	}
	return rp
}

func verifyLastArTxId(lastArTxId string, arNode string, owner string, tags []arTypes.Tag) error {
	if len(lastArTxId) == 0 {
		return nil
	}
	log.Debug("start verify lastArTxId", "lastArTxId", lastArTxId)
	arClient := goar.NewClient(arNode)
	tx, err := arClient.GetTransactionByID(lastArTxId)
	if err != nil {
		log.Error("arClient.GetTransactionByID(lastArTxId)", "err", err, "lastArTxId", lastArTxId)
		return errors.New("lastArId incorrect")
	}
	// verify owner
	from, err := utils.OwnerToAddress(tx.Owner)
	if err != nil {
		log.Error("ownerToAddress", "err", err)
		return err
	}
	if from != owner {
		return errors.New("lastArTxId not belong to correct owner")
	}
	// verify tags
	tagMap := make(map[string]struct{})
	for _, tag := range tx.Tags {
		tagMap[tag.Name] = struct{}{}
	}

	for _, tt := range utils.TagsEncode(tags) {
		if _, ok := tagMap[tt.Name]; !ok {
			return errors.New("lastArTxId get tx tags incorrect")
		}
	}
	return nil
}

// Run
func (rol *Rollup) Run(timeInterval time.Duration, maxOfRollup int) {

	select {
	case <-rol.stop:
	default:
	}

	// listen tx and add to pool
	go rol.listenTokenTxToPool()
	// seal token tx on chain
	go rol.sealTxOnChain(timeInterval, maxOfRollup)
}

// Close
func (rol *Rollup) Close() {
	close(rol.stop)
}

//  add token tx
func (rol *Rollup) AddTx() chan<- []byte {
	return rol.txChan
}

func (rol *Rollup) listenTokenTxToPool() {
	for {
		select {
		case <-rol.stop:
			return
		case tx := <-rol.txChan:
			parentHash := rol.lastAddPoolTokenTxHash
			txHash := types.HashBytes(tx)

			poolTx := &types.Transaction{
				TxData:   tx,
				TxId:     txHash,
				ParentId: parentHash,
			}
			// put tx to all tx bucket
			if err := rol.store.PutTokenTx(poolTx); err != nil {
				panic(err)
			}
			// put tx id to pool bucket
			if err := rol.store.PutPoolTokenTxId(poolTx.TxId); err != nil {
				panic(err)
			}

			// add to pool
			err := rol.txPool.AddTx(poolTx)
			if err != nil {
				log.Error("Add tx to tx pool error", "error", err)
				panic(err)
			} else {
				// update LastAddPoolTokenTxIdKey
				if err := rol.store.UpdateConstant(store.LastAddPoolTokenTxIdKey, []byte(txHash)); err != nil {
					panic(err)
				}
				rol.lastAddPoolTokenTxHash = txHash
				log.Info("success add one tokenTx to pool", "parentHash", parentHash, "tokenTxHash", txHash)
			}
		}
	}
}

func (rol *Rollup) sealTxOnChain(timeInterval time.Duration, maxOfRollup int) {
	ticker := time.NewTicker(timeInterval)
	for {
		select {
		case <-rol.stop:
			return
		case <-ticker.C:
			log.Info("seal tx on chain...")
			// load pending txs
			if rol.txPool.PendingTxNum() == 0 {
				log.Debug("tx pool pending cache is null")
				continue
			}
			if maxOfRollup > 0 {
				// peek txs
				txs, index := rol.txPool.PeekPendingTxs(maxOfRollup)
				if len(txs) == 0 || index < 0 {
					continue
				}
				log.Info("peek pool txs", "tx number", index+1)
				// txs on chain process
				data, err := txs.MarshalFromOnChain()
				if err != nil {
					panic(err)
				}
				arId, err := rol.txOnChain(data)
				if err != nil {
					log.Error("tx on chain failed", "last tx id", rol.lastArTxId, "error", err)
					break
				}
				log.Info("send ar tx on chain", "arId", arId)
				rol.setWatcherInfo(arId, len(txs))

				// watcher tx
				txStatus := arTxWatcher(rol.arWallet.Client, rol.arSeedingCli, arId)
				if !txStatus {
					// stop on chain
					log.Error("watcher tx status failed", "last tx id", rol.lastArTxId, "status", txStatus)
					break
				}
				rol.setWatcherInfo("", 0)

				// on chain success and modify some status
				// 1. modify db constants 'lastArTxId'
				if err := rol.store.UpdateConstant(store.LastArTxIdKey, []byte(arId)); err != nil {
					log.Error("modify lastArTxIdKey error", "error", err, "newValue", arId)
					panic(err)
				}
				// 2. modify db  constants 'lastOnChainTokenTxHash'
				if err := rol.store.UpdateConstant(store.LastOnChainTokenTxHashKey, []byte(txs[len(txs)-1].TxId)); err != nil {
					log.Error("modify lastOnChainTokenTxHashKey error", "error", err, "newValue", txs[len(txs)-1].TxId)
					panic(err)
				}
				// 3. delete poolTxIndex bucket txs index
				if err := rol.store.BatchDelPoolTokenTxId(txs); err != nil {
					panic(err)
				}
				// 4. change lastTxHash
				rol.setLastArTxId(arId)
				// 5. pop txs from tx pool
				rol.txPool.PopPendingTxs(index + 1)
				log.Info("seal pool token tx on chain success", "arId", arId)
				// log.Debug("view this token tx info ↓ ↓ ↓")
				// for index, tx := range txs {
				// 	log.Debug("token tx ", "index", index, "txId", tx.TxId, "parentId", tx.ParentId, "txData", tx.TxData)
				// }
			}
		}
	}
}

func (rol *Rollup) txOnChain(data []byte) (string, error) {
	var (
		arErr error
		arTx  arTypes.Transaction
	)
	parentIdTag := arTypes.Tag{
		Name:  types.ParentTag,
		Value: rol.GetLastArTxId(),
	}
	tags := append(rol.tags, parentIdTag)
	// retry 5 times
	for i := 1; i <= 5; i++ {
		speed := rol.GetTxSpeed()
		log.Debug("rollup tx speed", "speed", speed)
		arTx, arErr = rol.arWallet.SendDataSpeedUp(data, tags, speed)
		if arErr != nil {
			// retry
			num := time.Duration(i * 2)
			time.Sleep(time.Second * num)
			log.Warn("Retry sendData tx to ar network", " time", i)
			continue
		} else {
			break
		}
	}

	// support arTx is Transaction{}
	rol.putToArSeeding(arTx)

	if arErr == nil {
		return arTx.ID, nil
	}

	return "", arErr
}

func (rol *Rollup) putToArSeeding(arTx arTypes.Transaction) {
	if rol.arSeedingCli != nil {
		// submit tx
		if err := rol.arSeedingCli.SubmitTx(arTx); err != nil {
			log.Warn("rol.arSeedingCli.SubmitTx(arTx)", "err", err, "arId", arTx.ID)
			return
		}
		// broadcast tx data
		if err := rol.arSeedingCli.BroadcastTxData(arTx.ID); err != nil {
			log.Warn("arSeedingCli.BroadcastTxData(arTx.ID)", "err", err, "arId", arTx.ID)
		}
	}
}

func (r *Rollup) GetTxSpeed() int64 {
	if r.txSpeed <= 0 {
		return 50 // default 50
	}
	return r.txSpeed
}

func (r *Rollup) SetTxSpeed(newSpeed int64) {
	r.txSpeed = newSpeed
}

func (r *Rollup) GetPendingTxs() types.Transactions {
	pendingNum := r.txPool.PendingTxNum()
	txs, _ := r.txPool.PeekPendingTxs(pendingNum)
	return txs
}

func (r *Rollup) setLastArTxId(lastArId string) {
	r.locker.Lock()
	defer r.locker.Unlock()
	r.lastArTxId = lastArId
}

func (r *Rollup) GetLastArTxId() string {
	r.locker.Lock()
	defer r.locker.Unlock()
	return r.lastArTxId
}

func (r *Rollup) GetLastOnChainData() ([]byte, error) {
	tx, err := r.store.GetLastOnChainTx()
	if err != nil {
		return nil, err
	}
	return tx.TxData, nil
}

func (r *Rollup) setWatcherInfo(arId string, txNum int) {
	r.locker.Lock()
	defer r.locker.Unlock()
	r.watcherArTxId = arId
	r.curRollupTxNum = txNum
}

func (r *Rollup) GetWatcherInfo() (arId string, txNum int) {
	r.locker.Lock()
	defer r.locker.Unlock()
	arId = r.watcherArTxId
	txNum = r.curRollupTxNum
	return
}

func (r *Rollup) GetPendingTxNum() int {
	return r.txPool.PendingTxNum()
}

func (r *Rollup) Addr() string {
	return r.arWallet.Signer.Address
}

// watcher tx status from blockchain
func arTxWatcher(arCli *goar.Client, arSeedingCli *arseeding.ArSeedCli, arTxHash string) bool {
	// loop watcher on chain tx status
	// total time 59 minute
	tmp := 0
	for i := 1; i <= 21; i++ {
		// sleep
		num := 60 + i*10
		time.Sleep(time.Second * time.Duration(num))

		tmp += num
		log.Debug("watcher tx sleep time", "wait total time(s)", tmp)
		log.Debug("retry get tx status", "txHash", arTxHash)

		// watcher on-chain tx
		status, err := arCli.GetTransactionStatus(arTxHash)
		if err != nil {
			if err.Error() == goar.ErrPendingTx.Error() {
				log.Debug("tx is pending", "txHash", arTxHash)
			} else {
				log.Error("get rollup tx status", "err", err, "txHash", arTxHash)
			}
			continue
		}

		// when err is nil
		// 1. confirms block height must >= 3
		if status.NumberOfConfirmations < 3 {
			log.Debug("roll up tx must more than 2 block confirms", "txHash", arTxHash, "currentConfirms", status.NumberOfConfirmations)
			continue
		}

		// 2. must get tx data for chain
		data, err := tracker.FetchAndVerifyTxData(arTxHash, arCli, arSeedingCli)
		if err == nil && len(data) > 0 {
			// on chain success
			log.Info("tx on chain success", "txHash", arTxHash)
			return true
		}
		log.Error("arTxWatcher get tx data error", "txHash", arTxHash, "err", err)
	}

	sentry.CaptureException(fmt.Errorf("rollup arTxWatcher failed; txId: %s", arTxHash))
	return false
}
