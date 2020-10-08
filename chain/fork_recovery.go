package chain

import (
	"stashware/oo"
	"stashware/storage"
	"stashware/types"
	"stashware/wallet"
)

//fork_recovery.go

func ChainFindIt(css []types.TbChain, indep_hash string) (idx int) {
	idx = -1
	for i, cs := range css {
		if cs.IndepHash == indep_hash {
			idx = i
			break
		}
	}
	return
}
func ChainFindActive(css []types.TbChain) (idx int) {
	idx = -1
	for i, cs := range css {
		if cs.IsActive == int64(1) {
			idx = i
			break
		}
	}
	return
}

func RollbackBlock(nowb *types.Block) (err error) {
	oo.LogD("RollbackBlock to %d => %s", nowb.Height, nowb.IndepHash)

	txs, err := storage.GStore.ReadTxsByBlockHash(nowb.IndepHash)
	if nil != err {
		err = oo.NewError("ReadTxsByBlockHash %v", err)
	}
	for i, j := 0, len(txs)-1; i < j; i, j = i+1, j-1 {
		txs[i], txs[j] = txs[j], txs[i]
	}
	r_txs := txs

	wmap := map[string]*types.TbWallet{}
	wmap[nowb.RewardAddr] = &types.TbWallet{
		Address: nowb.RewardAddr,
		Balance: -nowb.RewardFee,
	}
	for _, tx := range r_txs {
		oaddr := wallet.StringPublicKeyToAddress(tx.Owner)
		new_w := &types.TbWallet{
			Address: oaddr,
			Balance: tx.Quantity + tx.Reward,
			LastTx:  tx.LastTx, //fixed
		}
		if wt, ok := wmap[oaddr]; ok {
			wt.Balance += new_w.Balance
			wt.LastTx = tx.LastTx
		} else {
			wmap[oaddr] = new_w
		}

		if len(tx.Target) > 0 {
			new_w = &types.TbWallet{
				Address: tx.Target,
				Balance: -tx.Quantity,
				LastTx:  "NO_LAST_TX",
			}
			if wt, ok := wmap[tx.Target]; ok {
				wt.Balance += new_w.Balance
			} else {
				wmap[tx.Target] = new_w
			}
		}
	}

	//rollback wallet_list
	for _, w1 := range nowb.WalletList {
		var w *types.TbWallet
		w, ok := wmap[w1.Wallet]
		if !ok {
			oo.LogW("block error..., w=%s", w1.Wallet)
			return
		}
		w.Balance += w1.Quantity

		if "NO_LAST_TX" == w.LastTx {
			w.LastTx = w1.LastTx
		}

		if err := storage.GStore.UpdateWallet(w.Address, w.LastTx, w.Balance); err != nil {
			oo.LogW("Failed to update wallet. %v, err:%v", w, err)
		}
	}

	//rollback tx.indep
	if err = storage.GStore.UpdateTxsRefBlock(nowb.Txs, ""); err != nil {
		return
	}

	//remove main net flag
	tb := nowb.TbBlock()
	tb.Network = types.NETWORK_TESTNET
	if err = storage.GStore.UpdateBlockFlag(tb); err != nil {
		return
	}

	return
}

func (this *BlockChain) ForkChain(b *types.Block) (err error) {

	new_block_mutex.Lock()
	defer new_block_mutex.Unlock()

	tipb := *TipBlock()
	if b.IndepHash == tipb.IndepHash {
		oo.LogD("found the same tip. cancel fork chain. %d %s",
			b.Height, b.IndepHash)

		return
	}

	var nowb *types.TbBlock
	for pre := tipb.IndepHash; pre != b.IndepHash; pre = nowb.PreviousBlock {
		if nowb, err = storage.GStore.ReadBlockByHash(pre); err != nil {
			oo.LogW("Failed to read block %s, err %v", pre, err)
			return
		}
		if err = RollbackBlock(nowb.Block()); err != nil {
			oo.LogW("Failed to rollback %s, err %v", pre, err)
			return
		}
	}
	oo.LogD("Succeed to rollback, from %d %s to %d %s", tipb.Height, tipb.IndepHash, b.Height, b.IndepHash)

	if err = GChain.StepTip(b); err != nil {
		oo.LogW("Failed to update tip %d %s, err %v", b.Height, b.IndepHash, err)
	}

	return
}

func TrimBlock(height int64) (err error) {
	abs, err := storage.GStore.ReadActiveBlocks(int64(0), height)
	if err != nil {
		return
	}
	oo.LogD("prepare trim to: %v", abs)
	if len(abs) == 0 {
		return
	}

	if err = storage.GStore.ClearTips(); err != nil {
		oo.LogW("Failed to clear tips. err %v", err)
		return
	}

	tb, err := storage.GStore.ReadBlockByHash(abs[0].IndepHash)
	if err != nil {
		oo.LogD("failed to read genesis block %s, err %v", abs[0].IndepHash, err)
		return
	}
	ChainSaveInit(tb.Block())

	for i := 1; i < len(abs); i++ {
		ab := &abs[i]
		if int64(i) != ab.Height {
			oo.LogD("height %d != index %d", ab.Height, i)
			return
		}
		tb, err = storage.GStore.ReadBlockByHash(ab.IndepHash)
		if err != nil {
			oo.LogD("Failed to read block %s, err %v", ab.IndepHash, err)
			return
		}
		b := tb.Block()

		var poa []byte
		poa, err = GChain.GetPoa(oo.HexDecStringPad32(b.PreviousBlock), b.Height)
		if err != nil {
			oo.LogW("Failed to calc poa, h=%d, prehash=%s. err %v", b.Height, b.PreviousBlock, err)
			return
		}
		if len(poa) > 0 {
			b.Poa = oo.HexEncodeToString(poa)
		}
		var BDS []byte
		BDS, err = GChain.GetBDS(b)
		if err != nil {
			oo.LogW("Failed to calc bds,  h=%d, prehash=%s. err %v", b.Height, b.PreviousBlock, err)
			return
		}
		GChain.OnNewBlock(b, BDS, false)
	}

	return
}
