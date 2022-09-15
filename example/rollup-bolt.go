package main

import (
	"github.com/everFinance/goar/types"
	"github.com/everFinance/turing/rollup"
	"github.com/everFinance/turing/store/schema"
	"time"
)

func main() {
	tags := []types.Tag{
		{Name: "Owner", Value: "uGx-QfBXSwABKxjha-00dI7vvfyqIYblY6Z5L6cyTFM"},
	}
	suggestLastArTxId := ""
	arOwner := "uGx-QfBXSwABKxjha-00dI7vvfyqIYblY6Z5L6cyTFM"
	arNode := "https://arweave.net"
	arWalletKeyPath := "./key.json"
	rol := rollup.New(suggestLastArTxId, arNode, "", arWalletKeyPath, arOwner, tags, schema.Config{})
	rol.Run(5*time.Second, 999)
}
