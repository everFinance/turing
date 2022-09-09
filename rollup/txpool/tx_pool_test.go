package txpool

import (
	"github.com/everFinance/turing/common"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

func TestTxPool_mergeShard(t *testing.T) {
	pool := TxPool{
		lastPendTxId: "0",
		pending:      NewTxSortShard(),
		queueMap:     make(map[string]*txSortShard),
		txMap:        make(map[string]struct{}),
		RW:           sync.RWMutex{},
	}

	shard01 := &txSortShard{
		headShard: "0",
		tailShard: "1",
		cache: common.Transactions{&common.Transaction{
			TxData:   []byte("aaa"),
			TxId:     "1",
			ParentId: "0",
		}},
		Mutex: sync.Mutex{},
	}

	shard02 := &txSortShard{
		headShard: "1",
		tailShard: "2",
		cache: common.Transactions{&common.Transaction{
			TxData:   []byte("bbb"),
			TxId:     "2",
			ParentId: "1",
		}},
		Mutex: sync.Mutex{},
	}

	shard03 := &txSortShard{
		headShard: "2",
		tailShard: "3",
		cache: common.Transactions{&common.Transaction{
			TxData:   []byte("ccc"),
			TxId:     "3",
			ParentId: "2",
		}},
		Mutex: sync.Mutex{},
	}

	shard04 := &txSortShard{
		headShard: "4",
		tailShard: "5",
		cache: common.Transactions{&common.Transaction{
			TxData:   []byte("ccc"),
			TxId:     "5",
			ParentId: "4",
		}},
		Mutex: sync.Mutex{},
	}

	pool.queueMap[shard03.headShard] = shard03
	pool.queueMap[shard02.headShard] = shard02
	pool.queueMap[shard04.headShard] = shard04
	pool.queueMap[shard01.headShard] = shard01

	// mergeShard
	pool.mergeShard()

	//  pending pool
	assert.Equal(t, "0", pool.pending.headShard)
	assert.Equal(t, "3", pool.pending.tailShard)
	assert.Equal(t, 3, len(pool.pending.cache))
	assert.Equal(t, shard01.cache[0], pool.pending.cache[0])
	assert.Equal(t, shard02.cache[0], pool.pending.cache[1])
	assert.Equal(t, shard03.cache[0], pool.pending.cache[2])

	assert.Equal(t, 1, len(pool.queueMap))
	assert.Equal(t, shard04, pool.queueMap[shard04.headShard])
}
