package chain

import (
	"fmt"

	// "stashware/miner"
	"stashware/oo"
	"stashware/storage"
	"stashware/types"
	"stashware/wallet"
)

var Errorf = fmt.Errorf

func VerifyBlock(block *types.Block) (verify bool) {
	// oo.LogD("new block %d, r=%d", block.Height, block.LastRetarget)
	err := oo.ValidateStruct(block)
	if nil != err {
		oo.LogD("VerifyBlock err ValidateStruct %v", err)
		return
	}

	old_blk, err := GetBlock(block.PreviousBlock)
	if nil != err {
		oo.LogD("VerifyBlock err GetBlock %v", err)
		return
	}

	// verify height
	if block.Height != old_blk.Height+1 {
		oo.LogD("VerifyBlock err block.Height[%d] != old_blk.Height+1[%d]", block.Height, old_blk.Height+1)
		return
	}

	var (
		txs        []*types.Transaction
		block_size int64
	)
	for _, tx_id := range block.Txs {
		full_tx, err := GetFullTx(tx_id)
		if nil != err {
			oo.LogD("VerifyBlock err GetFullTx %v", err)
			return
		}
		txs = append(txs, full_tx.Transaction())
		block_size += int64(len(full_tx.Data))
	}

	// verify weave_size
	// verify block_size
	if old_blk.WeaveSize+block_size != block.WeaveSize {
		oo.LogD("VerifyBlock %d err old_blk.WeaveSize[%d]+block_size[%d] != block.WeaveSize[%d]",
			block.Height, old_blk.WeaveSize, block_size, block.WeaveSize)
		return
	}
	if block_size != block.BlockSize {
		oo.LogD("VerifyBlock err block_size[%d] != block.BlockSize[%d]",
			block_size, block.BlockSize)
		return
	}

	// verify nonce

	// verify previous_block

	// verify timestamp
	err = verifyTimestamp(block, old_blk.Timestamp)
	if nil != err {
		oo.LogD("VerifyBlock err verifyTimestamp %v", err)
		return
	}

	// verify last_retarget
	last_retarget := CalcLastRetarget(block, old_blk.LastRetarget)
	if last_retarget != block.LastRetarget {
		oo.LogD("VerifyBlock %d err last_retarget[%d] != block.LastRetarget[%d]",
			block.Height, last_retarget, block.LastRetarget)
		return
	}

	// verify diff
	var (
		big_diff = block.BigDiff()
		big_hash = block.BigHash()
	)
	if big_hash.Cmp(big_diff) <= 0 {
		oo.LogD("VerifyBlock err big_hash[%064x] <= big_diff[%064x]",
			big_hash.Bytes(), big_diff.Bytes())
		return
	}
	next_diff := CalcNextRequiredDifficulty(old_blk)
	if 0 != next_diff.Cmp(big_diff) {
		if block.Height > types.Fork_2000_0920 {
			oo.LogD("VerifyBlock err next_diff[%064x] != big_diff[%064x]. old_blk[%d %x, diff %x, lastretarget %d, ts %d]",
				next_diff.Bytes(), big_diff.Bytes(), old_blk.Height, old_blk.IndepHash,
				old_blk.BigDiff().Bytes(), old_blk.LastRetarget, old_blk.Timestamp)
			return
		}
		oo.LogD("verify failed but fork point %d", block.Height)
	}

	// verify cumulative_diff
	next_c_diff := CalcNextCumulativeDiff(old_blk.CumulativeDiff, big_diff)
	if block.CumulativeDiff != next_c_diff {
		oo.LogD("VerifyBlock err block.CumulativeDiff[%d] != next_c_diff[%d]",
			block.CumulativeDiff, next_c_diff)
		return
	}

	var (
		bds_buf     = block.DataSegment()
		nonce_buf   = oo.HexDecodeString(block.Nonce)
		hash_buf    = oo.HexDecStringPad32(block.Hash)
		height_buf  = oo.IntToBytes(block.Height)
		pre_blk_buf = oo.HexDecStringPad32(block.PreviousBlock)
	)

	// verify dep_hash
	dep_hash := oo.HexEncToStringPad32(CalcDepHash(pre_blk_buf, bds_buf, nonce_buf, height_buf))
	if dep_hash != block.Hash {
		oo.LogD("VerifyBlock err dep_hash[%s] != block.Hash[%s]", dep_hash, block.Hash)
		return
	}

	// verify indep_hash
	indep_hash := oo.HexEncToStringPad32(CalcIndepHash(bds_buf, hash_buf, nonce_buf))
	if indep_hash != block.IndepHash {
		oo.LogD("VerifyBlock err indep_hash[%s] != block.IndepHash[%s]", indep_hash, block.IndepHash)
		return
	}

	// verify tx_root
	tx_root := CalcTxRootString(block)
	if tx_root != block.TxRoot {
		oo.LogD("VerifyBlock err tx_root[%s] != block.TxRoot[%s]", tx_root, block.TxRoot)
		return
	}

	// verify reward_addr
	err = verifyRewardAddr(block)
	if nil != err {
		oo.LogD("VerifyBlock err verifyRewardAddr %v", err)
		return
	}

	// verify reward_fee
	// verify reward_pool
	miner_reward, reward_pool, err := CalcRewardPool(block, old_blk.RewardPool)
	if nil != err {
		oo.LogD("VerifyBlock err CalcRewardPool %v", err)
		return
	}
	if block.RewardPool != reward_pool {
		oo.LogD("VerifyBlock err block.RewardPool[%d] != reward_pool[%d]",
			block.RewardPool, reward_pool)
		return
	}
	if block.RewardFee != miner_reward {
		oo.LogD("VerifyBlock err block.RewardFee[%d] != miner_reward[%d]",
			block.RewardFee, miner_reward)
		return
	}

	// verify wallet_list
	err = verifyWalletList(block, txs, miner_reward)
	if nil != err {
		oo.LogD("VerifyBlock %d %s err verifyWalletList %v", block.Height, block.IndepHash, err)
		return
	}

	// verify wallet_list_hash
	wl_hash := oo.HexEncToStringPad32(block.HashWalletList())
	if wl_hash != block.WalletListHash {
		oo.LogD("VerifyBlock err wl_hash[%s] != block.WalletListHash[%s]", wl_hash, block.WalletListHash)
		return
	}

	// verify tx
	if !VerifyBlockTxs(block.Height, txs[:]) {
		return
	}

	// no need to verify poa

	//verify pos
	if !VerifyPos(block.Height, block.RewardAddr) {
		oo.LogD("VerifyBlock err POS %d %s", block.Height, block.RewardAddr)
		return
	}

	return true
}

