package main

import (
	"encoding/json"
	"fmt"
	"github.com/everFinance/goar/types"
	ts "github.com/everFinance/turing/example/schema"
	"github.com/everFinance/turing/store/schema"

	"github.com/everFinance/turing/tracker"
)

func main() {
	tags := []types.Tag{
		{Name: "Owner", Value: "k9sXK8x5lMxxM-PbDZ13tCeZi6rOtlll5a6_rrc2oGM"},
	}
	arOwner := "k9sXK8x5lMxxM-PbDZ13tCeZi6rOtlll5a6_rrc2oGM"
	arNode := "https://arweave.net"
	arseed := ""
	cursor := uint64(0)
	dbCfg := schema.Config{}
	tr := tracker.New(tags, arNode, arseed, arOwner, dbCfg)
	tr.Run(cursor)
	for {
		comTx := <-tr.SubscribeTx()
		tx := &ts.Tx{}
		err := json.Unmarshal(comTx.Data, tx)
		if err != nil {
			panic(err)
		}
		fmt.Println(tx)
	}
}
