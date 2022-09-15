package rawdb

import "github.com/everFinance/everpay-go/common"

var log = common.NewLog("turing")

type KeyValueDB interface {
	Put(bucket, key string, value []byte) (err error)

	Get(bucket, key string) (data []byte, err error)

	GetAllKey(bucket string) (keys []string, err error)

	Delete(bucket, key string) (err error)

	Close() (err error)

	Clear() (err error)
}
