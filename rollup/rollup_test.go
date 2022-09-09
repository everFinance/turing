package rollup

import (
	"encoding/json"
	"fmt"
	"github.com/everFinance/goar/types"
	"github.com/stretchr/testify/assert"
	"reflect"
	"strings"
	"testing"
)

type Tx struct {
	From   string      `json:"from"`
	To     string      `json:"to"`
	Amount int         `json:"amount"`
	Extra  interface{} `json:"extra"`
}

func marshal(tx interface{}) []byte {
	data, err := json.Marshal(tx)
	if err != nil {
		panic(err)
	}
	return data
}

type aaa struct {
	AA string `json:"aa"`
}

func TestRollup(t *testing.T) {
	tx := &Tx{
		From:   "faaa",
		To:     "taaa",
		Amount: 1199,
		Extra: aaa{
			AA: "lklk",
		},
	}

	txs := make([]*Tx, 0)
	txs = append(txs, tx)
	data := marshal(txs)
	t.Logf("%s", data)

	newTx := make([]*Tx, 0)
	err := json.Unmarshal(data, &newTx)
	assert.NoError(t, err)
	t.Log(newTx[0].Extra)
}

func TestNew(t *testing.T) {
	type em struct{}
	fmt.Println(reflect.TypeOf(em{}).PkgPath())
	path := reflect.TypeOf(em{}).PkgPath()
	pkg := strings.Split(path, "/")
	t.Log(pkg[len(pkg)-1])
}

func TestRollup_verifyLastArTx(t *testing.T) {
	arNode := "https://arweave.net"
	owner := "dQzTM9hXV5MD1fRniOKI3MvPF_-8b2XDLmpfcMN9hi8"
	lastArId := "qU8ZxUPuzj8QGmR2D4sqabApC2GLgxkMb9r7O_6Z-Tc"
	tags := []types.Tag{
		types.Tag{Name: "App", Value: "everPay"},
		types.Tag{Name: "Version", Value: "1.0.0"},
		types.Tag{Name: "Owner", Value: "dQzTM9hXV5MD1fRniOKI3MvPF_-8b2XDLmpfcMN9hi8"},
		types.Tag{Name: "EthLocker", Value: "0xa7ae99C13d82dd32fc6445Ec09e38d197335F38a"},
	}
	err := verifyLastArTxId(lastArId, arNode, owner, tags)
	assert.NoError(t, err)
}

func TestRollup_arTxWatcher(t *testing.T) {
	// txId := "bGiww-UVuYdwYVNz5cKylPPFwhFdu9C8lWJuusW65Ek"
	// cli := goar.NewClient("https://arweave.net")
	// ok := arTxWatcher(cli, txId)
	// t.Log(ok)
}
