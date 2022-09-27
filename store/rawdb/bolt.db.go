package rawdb

import (
	"errors"
	"fmt"
	"github.com/everFinance/turing/store/schema"
	bolt "go.etcd.io/bbolt"
	"os"
	"path"
	"time"
)

const (
	boltAllocSize = 8 * 1024 * 1024
	boltName      = "seed.db"
)

type BoltDB struct {
	Db     *bolt.DB
	DbPath string
	DbFile string
}

func NewBoltDB(cfg schema.Config) (*BoltDB, error) {
	if len(cfg.DbPath) == 0 {
		return nil, errors.New("boltDb dir path can not null")
	}
	if err := os.MkdirAll(cfg.DbPath, os.ModePerm); err != nil {
		return nil, err
	}
	if len(cfg.DbFileName) == 0 {
		cfg.DbFileName = boltName
	}
	Db, err := bolt.Open(path.Join(cfg.DbPath, cfg.DbFileName), 0660, &bolt.Options{Timeout: 2 * time.Second, InitialMmapSize: 10e6})
	if err != nil {
		if err == bolt.ErrTimeout {
			return nil, errors.New("cannot obtain database lock, database may be in use by another process")
		}
		return nil, err
	}
	Db.AllocSize = boltAllocSize
	boltDB := &BoltDB{
		Db:     Db,
		DbPath: cfg.DbPath,
		DbFile: cfg.DbFileName,
	}
	if err := boltDB.Db.Update(func(tx *bolt.Tx) error {
		return createBuckets(tx, cfg.Bkt)
	}); err != nil {
		return nil, err
	}
	return boltDB, nil
}

func (s *BoltDB) Put(bucket, key string, value []byte) (err error) {
	err = s.Db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(bucket))
		return bkt.Put([]byte(key), value)
	})
	return
}

func (s *BoltDB) Get(bucket, key string) (data []byte, err error) {
	err = s.Db.View(func(tx *bolt.Tx) error {
		data = tx.Bucket([]byte(bucket)).Get([]byte(key))
		if data == nil {
			err = schema.ErrNotExist
			return err
		}
		return nil
	})
	return
}

func (s *BoltDB) GetAllKey(bucket string) (keys []string, err error) {
	keys = make([]string, 0)
	err = s.Db.View(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte(bucket)).ForEach(func(k, v []byte) error {
			keys = append(keys, string(k))
			return nil
		})
	})
	return
}

func (s *BoltDB) Delete(bucket, key string) (err error) {
	err = s.Db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte(bucket)).Delete([]byte(key))
	})
	return
}

func (s *BoltDB) Close() (err error) {
	return s.Db.Close()
}

func (s *BoltDB) Clear() (err error) {
	if _, err := os.Stat(s.DbPath); os.IsNotExist(err) {
		return nil
	}
	if err := os.Remove(path.Join(s.DbPath, s.DbFile)); err != nil {
		return fmt.Errorf("could not remove database file. error: %v", err)
	}
	return nil
}

func createBuckets(tx *bolt.Tx, buckets []string) error {
	if len(buckets) == 0 {
		buckets = append(buckets, schema.AllBkt...)
	}
	for _, bucket := range buckets {
		if _, err := tx.CreateBucketIfNotExists([]byte(bucket)); err != nil {
			return err
		}
	}
	return nil
}
