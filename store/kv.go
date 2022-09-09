package store

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"

	types "github.com/everFinance/turing/common"
	bolt "go.etcd.io/bbolt"
	"os"
	"path"
	"time"
)

// init log mode
var log = types.NewLog(types.StoreLogMode)

type Store struct {
	Db           *bolt.DB
	databasePath string
}

func NewKvStore(dirPath string, dbFileName string, bucketNames ...[]byte) (*Store, error) {
	if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
		return nil, err
	}
	dataFile := path.Join(dirPath, dbFileName)
	boltDB, err := bolt.Open(dataFile, 0660, &bolt.Options{Timeout: 1 * time.Second, InitialMmapSize: 10e6})
	if err != nil {
		if err == bolt.ErrTimeout {
			return nil, errors.New("cannot obtain database lock,database may be in use by another process")
		}
		return nil, err
	}
	boltDB.AllocSize = boltAllocSize
	kv := &Store{
		Db:           boltDB,
		databasePath: dirPath,
	}
	// create bucket
	if len(bucketNames) == 0 {
		bucketNames = append(bucketNames, AllTokenTxBucket, PoolTxIndex, ConstantBucket, AllSyncedTokenTxBucket)
	}
	if err := kv.Db.Update(func(tx *bolt.Tx) error {
		return createBuckets(tx, bucketNames...)
	}); err != nil {
		return nil, err
	}

	return kv, nil
}

// DatabasePath at which this database writes files.
func (kv *Store) DatabasePath() string {
	return kv.databasePath
}

func createBuckets(tx *bolt.Tx, buckets ...[]byte) error {
	for _, bucket := range buckets {
		if _, err := tx.CreateBucketIfNotExists(bucket); err != nil {
			return err
		}
	}
	return nil
}

// Close closes the underlying BoltDB database.
func (kv *Store) Close() error {
	return kv.Db.Close()
}

func (kv *Store) ClearDB(dbFileName string) error {
	if _, err := os.Stat(kv.databasePath); os.IsNotExist(err) {
		return nil
	}
	if err := os.Remove(path.Join(kv.databasePath, dbFileName)); err != nil {
		return fmt.Errorf("could not remove database file. error: %v", err)
	}
	return nil
}

// GetConstant
func (kv *Store) GetConstant(key []byte) (value string, err error) {
	err = kv.Db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(ConstantBucket)
		val1 := bkt.Get(key)
		if len(val1) == 0 {
			value = ""
		} else {
			value = string(val1)
		}
		return nil
	})
	return
}

// UpdateConstant
func (kv *Store) UpdateConstant(key []byte, val []byte) error {
	err := kv.Db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(ConstantBucket).Put(key, val)
	})
	return err
}

// LoadPoolTxFromDb
func (kv *Store) LoadPoolTxFromDb() (types.Transactions, error) {
	poolHashArr := make([][]byte, 0)
	err := kv.Db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(PoolTxIndex)
		return bkt.ForEach(func(k, v []byte) error {
			poolHashArr = append(poolHashArr, k)
			return nil
		})
	})

	log.Info("store load pool tx", "tx number", len(poolHashArr))
	if err != nil {
		return nil, err
	}

	poolTxs := make(types.Transactions, 0, len(poolHashArr))
	err = kv.Db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(AllTokenTxBucket)
		for _, txHash := range poolHashArr {
			val := bkt.Get(txHash)
			if len(val) == 0 {
				panic("can not get pending tx from allTokenTxBucket")
			} else {
				ttx := &types.Transaction{}
				err = json.Unmarshal(val, ttx)
				if err != nil {
					return err
				}
				poolTxs = append(poolTxs, ttx)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return poolTxs, nil
}

func (kv *Store) GetLastOnChainTx() (onChainTx *types.Transaction, err error) {
	txHash, err := kv.GetConstant(LastOnChainTokenTxHashKey)
	if err != nil {
		return nil, err
	}
	if txHash == "" {
		return nil, ErrNotExist
	}

	// get tokenTx from AllTokenTxBucket
	err = kv.Db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(AllTokenTxBucket)
		val := bkt.Get([]byte(txHash))
		if len(val) == 0 {
			return errors.New("not found last onchain tokenTx from rollup AllTokenTxBucket db")
		}
		tt := types.Transaction{}
		err = json.Unmarshal(val, &tt)
		if err != nil {
			return err
		}
		onChainTx = &tt
		return nil
	})
	return
}

// PutTokenTx
func (kv *Store) PutTokenTx(tokenTx *types.Transaction) error {
	val, err := tokenTx.Marshal()
	if err != nil {
		return err
	}
	err = kv.Db.Update(func(tx *bolt.Tx) error {
		allTxBkt := tx.Bucket(AllTokenTxBucket)
		if err := allTxBkt.Put([]byte(tokenTx.TxId), val); err != nil {
			return err
		}
		return nil
	})
	return err
}

// put tx id
func (kv *Store) PutPoolTokenTxId(txId string) error {
	err := kv.Db.Update(func(tx *bolt.Tx) error {
		txIndexBkt := tx.Bucket(PoolTxIndex)
		if err := txIndexBkt.Put([]byte(txId), []byte("0x01")); err != nil {
			return err
		}
		return nil
	})
	return err
}

// BatchDelPoolTokenTxId
func (kv *Store) BatchDelPoolTokenTxId(txs types.Transactions) error {
	err := kv.Db.Update(func(tx *bolt.Tx) error {
		// delete PoolTxIndex
		bkt := tx.Bucket(PoolTxIndex)
		for _, t := range txs {
			if err := bkt.Delete([]byte(t.TxId)); err != nil {
				log.Error("delete pooltxIndex error", "error", err, "txs", txs)
				return err
			}
		}
		return nil
	})
	return err
}

// itob returns an 64-byte big endian representation of v.
func itob(v uint64) []byte {
	b := make([]byte, 64)
	binary.BigEndian.PutUint64(b, v)
	return b
}

func btoi(b []byte) uint64 {
	return binary.BigEndian.Uint64(b)
}

// PutSubscribeTx
func (kv *Store) PutSubscribeTx(subTx types.SubscribeTransaction) (uint64, error) {
	value, err := subTx.Marshal()
	if err != nil {
		log.Error("json marshal subscribe transaction error", "err", err)
		return 0, err
	}

	var cursor uint64
	err = kv.Db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(AllSyncedTokenTxBucket)
		// use monotonically incrementing, for sort tx
		id, err := bkt.NextSequence()
		if err != nil {
			return err
		}
		cursor = id
		key := itob(id)
		if err := bkt.Put(key, value); err != nil {
			log.Error("store subscribe transaction error", "err", err)
			return err
		}
		return nil
	})

	return cursor, err
}

// UpdateLastProcessArTxId
func (kv *Store) UpdateLastProcessArTxId(id string) error {
	err := kv.Db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(ConstantBucket)
		if err := bkt.Put(LastProcessArTxIdKey, []byte(id)); err != nil {
			return err
		}
		return nil
	})
	return err
}

// LoadSubscribeTxs
func (kv *Store) LoadSubscribeTxsToStream(cursor uint64, txChan chan<- types.SubscribeTransaction) error {
	err := kv.Db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(AllSyncedTokenTxBucket)

		return bkt.ForEach(func(k, v []byte) error {
			if btoi(k) <= cursor {
				return nil
			}
			tx := &types.SubscribeTransaction{}
			err := tx.Unmarshal(v)
			if err != nil {
				return err
			}
			tx.CursorId = btoi(k)
			txChan <- *tx
			return nil
		})
	})
	if err != nil {
		return err
	}
	return nil
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
