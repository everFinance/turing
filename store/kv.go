package store

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"github.com/everFinance/goar/utils"
	"github.com/everFinance/turing/store/rawdb"
	"github.com/everFinance/turing/store/schema"

	types "github.com/everFinance/turing/common"
)

// init log mode
var log = types.NewLog(types.StoreLogMode)

type Store struct {
	KVDb rawdb.KeyValueDB
}

func NewBoltStore(cfg schema.Config) (*Store, error) {
	Db, err := rawdb.NewBoltDB(cfg)
	if err != nil {
		panic(err)
	}

	return &Store{
		KVDb: Db,
	}, nil
}

func NewS3Store(cfg schema.Config) (*Store, error) {
	Db, err := rawdb.NewS3DB(cfg)
	if err != nil {
		panic(err)
	}
	return &Store{KVDb: Db}, nil
}

// DatabasePath at which this database writes files.
// func (kv *Store) DatabasePath() string {
// 	return kv.databasePath
// }

// Close closes the underlying BoltDB database.
func (kv *Store) Close() error {
	return kv.KVDb.Close()
}

func (kv *Store) ClearDB() error {
	return kv.KVDb.Clear()
}

// GetConstant
func (kv *Store) GetConstant(key string) (value string, err error) {
	val, err := kv.KVDb.Get(schema.ConstantBucket, key)
	if err == schema.ErrNotExist {
		return "", nil
	}
	value = string(val)
	return
}

// UpdateConstant
func (kv *Store) UpdateConstant(key string, val []byte) error {
	return kv.KVDb.Put(schema.ConstantBucket, key, val)
}

// LoadPoolTxFromDb
func (kv *Store) LoadPoolTxFromDb() (types.Transactions, error) {
	poolHashArr, err := kv.KVDb.GetAllKey(schema.PoolTxIndex)

	log.Info("store load pool tx", "tx number", len(poolHashArr))
	if err != nil {
		return nil, err
	}

	poolTxs := make(types.Transactions, 0, len(poolHashArr))
	for _, txHash := range poolHashArr {
		val, err := kv.KVDb.Get(schema.AllTokenTxBucket, txHash)
		if err != nil {
			return nil, err
		}
		if len(val) == 0 {
			panic("can not get pending tx from allTokenTxBucket")
		} else {
			ttx := &types.Transaction{}
			err = json.Unmarshal(val, ttx)
			if err != nil {
				return nil, err
			}
			poolTxs = append(poolTxs, ttx)
		}
	}
	return poolTxs, nil
}

func (kv *Store) GetLastOnChainTx() (onChainTx *types.Transaction, err error) {
	txHash, err := kv.GetConstant(schema.LastOnChainTokenTxHashKey)
	if err != nil {
		return nil, err
	}
	if txHash == "" {
		return nil, schema.ErrNotExist
	}

	// get tokenTx from AllTokenTxBucket
	val, err := kv.KVDb.Get(schema.AllTokenTxBucket, txHash)
	if err != nil {
		return
	}
	if len(val) == 0 {
		return nil, errors.New("not found last onchain tokenTx from rollup AllTokenTxBucket db")
	}
	tt := types.Transaction{}
	err = json.Unmarshal(val, &tt)
	if err != nil {
		return
	}
	onChainTx = &tt
	return
}

// PutTokenTx
func (kv *Store) PutTokenTx(tokenTx *types.Transaction) error {
	val, err := tokenTx.Marshal()
	if err != nil {
		return err
	}
	return kv.KVDb.Put(schema.AllTokenTxBucket, tokenTx.TxId, val)
}

// put tx id
func (kv *Store) PutPoolTokenTxId(txId string) error {
	return kv.KVDb.Put(schema.PoolTxIndex, txId, []byte("0x01"))
}

// BatchDelPoolTokenTxId
func (kv *Store) BatchDelPoolTokenTxId(txs types.Transactions) (err error) {
	for _, tx := range txs {
		err = kv.KVDb.Delete(schema.PoolTxIndex, tx.TxId)
		return
	}
	return nil
}

// itob returns an 64-byte big endian representation of v.
func itob(v uint64) string {
	b := make([]byte, 64)
	binary.BigEndian.PutUint64(b, v)
	return utils.Base64Encode(b)
}

func btoi(base64Str string) uint64 {
	b, err := utils.Base64Decode(base64Str)
	if err != nil {
		panic(err)
	}
	return binary.BigEndian.Uint64(b)
}

// PutSubscribeTx
func (kv *Store) PutSubscribeTx(subTx types.SubscribeTransaction) (cursor uint64, err error) {
	value, err := subTx.Marshal()
	if err != nil {
		log.Error("json marshal subscribe transaction error", "err", err)
		return 0, err
	}

	seqNumBy, err := kv.GetConstant(schema.SeqNum)
	if err != nil {
		return
	}
	if len(seqNumBy) == 0 {
		seqNumBy = itob(uint64(1))
	}
	cursor = btoi(seqNumBy)
	err = kv.KVDb.Put(schema.AllSyncedTokenTxBucket, seqNumBy, value)
	if err != nil {
		return
	}
	// seqNum += 1
	err = kv.KVDb.Put(schema.ConstantBucket, schema.SeqNum, []byte(itob(cursor+1)))
	if err != nil {
		_ = kv.KVDb.Delete(schema.AllSyncedTokenTxBucket, seqNumBy)
	}
	return
}

// UpdateLastProcessArTxId
func (kv *Store) UpdateLastProcessArTxId(id string) error {
	return kv.KVDb.Put(schema.ConstantBucket, schema.LastProcessArTxIdKey, []byte(id))
}

// LoadSubscribeTxs
func (kv *Store) LoadSubscribeTxsToStream(cursor uint64, txChan chan<- types.SubscribeTransaction) (err error) {
	keys, err := kv.KVDb.GetAllKey(schema.AllSyncedTokenTxBucket)
	if err != nil {
		return
	}
	for _, key := range keys {
		if btoi(key) <= cursor {
			continue
		}
		val, err := kv.KVDb.Get(schema.AllSyncedTokenTxBucket, key)
		if err != nil {
			return err
		}
		tx := &types.SubscribeTransaction{}
		err = tx.Unmarshal(val)
		if err != nil {
			return err
		}
		tx.CursorId = btoi(key)
		txChan <- *tx
	}
	return
}

// // LoadRollupEverTxs load tx from rollup.db
// func (kv *Store) LoadRollupEverTxs(txChan chan<- types.Transaction) error {
// 	err := kv.Db.View(func(tx *bolt.Tx) error {
// 		bkt := tx.Bucket(AllTokenTxBucket)
//
// 		return bkt.ForEach(func(_, v []byte) error {
// 			tokenTx := &types.Transaction{}
// 			if err := tokenTx.Unmarshal(v); err != nil {
// 				return err
// 			}
// 			txChan <- *tokenTx
// 			return nil
// 		})
// 	})
// 	return err
// }
