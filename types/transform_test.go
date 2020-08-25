package types

import (
	"fmt"
	"testing"
)

func TestTransform(t *testing.T) {
	block := &Block{
		// Txs: []string{"HSC684fCN7ZdeFuqHVis8hQhTvNb7HFVzLWAvDMUw8Xe", "HSC684fCN7ZdeFuqHVis8hQhTvNb7HFVzLWAvDMUw8Xe"},
	}

	b := block.TbBlock()

	t.Logf("%#v", b.Txs)

	b2 := b.Block()

	t.Logf("%#v", b2.Txs)
}

func TestTags(t *testing.T) {
	tag := Tag{
		Name:  "key",
		Value: "value",
	}
	t.Logf("%x\n", tag.BinaryEncode())
}

func TestTxPayout(t *testing.T) {
	for _, h := range []int64{1, 10e4, 100e4, 1000e4, 1e8} {
		pe_cost, nyear, gby_now := PecostGb(h)
		fmt.Printf("h=%d, year=%d, gby=%.10f, pe_cost=%.10f SWR:\n", h, nyear, float64(gby_now)/1e10, float64(pe_cost)/1e10)
		for _, dsize := range []int64{0, 1 << 20, 10 << 20, 1 << 30} {
			endowFee, minerFee := TxPayout(h, dsize, false)

			fmt.Printf("\tdsize=% 10d, \te=% 10d + m=% 10d \t ==> p=%.10f SWR\n", dsize, endowFee, minerFee, float64(endowFee+minerFee)/1e10)
		}
	}
}

func TestSortWalletTuples(t *testing.T) {
	arr := []WalletTuple{
		WalletTuple{
			Wallet: "9Wb3To8Gz9RwDLM5NRS1byfumf4A6GuF6nK3ne11wRAZ",
		},
		WalletTuple{
			Wallet: "4PbtQbcAMRtPPUDu2ocUnyVYB8PpDHSnNh5QavQ4krcb",
		},
		WalletTuple{
			Wallet: "EG8niAmff3BMnTz9HEWDpFVBTXWkNxeK4aCVVDfQKUt9",
		},
		WalletTuple{
			Wallet: "47UATwSi9BJLHEnWh7pwMF1sjBgp4rPNVpiS9jYn3XRm",
		},
	}

	SortWalletTuples(arr)

	for _, val := range arr {
		t.Logf("%s\n", val.Wallet)
	}
}
