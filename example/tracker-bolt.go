package main

import (
	"fmt"
	"github.com/everFinance/goar/types"
	"github.com/everFinance/turing/store/schema"
	"github.com/everFinance/turing/tracker"
)

func main() {
	tags := []types.Tag{
		{Name: "Owner", Value: "uGx-QfBXSwABKxjha-00dI7vvfyqIYblY6Z5L6cyTFM"},
	}
	arOwner := "uGx-QfBXSwABKxjha-00dI7vvfyqIYblY6Z5L6cyTFM"
	arNode := "https://arweave.net"
	arseed := ""
	cursor := uint64(7)
	dbCfg := schema.Config{}
	tr := tracker.New(tags, arNode, arseed, arOwner, dbCfg)
	tr.Run(cursor)
	for {
		tx := <-tr.SubscribeTx()
		fmt.Println(tx.CursorId)
	}
}
