package tracker

import (
	"github.com/everFinance/goar"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_fetchTxIds(t *testing.T) {
	// arNode := "https://arweave.net"
	// cli := goar.NewClient(arNode, "http://127.0.0.1:8001")
	// tags := []types.Tag{
	// 	types.Tag{Name: "App", Value: "everPay"},
	// 	types.Tag{Name: "Version", Value: "1.0.0"},
	// 	types.Tag{Name: "Owner", Value: "uGx-QfBXSwABKxjha-00dI7vvfyqIYblY6Z5L6cyTFM"},
	// 	types.Tag{Name: "EthLocker", Value: "0x38741a69785e84399fcf7c5ad61d572f7ecb1dab"},
	// 	// TODO: other chain tag, eg: ArLocker, ***graphQL do not use other chain tag***
	// }
	// ids := MustFetchTxIdsByNativeMethod(tags, "", cli)
	// t.Log(ids)
	// bb, err := json.Marshal(ids)
	// assert.NoError(t, err)
	// ioutil.WriteFile("./newIds.json", bb, 0777)
}

func TestFetchIDsByTags(t *testing.T) {
	// tags := []types.Tag{
	// 	types.Tag{Name: "App", Value: "everPay"},
	// 	types.Tag{Name: "Version", Value: "1.0.0"},
	// 	types.Tag{Name: "Owner", Value: "uGx-QfBXSwABKxjha-00dI7vvfyqIYblY6Z5L6cyTFM"},
	// 	types.Tag{Name: "EthLocker", Value: "0x38741a69785e84399fcf7c5ad61d572f7ecb1dab"},
	// 	// TODO: other chain tag, eg: ArLocker, ***graphQL do not use other chain tag***
	// }
	//
	// arNode := "https://arweave.net"
	// cli := goar.NewClient(arNode, "http://127.0.0.1:8001")
	// ids, err := FetchIDsByTags(tags, "", cli)
	// assert.NoError(t, err)
	// bb, err := json.Marshal(ids)
	// assert.NoError(t, err)
	// ioutil.WriteFile("./grahqlIds.json", bb, 0777)
}

func TestNew(t *testing.T) {
	// jsIds01, err := ioutil.ReadFile("./newIds.json")
	// assert.NoError(t, err)
	//
	// jsIds02, err := ioutil.ReadFile("./grahqlIds.json")
	// assert.NoError(t, err)
	//
	// ids01 := make([]string, 0)
	// err = json.Unmarshal(jsIds01, &ids01)
	// assert.NoError(t, err)
	//
	// ids02 := make([]string, 0)
	// err = json.Unmarshal(jsIds02, &ids02)
	// assert.NoError(t, err)
	//
	// for i, v := range ids02 {
	// 	if ids01[i+2] != v {
	// 		panic(v)
	// 	}
	// }
}

func Test_getParentIdByTags(t *testing.T) {
	arNode := "https://arweave.net"
	cli := goar.NewClient(arNode)
	arId := "gdXUJuj9EZm99TmeES7zRHCJtnJoP3XgYo_7KJNV8Vw"
	parentId, err := getParentIdByTags(arId, cli)
	assert.NoError(t, err)
	assert.Equal(t, "z2W7TtaXOdOrdhzK8IXm_fi5yW4NjbW3QJv56lce8nU", parentId)
}