func VerifyTx(height int64, tx *types.Transaction, txs []*types.Transaction) (verify bool) {
	data_len, check := VerifyTxData(tx)
	if false == check {
		return
	}

	from := wallet.StringPublicKeyToAddress(tx.Owner)
	if from == tx.Target {
		oo.LogD("VerifyTx err sender == receiver")
		return
	}
	w0, err := storage.GStore.ReadWalletByAddress(from)
	if nil != err {
		oo.LogD("VerifyTx err storage.ReadWalletByAddress(%s) %v", from, err)
		return
	}

	var (
		target           = tx.Target
		last_tx          = w0.LastTx
		send_pending_amt int64
		is_new_addr      bool
	)
	if "" != target {
		var w1 *types.TbWallet
		w1, err = storage.GStore.ReadWalletByAddress(target)
		if nil != err {
			if storage.ErrNoRows != err {
				oo.LogD("VerifyTx err storage.ReadWalletByAddress(%s) %v", target, err)
				return
			}
			is_new_addr = true

		} else {
			if "" == w1.LastTx {
				is_new_addr = true
			}
		}
	}
	for _, val := range txs {
		if from == val.From {
			last_tx = val.ID
			send_pending_amt += val.Quantity
		}
		if "" != target {
			if val.Target == target || val.From == target {
				is_new_addr = false
			}
		}
	}

	if last_tx != tx.LastTx {
		oo.LogD("VerifyTx err last_tx[%s] != tx.LastTx[%s]", last_tx, tx.LastTx)
		return
	}

	_, lock_amt, err := GetAddrLockedBalance(height, from)
	if nil != err {
		oo.LogD("VerifyTx err GetAddrLockedBalance(%d, %s) %v", height, from, err)
		return
	}
	//swr -->
	lock_amt = lock_amt * types.MINTOKEN_PER_TOKEN

	endow_fee, miner_fee := types.TxPayout(height, data_len, is_new_addr)
	if tx.Reward < endow_fee {
		oo.LogD("VerifyTx err tx_too_cheap tx.Reward[%d] < endow_fee[%d]", tx.Reward, endow_fee)
		return
	}
	need_fee := endow_fee + miner_fee
	if w0.Balance-lock_amt-send_pending_amt < tx.Quantity+need_fee {
		oo.LogD("VerifyTx err w0.Balance[%d]-lock_amt[%d]-send_pending_amt[%d] < tx.Quantity[%d]+need_fee[%d]", w0.Balance, lock_amt, send_pending_amt, tx.Quantity, need_fee)
		return
	}

	if false == VerifyTxSign(tx) {
		return
	}

	return true
}

func VerifySubmitTx(height int64, tx *types.Transaction) (verify bool) {
	ids, err := storage.GStore.TxPoolGet()
	if nil != err {
		oo.LogD("VerifySubmitTx ReadPendingTxs %v", err)
		return
	}

	var (
		txs []*types.Transaction
	)

	for _, id := range ids {
		full_tx, err := GetFullTx(id)
		if nil != err {
			oo.LogD("VerifySubmitTx GetFullTx(%s) %v", id, err)
			return
		}
		txs = append(txs, full_tx.Transaction())
	}

	if false == VerifyTx(height, tx, txs) {
		return
	}

	return true
}

func VerifyBlockTxs(height int64, txs []*types.Transaction) (verify bool) {
	for i, tx := range txs {
		if false == VerifyTx(height, tx, txs[0:i]) {
			return
		}
	}

	return true
}

func VerifyPos(height int64, addr string) (verify bool) {
	need_swr, err := GetMinerPledgeNeed(height, addr)
	if err != nil {
		oo.LogD("Failed to get Miner need. err=%v", err)
		return
	}

	if need_swr > 0 {
		w, err := storage.GStore.ReadWalletByAddress(addr)
		if err != nil {
			oo.LogD("h %d, addr not active: %s", height, addr)
			return
		}
		if w.Balance < need_swr*types.MINTOKEN_PER_TOKEN {
			oo.LogD("h %d, balance %d < need %d, skip mined", height, w.Balance/types.MINTOKEN_PER_TOKEN, need_swr)
			return
		}
	}
	verify = true
	return
}
