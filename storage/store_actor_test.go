package storage

import (
	"log"
	"os"
	"testing"

	"path/filepath"
	"stashware/types"
)

var (
	storer *StoreActor
)

func init() {

	var (
		err error
		cwd = filepath.Dir(os.Args[0]) // C:\Users\ADMINI~1\AppData\Local\Temp\go-build279387310\b001
	)
	storer, err = NewStoreActor(filepath.Join(cwd, "sqlite3.db"), []string{cwd})
	if nil != err {
		log.Fatal(err)
	}
}

// go test -v -run=TestFull
func TestFull(t *testing.T) {
	TestWriteReadBlock(t)
	TestWriteReadTx(t)
	TestWriteReadTxData(t)

	full_tx, err := storer.ReadFullTx("id")
	if nil != err {
		t.Fatal(err)
	}
	t.Logf("==== ==== %#v", full_tx)

	blks, err := storer.ReadFullBlockByHeight(1e2)
	if nil != err {
		t.Fatal(err)
	}
	t.Logf("==== ==== %#v", blks)

	blk, err := storer.ReadFullBlockByHash("indep_hash")
	if nil != err {
		t.Fatal(err)
	}
	t.Logf("==== ==== %#v", blk)
}

// go test -v -run=TestWriteReadBlock
func TestWriteReadBlock(t *testing.T) {
	block := &types.TbBlock{
		IndepHash:     "indep_hash",
		Hash:          "hash",
		HashList:      "hash_list",
		Height:        1e2,
		PreviousBlock: "previous_block",
		Nonce:         "nonce",
		Timestamp:     1593425138,
		LastRetarget:  1593425138,
		Diff:          "Diff",
		RewardAddr:    "reward_addr",
		RewardPool:    1e4,
		WeaveSize:     1e5,
		BlockSize:     1e6,
		Tags:          "tags",
		WalletList:    "wallet_list",
	}
	err := storer.WriteBlock(block)
	if nil != err {
		t.Fatal(err)
	}

	blocks, err := storer.ReadBlockByHeight(block.Height)
	if nil != err {
		t.Fatal(err)
	}
	t.Logf("%#v", blocks)

	block2, err := storer.ReadBlockByHash(block.IndepHash)
	if nil != err {
		t.Fatal(err)
	}
	t.Logf("%#v", block2)
}

// go test -v -run=TestWriteReadTx
func TestWriteReadTx(t *testing.T) {
	txs := []types.TbTransaction{
		types.TbTransaction{
			ID:             "id",
			BlockIndepHash: "indep_hash",
			LastTx:         "last_tx",
			Owner:          "owner",
			FromAddress:    "from_address",
			Target:         "target",
			Quantity:       1e2,
			Signature:      "signature",
			Reward:         1e3,
			Tags:           "tags",
		},
	}
	err := storer.WriteTxs(txs)
	if nil != err {
		t.Fatal(err)
	}
	err = storer.UpdateTxsRefBlock([]string{"id"}, "update_block_hash")
	if nil != err {
		t.Fatal(err)
	}

	tx2, err := storer.ReadTx(txs[0].ID)
	if nil != err {
		t.Fatal(err)
	}
	t.Logf("%#v", tx2)
}

// go test -v -run=TestWriteReadTxData
func TestWriteReadTxData(t *testing.T) {
	err := storer.WriteTxData("id", []byte("1234567890"))
	if nil != err {
		t.Fatal(err)
	}

	buf, err := storer.ReadTxData("id")
	if nil != err {
		t.Fatal(err)
	}
	t.Logf("%s", buf)
}

// go test -v -run=TestWallet
func TestWallet(t *testing.T) {
	err := storer.UpdateWallet("address", "pub_key", "last_tx", 1e8)
	if nil != err {
		log.Fatal(err)
	}

	wl, err := storer.ReadWalletList()
	if nil != err {
		log.Fatal(err)
	}
	t.Logf("%v", wl)

	w1, err := storer.ReadWalletByAddress("address")
	if nil != err {
		log.Fatal(err)
	}
	t.Logf("%#v", w1)
}

// go test -v -run=TestChainTips
func TestChainTips(t *testing.T) {
	err := storer.UpdateChainTips([]types.TbChain{
		types.TbChain{
			IndepHash: "indep_hash",
			Height:    1e2,
			HashList:  "hash_list",
			IsActive:  1,
		},
	})
	if nil != err {
		log.Fatal(err)
	}

	tips, err := storer.ReadChainTips()
	if nil != err {
		log.Fatal(err)
	}
	t.Logf("%#v", tips)
}
