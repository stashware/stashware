package chain

import (
	"testing"

	"stashware/types"
)

func TestCalcTxRoot(t *testing.T) {
	block := &types.Block{
		Txs: []string{"GEgNkLvkRnM6sPWg8vxGjCdWrcE1qWpP86pyDvxr7zBr"},
	}

	buf := CalcTxRoot(block)

	t.Logf("%064x", buf)
}
