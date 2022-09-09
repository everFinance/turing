package common

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTransaction_Marshal(t *testing.T) {
	tx := Transaction{
		TxData:   []byte("aaabbbccc"),
		TxId:     "0x12aa",
		ParentId: "0x11aa",
	}
	data, err := tx.Marshal()
	assert.NoError(t, err)

	tx1 := &Transaction{}
	err = tx1.Unmarshal(data)
	assert.NoError(t, err)
	t.Log(tx1)
}

func TestTransactions_Marshal(t *testing.T) {
	txs := Transactions{
		&Transaction{
			TxData:   []byte("aad"),
			TxId:     "0x22",
			ParentId: "0x11",
		},
		&Transaction{
			TxData:   []byte("bbf"),
			TxId:     "0xff",
			ParentId: "0xee",
		},
	}
	data, err := txs.Marshal()
	assert.NoError(t, err)

	var txs1 = make(Transactions, 0)
	err = txs1.Unmarshal(data)
	assert.NoError(t, err)
	t.Log(txs1[0])
}

func TestTransactions_Sorts(t *testing.T) {
	tx01 := &Transaction{
		TxData:   []byte("tx01 data"),
		TxId:     "111",
		ParentId: "",
	}
	tx02 := &Transaction{
		TxData:   []byte("tx02 data"),
		TxId:     "222",
		ParentId: "111",
	}
	tx03 := &Transaction{
		TxData:   []byte("tx03 data"),
		TxId:     "333",
		ParentId: "222",
	}
	tx04 := &Transaction{
		TxData:   []byte("tx04 data"),
		TxId:     "444",
		ParentId: "333",
	}
	tx05 := &Transaction{
		TxData:   []byte("tx05 data"),
		TxId:     "555",
		ParentId: "444",
	}
	tx06 := &Transaction{
		TxData:   []byte("tx06 data"),
		TxId:     "666",
		ParentId: "555",
	}
	tx07 := &Transaction{
		TxData:   []byte("tx07 data"),
		TxId:     "777",
		ParentId: "666",
	}
	tx08 := &Transaction{
		TxData:   []byte("tx08 data"),
		TxId:     "888",
		ParentId: "777",
	}
	tx09 := &Transaction{
		TxData:   []byte("tx09 data"),
		TxId:     "999",
		ParentId: "888",
	}
	tx10 := &Transaction{
		TxData:   []byte("tx10 data"),
		TxId:     "101010",
		ParentId: "999",
	}

	txs := Transactions{tx01, tx02, tx03, tx09, tx07, tx08, tx10, tx05, tx04, tx06}
	newTxs, err := TxsSort(txs, "")
	assert.NoError(t, err)

	for _, val := range newTxs {
		t.Log(string(val.TxData))
	}

	txs02 := Transactions{}
	tt, err := TxsSort(txs02, "")
	assert.NoError(t, err)
	t.Log(tt)
}
