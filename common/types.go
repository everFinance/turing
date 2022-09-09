package common

import (
	"encoding/json"
	"errors"
)

const (
	RollupLogMode  = "rollup"
	TxPoolLogMode  = "rollupTxPool"
	TrackerLogMode = "tracker"
	StoreLogMode   = "store"
	ParentTag      = "parent_id"

	genesisParentId = "genesis-parent-id"
)

type Transactions []*Transaction

func TxsSort(txs Transactions, lastPendTokenTxHash string) (Transactions, error) {
	if len(txs) <= 1 {
		return txs, nil
	}
	IdMap := make(map[string]*Transaction) // key: txId
	for _, tx := range txs {
		if _, ok := IdMap[tx.TxId]; ok {
			return nil, errors.New("exist more than two txs has same TxId in loaded pool db txs")
		} else {
			IdMap[tx.TxId] = tx
		}
	}
	parentIdMap := make(map[string]*Transaction)
	for _, tx := range txs {
		var parentId string
		if len(tx.ParentId) == 0 { // genesis tx
			parentId = genesisParentId
		} else {
			parentId = tx.ParentId
		}
		if _, ok := parentIdMap[parentId]; ok {
			return nil, errors.New("exist more than two txs has same ParentTxId in loaded pool db txs")
		} else {
			parentIdMap[parentId] = tx
		}
	}

	// 通过 lastPendTokenTxHash 找出当前交易池中的第一笔交易
	if len(lastPendTokenTxHash) == 0 {
		lastPendTokenTxHash = genesisParentId
	}
	var headTx *Transaction
	if tx, ok := parentIdMap[lastPendTokenTxHash]; !ok {
		// 不存在则panic
		panic("pool tx is not continuous")
	} else {
		headTx = &Transaction{
			TxData:   tx.TxData,
			TxId:     tx.TxId,
			ParentId: tx.ParentId,
		}
	}

	tailTx := headTx
	pendingTxs := make(Transactions, 0, len(txs))
	pendingTxs = append(pendingTxs, tailTx)
	queueTxs := make(Transactions, 0, len(txs))

	for len(parentIdMap) > 1 { // 其中包括headTx， 所以需要大于1
		nextTx, ok := parentIdMap[tailTx.TxId]
		if ok {
			// add pending
			pendingTxs = append(pendingTxs, nextTx)
			tailTx = nextTx
			// delete parentId
			delete(parentIdMap, nextTx.ParentId)
		} else {
			break
		}
	}
	if len(parentIdMap) > 1 {
		// delete headTx
		delete(parentIdMap, lastPendTokenTxHash)
		// add 剩下的 txs
		for _, tx := range parentIdMap {
			queueTxs = append(queueTxs, tx)
		}
	}

	return append(pendingTxs, queueTxs...), nil
}

type Transaction struct {
	TxData   []byte `json:"tx_data"`
	TxId     string `json:"tx_id"`     // sha(TxData)
	ParentId string `json:"parent_id"` // the last token tx id
}

func (tx Transaction) Marshal() ([]byte, error) {
	return json.Marshal(tx)
}

func (tx *Transaction) Unmarshal(data []byte) error {
	return json.Unmarshal(data, tx)
}

func (txs Transactions) Marshal() ([]byte, error) {
	return json.Marshal(txs)
}

// delete tx_id and parent_id
func (txs Transactions) MarshalFromOnChain() ([]byte, error) {
	lightTxs := make(Transactions, 0, len(txs))
	for _, tx := range txs {
		lightTxs = append(lightTxs, &Transaction{
			TxData: tx.TxData,
		})
	}
	return lightTxs.Marshal()
}

func (txs *Transactions) Unmarshal(data []byte) error {
	return json.Unmarshal(data, txs)
}

type SubscribeTransaction struct {
	CursorId  uint64 // boltdb cursor id
	ID        string // ar tx id
	Owner     string // ar tx sender
	Data      []byte // token tx content
	Timestamp int64  // ar tx timestamp
	Height    int64  // ar tx height
}

func (st SubscribeTransaction) Marshal() ([]byte, error) {
	return json.Marshal(st)
}

func (st *SubscribeTransaction) Unmarshal(data []byte) error {
	return json.Unmarshal(data, st)
}
