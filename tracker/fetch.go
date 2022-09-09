package tracker

import (
	"encoding/json"
	"errors"
	"fmt"
	arseeding "github.com/everFinance/arseeding/sdk"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/everFinance/goar"

	"github.com/everFinance/goar/types"
	"github.com/everFinance/goar/utils"
)

func (t *Tracker) mustFetchDataByID(id string) (owner string, timestamp, height int64, data []byte) {
	data = MustFetchTxData(id, t.arClient, t.arSeedingCli)
	owner, timestamp, height = MustFetchTxInfo(id, t.arClient)

	return
}

func MustFetchTxInfo(id string, c *goar.Client) (ownerAddr string, timestamp, height int64) {
	if c == nil {
		c = goar.NewClient("https://arweave.net")
	}

	var err error
	for {
		ownerAddr, timestamp, height, err = FetchTxInfo(id, c)
		if err == nil {
			break
		}

		log.Warn("fetch tx timestamp failed, retry 10 secs", "timestamp", timestamp, "height", height, "id", id, "err", err)
		time.Sleep(10 * time.Second)
	}

	return
}

func FetchTxInfo(id string, c *goar.Client) (address string, timestamp, height int64, err error) {
	if c == nil {
		c = goar.NewClient("https://arweave.net")
	}

	owner, err := c.GetTransactionField(id, "owner")
	if err != nil {
		return
	}
	address, err = utils.OwnerToAddress(owner)
	if err != nil {
		return
	}

	status, err := c.GetTransactionStatus(id)
	if err != nil {
		return
	}
	height = int64(status.BlockHeight)

	block, err := c.GetBlockByHeight(height)
	if err != nil {
		return
	}
	timestamp = block.Timestamp

	return
}

func MustFetchTxData(id string, c *goar.Client, arSeedingCli *arseeding.ArSeedCli) (res []byte) {
	if c == nil {
		c = goar.NewClient("https://arweave.net")
	}

	var err error
	for {
		res, err = FetchAndVerifyTxData(id, c, arSeedingCli)
		if err == nil {
			break
		}

		log.Warn("fetch tx data failed, retry 10 secs", "id", id, "body", string(res), "err", err)
		time.Sleep(10 * time.Second)
	}

	return
}

func FetchAndVerifyTxData(id string, c *goar.Client, arSeedingCli *arseeding.ArSeedCli) (res []byte, err error) {
	if c == nil {
		c = goar.NewClient("https://arweave.net")
	}

	// get tx data by gateway
	res, err = c.GetTransactionDataByGateway(id)
	if err != nil {
		// get tx data by arweave node
		log.Warn("fetch tx data failed, retry with arweave node", "id", id, "err", err)
		res, err = c.GetTransactionData(id, "json")
		if err != nil {
			log.Warn("fetch tx data by gateway failed", "id", id, "err", err)
			if arSeedingCli == nil {
				return
			}

			// get tx data from arseeding if exist arseeding server
			res, err = arSeedingCli.ACli.GetTransactionData(id, "json")
			if err != nil {
				log.Warn("fetch tx data by arseeding failed", "id", id, "err", err)
				return
			}
		}
	}

	// verify data
	chunks := utils.GenerateChunks(res)
	dataRoot := utils.Base64Encode(chunks.DataRoot)

	// get tx root
	var dr string
	dr, err = c.GetTransactionField(id, "data_root")
	if err != nil {
		log.Error("get arTx dataRoot failed", "err", err, "arId", id)
		// 2021/12/2 22:10 arweave.net /tx/{{arId}}/data_root return Not Found;
		// so try use /tx/{{ar_id}} to get dataRoot
		var tx *types.Transaction
		tx, err = c.GetTransactionByID(id)
		if err != nil {
			log.Warn("get arTx from arweave.net failed", "err", err, "id", id)
			if arSeedingCli == nil {
				return
			}
			// get tx from arseeding if exist arseeding server
			tx, err = arSeedingCli.ACli.GetTransactionByID(id)
			if err != nil {
				log.Warn("fetch tx by arseeding failed", "id", id, "err", err)
				return
			}
		}
		dr = tx.DataRoot
	}

	if dataRoot != dr {
		err = errors.New("invalid data_root")
	}

	return res, err
}

// Use GateWay fetch data
// TODO: remove this, use v2 download chunks data
// Deprecated: Use goar GetTransactionByGateway instead
func FetchAndVerifyTxDataIndirect(id string, c *goar.Client) (res []byte, err error) {
	if c == nil {
		c = goar.NewClient("https://arweave.net")
	}

	resp, err := http.Get(fmt.Sprintf("https://arweave.net/%v", id))
	if err != nil {
		return
	}

	defer resp.Body.Close()
	res, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	// verify data
	chunks := utils.GenerateChunks(res)
	dataRoot := utils.Base64Encode(chunks.DataRoot)

	dr, err := c.GetTransactionField(id, "data_root")
	if err != nil {
		return
	}

	if dataRoot != string(dr) {
		err = errors.New("invalid data_root")
	}

	return
}

