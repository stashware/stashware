package chain

import (
	"stashware/oo"
	"stashware/storage"
	"stashware/types"
	"stashware/wallet"
)

//
// impl calculate_reward_pool_perpetual
//
func CalcRewardPool(block *types.Block, old_pool int64) (finder_reward, reward_pool int64, err error) {
	txs_cost, txs_reward, err := CalcTxsCostReward(block)
	if nil != err {
		err = Errorf("CalcTxsCostReward %v", err)
		return
	}

	var (
		inflation   = types.Rinf(block.Height)
		base_reward = inflation + txs_reward

		cost_per_gb_per_block, _, _ = types.PecostGb(block.Height)
		burden                      = int64(float64(block.WeaveSize) * float64(cost_per_gb_per_block) / float64(1024*1024*1024))

		sw       = burden - base_reward
		new_pool = old_pool + txs_cost
	)

	if sw <= 0 {
		finder_reward, reward_pool = base_reward, new_pool

	} else {
		take := sw
		if take < new_pool {
			take = new_pool
		}
		finder_reward, reward_pool = base_reward+take, new_pool-take
	}

	return
}

func CalcTxsCostReward(block *types.Block) (txs_endow, txs_reward int64, err error) {
	for _, tx_id := range block.Txs {
		var tx *types.Transaction
		tx, err = GetTransaction(tx_id)
		if nil != err {
			err = Errorf("GetTransaction %v", err)
			return
		}
		total_fee := tx.Reward

		new_target := false
		if len(tx.Target) > 0 {
			if _, err := storage.GStore.ReadWalletByAddress(tx.Target); err != nil {
				new_target = true
			}
		}
		tx_endow, tx_miner_fee := types.TxPayout(block.Height, int64(len(tx.Data)), new_target)
		if tx_endow+tx_miner_fee > total_fee {
			err = Errorf("tx %s fee error %d+%d > %d", tx_id, tx_endow, tx_miner_fee, total_fee)
			return
		}
		tx_miner_fee = total_fee - tx_endow

		txs_endow += tx_endow
		txs_reward += tx_miner_fee
	}

	return
}

func CalcLastRetarget(block *types.Block, last_last_retarget int64) int64 {
	if 0 == block.Height%types.RETARGET_BLOCKS {
		return block.Timestamp
	}

	return last_last_retarget
}

func CalcTxRoot(block *types.Block) []byte {
	var arr [][]byte
	for _, tx_id := range block.Txs {
		arr = append(arr, oo.HexDecStringPad32(tx_id))
	}

	return GenerateMerkerRoot(oo.Sha256, arr)
}

func CalcTxRootString(block *types.Block) string {
	return oo.HexEncToStringPad32OrNil(CalcTxRoot(block))
}

func CalcWalletList(block *types.Block, txs []*types.Transaction) (wlist []types.WalletTuple, err error) {
	var wmap = map[string]*types.WalletTuple{}

	wmap[block.RewardAddr] = &types.WalletTuple{
		Wallet:   block.RewardAddr,
		Quantity: 0,
	}

	for _, tx := range txs {

		oaddr := wallet.StringPublicKeyToAddress(tx.Owner)

		new_w := &types.WalletTuple{
			Wallet:   oaddr,
			Quantity: -(tx.Quantity + tx.Reward),
			LastTx:   tx.ID,
		}
		if wt, ok := wmap[oaddr]; ok {
			wt.Quantity += new_w.Quantity
			wt.LastTx = new_w.LastTx
		} else {
			wmap[oaddr] = new_w
		}

		if len(tx.Target) > 0 {
			new_w = &types.WalletTuple{
				Wallet:   tx.Target,
				Quantity: +tx.Quantity,
				// LastTx: tx.ID,
			}
			if wt, ok := wmap[tx.Target]; ok {
				wt.Quantity += new_w.Quantity
			} else {
				wmap[tx.Target] = new_w
			}
		}
	}

	for addr, wt := range wmap {
		w, err1 := storage.GStore.ReadWalletByAddress(addr)
		if err1 != nil { //new wallet
			w = &types.TbWallet{
				Address: wt.Wallet,
				Balance: wt.Quantity,
				LastTx:  wt.LastTx,
			}
		} else {
			w.Balance += wt.Quantity
			if len(wt.LastTx) > 0 {
				w.LastTx = wt.LastTx
			}
		}

		if w.Balance < 0 {
			oo.LogW("logic failed. %s balance %d < 0", addr, w.Balance)
			err = err1
			return
		}
		wlist = append(wlist, types.WalletTuple{
			Wallet:   w.Address,
			Quantity: w.Balance,
			LastTx:   w.LastTx,
		})
	}

	if len(wlist) > 0 {
		types.SortWalletTuples(wlist)
	}

	return
}

func CalcDepHash(key, bds, nonce, height []byte) []byte {
	var (
		buf []byte
		cal []byte
	)
	buf = append(buf, bds...)
	buf = append(buf, nonce...)
	buf = append(buf, height...)

	cal, _ = oo.RandomX(key, buf)

	return cal
}

func CalcIndepHash(bds, hash, nonce []byte) []byte {
	var (
		buf []byte
	)
	buf = append(buf, bds...)
	buf = append(buf, hash...)
	buf = append(buf, nonce...)

	return oo.Sha256(buf)
}
