package main

import (
	"encoding/json"
	"fmt"
	"github.com/everFinance/goar/types"
	ts "github.com/everFinance/turing/example/schema"
	"github.com/everFinance/turing/rollup"
	"github.com/everFinance/turing/store/schema"
	"time"
)

func main() {
	tags := []types.Tag{
		{Name: "App", Value: "turing-test"},
		{Name: "Owner", Value: "k9sXK8x5lMxxM-PbDZ13tCeZi6rOtlll5a6_rrc2oGM"},
	}
	suggestLastArTxId := ""
	arOwner := "k9sXK8x5lMxxM-PbDZ13tCeZi6rOtlll5a6_rrc2oGM"
	arNode := "https://arweave.net"
	arWalletKeyPath := "./k9s.json"
	rol := rollup.New(suggestLastArTxId, arNode, "", arWalletKeyPath, arOwner, tags, schema.Config{})
	rol.Run(2*time.Minute, 999)
	feedData(rol.AddTx())
}

func feedData(ch chan<- []byte) {
	ticker := time.NewTicker(30 * time.Second)
	var cnt int64
	for {
		select {
		case <-ticker.C:
			tx := &ts.Tx{
				Name:      fmt.Sprintf("test-%v", cnt),
				Timestamp: time.Now().UnixMilli(),
			}
			data, err := json.Marshal(tx)
			if err != nil {
				panic(err)
			}
			cnt += 1
			ch <- data
		}
	}
}