func MustFetchLastIDByTags(tags []types.Tag, c *goar.Client) string {
	if c == nil {
		c = goar.NewClient("https://arweave.net")
	}

	edges, _ := MustFetchEdgesByTags(tags, "", c)
	if len(edges) == 0 {
		return ""
	}

	return edges[0].Node.ID
}

func MustFetchIDsByTags(tags []types.Tag, afterID string, c *goar.Client) (ids []string) {
	if c == nil {
		c = goar.NewClient("https://arweave.net")
	}

	var err error
	for {
		ids, err = FetchIDsByTags(tags, afterID, c)
		if err == nil {
			break
		}

		log.Warn("fetch tx ids failed, retry 10 secs", "tags", tags, "afterID", afterID, "err", err)
		time.Sleep(10 * time.Second)
	}

	return
}

func FetchIDsByTags(tags []types.Tag, afterID string, c *goar.Client) (ids []string, err error) {
	if c == nil {
		c = goar.NewClient("https://arweave.net")
	}

	nodes := mustFetchNodesByTagsRec(tags, afterID, "", c)
	if len(nodes) == 1 {
		ids = []string{nodes[0].ID}
		return
	}

	if len(nodes) == 0 {
		log.Info("no new nodes found")
		return
	}

	// check prevID sort，make sure get all tx
	prevID := ""
	if len(nodes) > 1 {
		prevID = nodes[0].ID
	}

	for _, node := range nodes {
		if prevID != node.ID {
			continue
		}

		ids = append(ids, node.ID)
		prevID = node.Tags[len(node.Tags)-1].Value
	}

	if prevID != afterID {
		err = fmt.Errorf("prevID error: prevID=%v", prevID)
	}

	return
}

// recursive
func mustFetchNodesByTagsRec(tags []types.Tag, afterID, afterCursor string, c *goar.Client) (nodes []GNode) {
	edges, hasNextPage := MustFetchEdgesByTags(tags, afterCursor, c)
	if len(edges) == 0 {
		return
	}

	for _, edge := range edges {
		if edge.Node.ID == afterID {
			return
		}

		nodes = append(nodes, edge.Node)
	}

	if !hasNextPage {
		return
	}

	lastEdge := edges[len(edges)-1]
	return append(nodes, mustFetchNodesByTagsRec(tags, afterID, lastEdge.Cursor, c)...)
}

// GraphQL

// GPageInfo is GraphQL Resposne pageInfo struct
type GPageInfo struct{ HasNextPage bool }

type GBlock struct {
	ID        string
	Timestamp int64
	Height    int64
}

type GNode struct {
	ID    string
	Tags  []types.Tag
	Block GBlock
}

// GEdges is GraphQL Response edges struct
type GEdge struct {
	Cursor string
	Node   GNode
}

type GTransaction struct {
	Transactions struct {
		PageInfo GPageInfo
		Edges    []GEdge
	}
}

// MustFetchEdgesByTags just return confirmed tx
func MustFetchEdgesByTags(tags []types.Tag, afterCur string, c *goar.Client) (edges []GEdge, hasNextPage bool) {
	if c == nil {
		c = goar.NewClient("https://arweave.net")
	}

	var err error
	for {
		edges, hasNextPage, err = FetchEdgesByTags(tags, afterCur, c)
		if err == nil {
			break
		}

		log.Warn("fetch edges failed, retry 10 secs", "tags", tags, "afterCur", afterCur, "err", err)
		time.Sleep(10 * time.Second)
	}

	return
}

// FetchEdgesByTags just return confirmed tx
func FetchEdgesByTags(tags []types.Tag, afterCur string, c *goar.Client) (edges []GEdge, hasNextPage bool, err error) {
	if c == nil {
		c = goar.NewClient("https://arweave.net")
	}

	query := `{
	transactions(
		first: 100
		tags: ` + tagsToQry(tags) + `
	) { pageInfo { hasNextPage } edges { cursor node { id tags { name value } block { timestamp }}}}}`
	if afterCur != "" {
		query = `{
		transactions(
			first: 100
			after: "` + afterCur + `"
			tags: ` + tagsToQry(tags) + `
		) { pageInfo { hasNextPage } edges { cursor node { id tags { name value } block { timestamp }}}}}`
	}

	data, err := c.GraphQL(query)
	if err != nil {
		return
	}

	txs := GTransaction{}

	if err = json.Unmarshal(data, &txs); err != nil {
		return
	}

	hasNextPage = txs.Transactions.PageInfo.HasNextPage

	// filter pending tx
	for _, edge := range txs.Transactions.Edges {
		if edge.Node.Block.Timestamp != 0 {
			edges = append(edges, edge)
		}
	}

	return
}

// tagsToQry: format tags to `{name:"N",values:"V"}` string
func tagsToQry(tags []types.Tag) (qry string) {
	strs := []string{}
	for _, tag := range tags {
		strs = append(strs, fmt.Sprintf(`{name:"%s",values:"%s"}`, tag.Name, tag.Value))
	}
	return fmt.Sprintf(`[%s]`, strings.Join(strs, ","))
}
