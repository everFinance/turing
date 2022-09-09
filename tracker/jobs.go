package tracker

import (
	"github.com/everFinance/turing/common"
	types "github.com/everFinance/turing/common"
	"github.com/everFinance/turing/store"
	"time"
)

var log = common.NewLog(types.TrackerLogMode)

func (t *Tracker) runJobs() {
	t.scheduler.Every(2).Minutes().SingletonMode().Do(t.jobTxsPullFromChain)

	t.scheduler.StartAsync()
}

func (t *Tracker) jobTxsPullFromChain() {
	log.Info("job txs pull running...")
	defer func() {
		log.Info("job txs pull done")
		time.Sleep(5 * time.Second)
		t.isSynced = true
	}()

	// get all unprocessed txs
	// null value is return ""
	lastTxID, err := t.store.GetConstant(store.LastProcessArTxIdKey)
	if err != nil {
		log.Error("can not get lastTxID", "err", err)
		return
	}

	ids := reverseIDs(MustFetchTxIdsByNativeMethod(t.qryTags, lastTxID, t.arClient))

	// process txs
	arTxsCounts := 0
	bizTxsCounts := 0

	for _, id := range ids {
		owner, timestamp, height, data := t.mustFetchDataByID(id)

		// 1. process
		// Only execute transactions where the owner and arOwner are the same or where the arOwner is ""
		num := 0
		if t.arOwner == "" || t.arOwner == owner {
			num = t.processTxs(id, owner, timestamp, height, data)
		}

		// 2. update last process ar tx id to store
		if err := t.store.UpdateLastProcessArTxId(id); err != nil {
			log.Error("update last process ar tx id store error", "err", err)
			panic(err)
		}

		log.Info("process txs", "arTxId", id, "number", num) // If num = 0 means the arTx is not part of the transaction on the arOwner chain
		bizTxsCounts += num
		arTxsCounts++
	}

	log.Info("job txs processed", "arTxCounts", arTxsCounts, "bizTxsCount", bizTxsCounts)
}

// processTxs
func (t *Tracker) processTxs(id, owner string, timestamp, height int64, data []byte) (counts int) {
	log.Debug("start process token txs", "ar tx id ", id)
	counts = 0
	bizTxs := types.Transactions{}
	if err := bizTxs.Unmarshal(data); err != nil {
		log.Error("can not unmarshal txs", "data", data, "err", err)
		panic(err)
	}

	for _, bizTx := range bizTxs {
		subTx := types.SubscribeTransaction{
			ID:        id,
			Owner:     owner,
			Data:      bizTx.TxData,
			Timestamp: timestamp,
			Height:    height,
		}

		// 1. store tx info to kv
		cursorId, err := t.store.PutSubscribeTx(subTx)
		if err != nil {
			panic(err)
		}
		subTx.CursorId = cursorId
		// 2. subscribe tx
		t.subscribeTx <- subTx
		counts++
	}
	return
}

func reverseIDs(ids []string) (rIDs []string) {
	for i := len(ids) - 1; i >= 0; i-- {
		rIDs = append(rIDs, ids[i])
	}
	return
}
