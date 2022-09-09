package txpool

import (
	types "github.com/everFinance/turing/common"
	"sync"
)

var log = types.NewLog(types.TxPoolLogMode)

type txSortShard struct {
	headShard string // first tx's parent id
	tailShard string // last tx id
	cache     types.Transactions
	sync.Mutex
}

func NewTxSortShard() *txSortShard {
	return &txSortShard{
		headShard: "",
		tailShard: "",
		cache:     make(types.Transactions, 0),
		Mutex:     sync.Mutex{},
	}
}

func (shard *txSortShard) push(tx *types.Transaction) error {
	if len(shard.cache) != 0 && tx.ParentId != shard.tailShard {
		return ErrPushShard
	}
	shard.pushTxs(types.Transactions{tx})
	return nil
}

func (shard *txSortShard) canPush(tx *types.Transaction) bool {
	shard.Lock()
	defer shard.Unlock()
	if len(shard.cache) == 0 {
		return true
	}
	return shard.tailShard == tx.ParentId
}

func (shard *txSortShard) clean() {
	shard.tailShard = ""
	shard.headShard = ""
	shard.cache = make(types.Transactions, 0)
	shard.Mutex = sync.Mutex{}
}

// popTxs
func (shard *txSortShard) popTxs(txNum int) types.Transactions {
	shard.Lock()
	defer shard.Unlock()
	if len(shard.cache) == 0 || txNum == 0 {
		return types.Transactions{}
	}
	if len(shard.cache) <= txNum {
		// all pop
		pop := make(types.Transactions, len(shard.cache))
		copy(pop, shard.cache)
		employShard := NewTxSortShard()
		shard.headShard = employShard.headShard
		shard.tailShard = employShard.tailShard
		shard.cache = employShard.cache
		return pop
	} else {
		pop := make(types.Transactions, txNum)
		copy(pop, shard.cache[:txNum])
		shard.headShard = shard.cache[txNum].ParentId
		shard.cache = shard.cache[txNum:]
		return pop
	}
}

func (shard *txSortShard) peekTxs(txNum int) (types.Transactions, int) {
	shard.Lock()
	defer shard.Unlock()
	if len(shard.cache) == 0 || txNum == 0 {
		return types.Transactions{}, -1
	}
	if len(shard.cache) <= txNum {
		// all pop
		pop := make(types.Transactions, len(shard.cache))
		copy(pop, shard.cache)
		return pop, len(shard.cache) - 1
	} else {
		pop := make(types.Transactions, txNum)
		copy(pop, shard.cache[:txNum])

		return pop, txNum - 1
	}
}

func (shard *txSortShard) pushTxs(txs types.Transactions) {
	if len(txs) == 0 {
		return
	}
	shard.Lock()
	if len(shard.cache) == 0 {
		shard.headShard = txs[0].ParentId
	}
	shard.cache = append(shard.cache, txs...)
	shard.tailShard = txs[len(txs)-1].TxId
	shard.Unlock()
}

type TxPool struct {
	lastPendTxId string                  // last pending tx Id; if value == "", means first run server
	pending      *txSortShard            // can be sealed txs at next block
	queueMap     map[string]*txSortShard // can be sealed in the future; key is shard head parentId
	txMap        map[string]struct{}     // key is txId
	RW           sync.RWMutex
}

func NewTxPool(lastPendTokenTxHash string) *TxPool {
	pool := &TxPool{
		lastPendTxId: lastPendTokenTxHash,
		pending:      NewTxSortShard(),
		queueMap:     make(map[string]*txSortShard),
		txMap:        make(map[string]struct{}),
		RW:           sync.RWMutex{},
	}
	return pool
}

func (pool *TxPool) PendingTxNum() int {
	return len(pool.pending.cache)
}

func (pool *TxPool) TotalCacheTxNum() int {
	return len(pool.txMap)
}

func (pool *TxPool) PopPendingTxs(txNum int) types.Transactions {
	pool.RW.Lock()
	defer pool.RW.Unlock()
	txs := pool.pending.popTxs(txNum)
	if len(txs) > 0 {
		pool.lastPendTxId = txs[len(txs)-1].TxId
	}
	return txs
}

