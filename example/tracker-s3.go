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
	cursor := uint64(40)
	dbCfg := schema.Config{
		UseS3:     true,
		AccKey:    "",
		SecretKey: "MOPfueG+//",
		BktPrefix: "turing",
		Region:    "ap-northeast-1",
	}
	tr := tracker.New(tags, arNode, arseed, arOwner, dbCfg)
	tr.Run(cursor)
	for {
		tx := <-tr.SubscribeTx()
		fmt.Println(tx.CursorId)
	}
}
