package tracker

import (
	"testing"
)

func TestMustFetchTxInfo(t *testing.T) {
	// addr, _, _ := MustFetchTxInfo("xkAigg50YoM6pisCXn0jk_6LDV_N9zHLXHYmgQj12BU", nil)
	// assert.Equal(t, "dQzTM9hXV5MD1fRniOKI3MvPF_-8b2XDLmpfcMN9hi8", addr)
	//
	// _, ts, height := MustFetchTxInfo("G-1t0Lqysin897HC3IV8xu_Mr884B-Mo5YEnlhUH54k", nil)
	// assert.Equal(t, int64(1600878676), ts)
	// assert.Equal(t, int64(533628), height)
}

func TestMustFetchTxData(t *testing.T) {
	// res := MustFetchTxData("xkAigg50YoM6pisCXn0jk_6LDV_N9zHLXHYmgQj12BU", nil)
	// assert.Equal(t, `[{"nonce":"1598169690000","type":"claim","from":"0xa06b79E655Db7D7C3B3E7B2ccEEb068c3259d0C9","to":"","amount":"100","sign":"0xaa7ff25bb0cd26e199d99cf2f2d771248c6e26b9c35eee62d70b45aa2f0284ea5dbf196d1159603fbdc23d6d6be73ce4235c97de36e3bc1e06f188a82a3d3f5c1b"},{"nonce":"1598169779000","type":"transfer","from":"0xa06b79E655Db7D7C3B3E7B2ccEEb068c3259d0C9","to":"0xDc19464589c1cfdD10AEdcC1d09336622b282652","amount":"30","sign":"0x08a308a67440a07ab3bbd5075388732965c21f1885923de0429e2a98926a200339aff56240556df1e4c9486c8d94370a963d8d960e31730fa6eff49477d566bc1b"}]`, string(res))
}

func TestFetchAndVerifyTxData(t *testing.T) {
	// res, err := FetchAndVerifyTxData("ynL_FsK3pIxau8reX3SE-Nrkw3NdRiBA2y8EDPtI8OU",goar.NewClient("https://arweave.net"),arseeding.New("https://seed-dev.everpay.io"))
	// assert.NoError(t, err)
	// t.Log(string(res))
}

// func TestMustFetchLastIDByTags(t *testing.T) {
// 	id := MustFetchLastIDByTags(
// 		[]types.Tag{
// 			types.Tag{Name: "App", Value: "everToken"},
// 			types.Tag{Name: "Symbol", Value: "ETH"},
// 			types.Tag{Name: "Owner", Value: "dQzTM9hXV5MD1fRniOKI3MvPF_-8b2XDLmpfcMN9hi8"},
// 		}, nil,
// 	)
// 	t.Log("id", id)
// }

// func TestMustFetchIDsByTags(t *testing.T) {
// 	ids := MustFetchIDsByTags([]types.Tag{
// 		types.Tag{Name: "App", Value: "everToken"},
// 		types.Tag{Name: "Symbol", Value: "ETH"},
// 		types.Tag{Name: "Owner", Value: "dQzTM9hXV5MD1fRniOKI3MvPF_-8b2XDLmpfcMN9hi8"},
// 	}, "", nil)
// 	fmt.Println(ids)
// 	assert.Equal(t, 2, len(ids))
// }

// func TestFetchIDsByTags(t *testing.T) {
// 	ids, err := FetchIDsByTags([]types.Tag{
// 		types.Tag{Name: "TokenSymbol", Value: "ROL"},
// 		types.Tag{Name: "CreatedBy", Value: "dQzTM9hXV5MD1fRniOKI3MvPF_-8b2XDLmpfcMN9hi8"},
// 	}, "sXChMxsx2w7TybqPbhLAYFWHjchob-22wysvblyP8OY", nil)
// 	assert.NoError(t, err)
// 	assert.Equal(t, 2, len(ids))
// }

func TestMustFetchNodesByTagsRec(t *testing.T) {
	// nodes := mustFetchNodesByTagsRec([]types.Tag{
	// 	types.Tag{Name: "TokenSymbol", Value: "ROL"},
	// 	types.Tag{Name: "CreatedBy", Value: "dQzTM9hXV5MD1fRniOKI3MvPF_-8b2XDLmpfcMN9hi8"},
	// }, "sXChMxsx2w7TybqPbhLAYFWHjchob-22wysvblyP8OY", "", nil)
	// assert.Equal(t, 2, len(nodes))
}

func TestMustFetchEdgesByTags(t *testing.T) {
	// edges, hashNextPage := MustFetchEdgesByTags([]types.Tag{
	// 	types.Tag{Name: "TokenSymbol", Value: "ROL"},
	// 	types.Tag{Name: "CreatedBy", Value: "dQzTM9hXV5MD1fRniOKI3MvPF_-8b2XDLmpfcMN9hi8"},
	// }, "", nil)
	// assert.True(t, hashNextPage)
	// assert.Equal(t, 100, len(edges))
}

func TestFetchEdgesByTags(t *testing.T) {
	// edges, hashNextPage, err := FetchEdgesByTags([]types.Tag{
	// 	types.Tag{Name: "TokenSymbol", Value: "ROL"},
	// 	types.Tag{Name: "CreatedBy", Value: "dQzTM9hXV5MD1fRniOKI3MvPF_-8b2XDLmpfcMN9hi8"},
	// }, "", nil)
	// assert.NoError(t, err)
	// assert.True(t, hashNextPage)
	// assert.Equal(t, 100, len(edges))
}
