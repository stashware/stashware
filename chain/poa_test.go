package chain

import (
	// "fmt"
	"stashware/oo"
	"stashware/storage"
	"testing"
)

func TestGetPoa(t *testing.T) {
	t.Logf("poa====")

	var err error
	storage.GStore, err = storage.NewStoreActor("~/Downloads/stashware-v0.6.3-win64/data/stashware.db", []string{"~/Downloads/stashware-v0.6.3-win64/data/txs"})
	if err != nil {
		t.Fatalf("err %v", err)
	}
	poa, err := GChain.GetPoa(oo.HexDecStringPad32("ca5a3f3745ec79a0e344d8c50246ee860dc7b2adfe6a30779959ccd963731a1f"), 9013)
	if err != nil {
		t.Fatalf("err %v", err)
	}
	t.Logf("poa =: %x", poa)
	_ = storage.GStore
}

func TestDepHash(t *testing.T) {
	var err error
	storage.GStore, err = storage.NewStoreActor("~/Downloads/stashware-v0.6.3-win64/data/stashware.db", []string{"~/Downloads/stashware-v0.6.3-win64/data/txs"})
	if err != nil {
		t.Fatalf("err %v", err)
	}

	tb, err := storage.GStore.ReadBlockByHeight(9013)
	if err != nil {
		t.Fatalf("err %v", err)
	}
	t.Logf("block %v", *tb)

	poa, err := GChain.GetPoa(oo.HexDecStringPad32(tb.PreviousBlock), tb.Height)
	if err != nil {
		t.Fatalf("err %v", err)
	}

	b := tb.Block()
	if len(poa) > 0 {
		b.Poa = oo.HexEncodeToString(poa)
	}
	t.Logf("poa %8x", poa)

	BDS, err := GChain.GetBDS(b)
	if err != nil {
		t.Fatalf("err %v", err)
	}
	t.Logf("BDS %8x", BDS)

	ok := VerifyBlock(b)
	if !ok {
		t.Fatalf("ok? %v", ok)
	}

}

// func TestMain(m *testing.M) {
// 	fmt.Printf("kkkk")
// }
