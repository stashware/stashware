package chain

import (
	"stashware/types"
	"testing"
)

func TestPledge(t *testing.T) {
	hs := []int64{100, 77321, 100000, 177321}
	for i := 0; i < len(hs); i++ {
		x := types.PledgeBase(hs[i])
		y := types.LockBlocks(hs[i])
		t.Logf("%d: %d >= %d blocks %d", i, hs[i], x, y)
	}

}