func (pool *TxPool) PeekPendingTxs(txNum int) (types.Transactions, int) {
	pool.RW.Lock()
	defer pool.RW.Unlock()
	return pool.pending.peekTxs(txNum)
}

// AddTx
func (pool *TxPool) AddTx(tx *types.Transaction) error {
	pool.RW.Lock()
	defer pool.RW.Unlock()
	if tx == nil {
		return ErrInvalidTx
	}
	err := pool.addTx(tx)
	if err != nil {
		return err
	}
	// merge queueMap and pending shard
	pool.mergeShard()
	return nil
}

// mergeShard
func (pool *TxPool) mergeShard() {
	// merge queueMap in the shards
	if len(pool.queueMap) > 1 {
	Loop:
		headShardArr := make([]string, 0, len(pool.queueMap))
		for key, _ := range pool.queueMap {
			headShardArr = append(headShardArr, key)
		}

		// loop compare that shards can be merge
		for i := 0; i < len(headShardArr); i++ {
			for j := 0; j < len(headShardArr); j++ {
				if i == j {
					continue
				}
				if pool.queueMap[headShardArr[i]].tailShard == pool.queueMap[headShardArr[j]].headShard {
					// merge
					newTxShard := NewTxSortShard()
					newTxShard.headShard = pool.queueMap[headShardArr[i]].headShard
					newTxShard.tailShard = pool.queueMap[headShardArr[j]].tailShard
					newTxShard.cache = append(pool.queueMap[headShardArr[i]].cache, pool.queueMap[headShardArr[j]].cache...)
					// add to be new shard
					pool.queueMap[newTxShard.headShard] = newTxShard
					// delete old shard， note can not delete 'pool.queueMap[headShardArr[i]]' shard
					delete(pool.queueMap, pool.queueMap[headShardArr[j]].headShard)
					// only need to find one pair can merge shard and goto again
					goto Loop
				}
			}
		}

		// second merge pending and queueMap
		if len(pool.pending.cache) == 0 {
			// compare lastPendTxId and queueMap
			if txShard, ok := pool.queueMap[pool.lastPendTxId]; ok {
				// become pending queue
				pool.pending.pushTxs(txShard.cache)
				// delete shard
				delete(pool.queueMap, pool.lastPendTxId)
			}
		} else {
			// compare pending shard and queueMap shard
			pendingTailShard := pool.pending.tailShard
			if txShard, ok := pool.queueMap[pendingTailShard]; ok {
				// merged
				pool.pending.pushTxs(txShard.cache)
				// delete queue shard
				delete(pool.queueMap, txShard.headShard)
			}
		}
	}
}

// addTx
func (pool *TxPool) addTx(tx *types.Transaction) (err error) {
	if tx == nil {
		err = ErrInvalidTx
		return
	}
	if pool.isExist(tx) {
		err = ErrTxIsExist
		return
	}
	defer func() {
		if err == nil {
			pool.txMap[tx.TxId] = struct{}{}
		}
	}()

	// add first tx
	if pool.lastPendTxId == "" && len(pool.pending.cache) == 0 {
		// push to pending queue
		err = pool.pending.push(tx)
		if err != nil {
			log.Error("push tx to shard error", "err", err)
			return err
		}
		return nil
	}

	// not first tx
	// check if can be push pending queue
	if pool.lastPendTxId == tx.ParentId {
		// pending queue first tx
		pool.pending.clean()
		_ = pool.pending.push(tx)
		return nil
	} else {
		// check if can push pending queue last
		if len(pool.pending.cache) > 0 && pool.pending.canPush(tx) {
			err = pool.pending.push(tx)
			return err
		}
	}

	// check is it possible to push queue
	var tag = false
	for _, que := range pool.queueMap {
		if que.canPush(tx) {
			err = que.push(tx)
			if err != nil {
				log.Error("push tx error", "err", err)
				return err
			}
			tag = true
			break
		}
	}

	// can not push exist queue shard, need create new queue shard
	if !tag {
		// need create new shard
		newQue := NewTxSortShard()
		_ = newQue.push(tx)
		pool.queueMap[tx.ParentId] = newQue
	}

	return nil
}

func (pool *TxPool) isExist(tx *types.Transaction) bool {
	_, ok := pool.txMap[tx.TxId]
	return ok
}
