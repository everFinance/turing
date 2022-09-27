package store

import (
	"github.com/everFinance/turing/store/schema"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewBoltStore(t *testing.T) {
	dirPath := "../bolt"
	dbName := "tracker.db"
	kv, err := NewBoltStore(schema.Config{DbPath: dirPath, DbFileName: dbName})
	assert.NoError(t, err)
	err = kv.UpdateConstant(schema.LastProcessArTxIdKey, []byte("f8-DEvvlAY7qPRRaHm96Sc7b1EcssK8J0IfOmQaUU2c"))
	assert.NoError(t, err)
}

func TestNewS3Store(t *testing.T) {
	dbCfg := schema.Config{
		UseS3:     true,
		AccKey:    "AKIATZSGGOHI72GMNSO7",
		SecretKey: "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",
		BktPrefix: "turing",
		Region:    "ap-northeast-1",
	}
	kv, err := NewS3Store(dbCfg)
	assert.NoError(t, err)
	err = kv.UpdateConstant(schema.LastProcessArTxIdKey, []byte("f8-DEvvlAY7qPRRaHm96Sc7b1EcssK8J0IfOmQaUU2c"))
	assert.NoError(t, err)
}

func TestClearS3(t *testing.T) {
	dbCfg := schema.Config{
		UseS3:     true,
		AccKey:    "AKIATZSGGOHI72GMNSO7",
		SecretKey: "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",
		BktPrefix: "turing",
		Region:    "ap-northeast-1",
	}
	kv, err := NewS3Store(dbCfg)
	assert.NoError(t, err)
	err = kv.ClearDB()
	assert.NoError(t, err)
}

// func TestNewKvStore(t *testing.T) {
// 	dirPath := "./"
// 	dbName := "test.Db"
// 	kv, err := NewKvStore(dirPath, dbName, AllTokenTxBucket, PoolTxIndex, ConstantBucket)
// 	assert.NoError(t, err)
// 	k1 := []byte("key01")
// 	v1 := []byte("value01")
// 	err = kv.Db.Update(func(tx *bolt.Tx) error {
// 		bkt := tx.Bucket(AllTokenTxBucket)
// 		err = bkt.Put(k1, v1)
// 		return err
// 	})
// 	assert.NoError(t, err)
//
// 	var val []byte
// 	kv.Db.View(func(tx *bolt.Tx) error {
// 		val = tx.Bucket(AllTokenTxBucket).Get(k1)
// 		return nil
// 	})
// 	assert.Equal(t, val, v1)
//
// 	err = kv.Db.Update(func(tx *bolt.Tx) error {
// 		return tx.Bucket(AllTokenTxBucket).Delete(k1)
// 	})
// 	assert.NoError(t, err)
//
// 	kv.Db.View(func(tx *bolt.Tx) error {
// 		val := tx.Bucket(AllTokenTxBucket).Get(k1)
// 		assert.Nil(t, val)
// 		return nil
// 	})
//
// 	err = kv.Db.View(func(tx *bolt.Tx) error {
// 		return tx.Bucket(AllTokenTxBucket).ForEach(func(k, v []byte) error {
// 			t.Log(k, v)
// 			return nil
// 		})
// 	})
// 	assert.NoError(t, err)
// 	err = kv.Db.View(func(tx *bolt.Tx) error {
// 		return tx.Bucket(PoolTxIndex).ForEach(func(k, v []byte) error {
// 			t.Log(string(k), string(v))
// 			return nil
// 		})
// 	})
//
// }

// func TestStore_LoadSubscribeTxsToStream(t *testing.T) {
// 	dirPath := "./"
// 	dbName := "test.Db"
// 	kv, err := NewKvStore(dirPath, dbName)
// 	assert.NoError(t, err)

// 	for i := 0; i < 100; i++ {
// 		id := fmt.Sprintf("aaa, %d", i)
// 		data := fmt.Sprintf("zzzz, %d", i)
// 		subTx := common.SubscribeTransaction{
// 			ID:    id,
// 			Owner: "0x22",
// 			Data:  []byte(data),
// 		}
// 		err = kv.PutSubscribeTx(subTx)
// 		assert.NoError(t, err)
// 	}

// txChan := make(chan common.SubscribeTransaction,100)
// go func() {
// 	err := kv.LoadSubscribeTxsToStream(txChan)
// 	assert.NoError(t, err)
// }()
//
// for {
// 	select {
// 	case tx := <- txChan:
// 		t.Log(tx)
// 	}
// }
// }

// func TestStore_BackupDB(t *testing.T) {
// 	dirPath := "./"
// 	dbName := "test.Db"
// 	kv, err := NewKvStore(dirPath, dbName)
// 	assert.NoError(t, err)

// 	err = os.MkdirAll("backup", os.ModePerm)
// 	assert.NoError(t, err)
// 	f, err := os.OpenFile("./backup/backup-kv", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
// 	assert.NoError(t, err)

// 	// backup db
// 	err = kv.Db.View(func(tx *bolt.Tx) error {
// 		_, err := tx.WriteTo(f)
// 		return err
// 	})
// 	assert.NoError(t, err)
// 	f.Close()
// }

// pay tx
type Transaction struct {
	TokenSymbol  string `json:"tokenSymbol"`
	Action       string `json:"action"`
	From         string `json:"from"`
	To           string `json:"to"`
	Amount       string `json:"amount"`
	Fee          string `json:"fee"`
	FeeRecipient string `json:"feeRecipient"`
	Nonce        string `json:"nonce"`
	TokenID      string `json:"tokenID"`
	ChainType    string `json:"chainType"`
	ChainID      string `json:"chainID"`
	Data         string `json:"data"`
	Version      string `json:"version"`
	Sig          string `json:"sig"`

	ArOwner string `json:"-"`
	ArTxID  string `json:"-"`
}

//
func TestStore_LoadPoolTxFromDb(t *testing.T) {
	// 	dirPath := "./"
	// 	dbName := "rollup.db"
	// 	 kv , err := NewKvStore(dirPath, dbName)
	// 	 assert.NoError(t, err)
	// 	 txChan := make(chan common.Transaction)
	// 	 count := 0
	// 	 go func() {
	// 	 	for {
	// 			select {
	// 			case tt := <-txChan:
	// 				tx := Transaction{}
	// 				if err := json.Unmarshal(tt.TxData,&tx); err != nil {
	// 					panic(err)
	// 				}
	// 				if tx.From == "0x1C7c965F11850A931E64D16347439c22fC972f70" {
	// 					t.Log(tx)
	// 				}
	// 				count++
	// 			}
	// 		}
	// 	 }()
	//
	// 	 // err = kv.LoadSubscribeTxsToStream(0,txChan)
	// 	 err = kv.LoadRollupEverTxs(txChan)
	// 	 assert.NoError(t, err)
	// 	 t.Log(count)
}

func TestStore_GetLastOnChainTokenTx(t *testing.T) {
	// dirPath := "./"
	// dbName := "rollup0425.db"
	// kv, err := NewKvStore(dirPath, dbName)
	// assert.NoError(t, err)
	// tokenTx, err := kv.GetLastOnChainTokenTx()
	// assert.NoError(t, err)
	//
	// payTx := Transaction{}
	// err = json.Unmarshal(tokenTx.TxData, &payTx)
	// assert.NoError(t, err)
	// t.Log(payTx)
}
