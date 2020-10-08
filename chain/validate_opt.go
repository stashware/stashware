package chain

import (
	"stashware/oo"
	"stashware/types"
	"stashware/wallet"
)

var (
	max_deviation = int64(types.JOIN_CLOCK_TOLERANCE*2 + types.CLOCK_DRIFT_MAX)
)

func verifyTimestamp(block *types.Block, base_ts int64) (err error) {
	// var (
	// 	up_limit   = base_ts + max_deviation
	// 	down_limit = base_ts - types.MINING_TIMESTAMP_REFRESH_INTERVAL - types.MAX_BLOCK_PROPAGATION_TIME - max_deviation
	// )
	// if block.Timestamp > up_limit || block.Timestamp < down_limit {
	// 	err = Errorf("block.Timestamp[%d] out of [%d, %d]",
	// 		block.Timestamp, down_limit, up_limit)
	// 	return
	// }

	nowts := oo.TimeNowUnix()
	if block.Timestamp <= base_ts || block.Timestamp > nowts+types.MAX_BLOCK_PROPAGATION_TIME {
		err = Errorf("block[%d].Timestamp[%d] <= base_ts[%d] || block.Timestamp[%d] > nowts[%d]+MAX_BLOCK_PROPAGATION_TIME",
			block.Height, block.Timestamp, base_ts, block.Timestamp, nowts)
		return
	}

	return
}

func verifyWalletList(block *types.Block, txs []*types.Transaction, miner_reward int64) (err error) {
	wl, err := CalcWalletList(block, txs)
	if nil != err {
		err = Errorf("CalcWalletList %v", err)
		return
	}

	if len(wl) != len(block.WalletList) {
		err = Errorf("len(wl)[%d] != len(block.WalletList)[%d]",
			len(wl), len(block.WalletList))
		return
	}

	for i := range wl {
		if block.RewardAddr == wl[i].Wallet {
			wl[i].Quantity += miner_reward
		}
		if wl[i].Wallet != block.WalletList[i].Wallet ||
			wl[i].Quantity != block.WalletList[i].Quantity ||
			wl[i].LastTx != block.WalletList[i].LastTx {
			err = Errorf("not equal calc[%#v] args[%#v]",
				wl, block.WalletList)
			return
		}
	}

	return
}

func verifyRewardAddr(block *types.Block) (err error) {
	return
}

func VerifyTxData(tx *types.Transaction) (data_len int64, verify bool) {
	if "" == tx.Target && "" == tx.Data {
		oo.LogD("VerifyTxData err target and data are empty.")
		return
	}
	var (
		data_raw = tx.RawData()
	)
	data_len = int64(len(data_raw))

	if data_len > types.TX_DATA_SIZE_LIMIT {
		oo.LogD("VerifyTxData err data_len[%d] > types.TX_DATA_SIZE_LIMIT[%d]",
			data_len, types.TX_DATA_SIZE_LIMIT)
		return
	}
	if "" != tx.DataHash || "" != tx.Data {
		if 0 == data_len {
			oo.LogD("VerifyTxData tx.DataHash[%s], data_raw[%x]", tx.DataHash, data_raw)
			return
		}
		data_hash := oo.HexEncToStringPad32(oo.Sha256(data_raw))
		if data_hash != tx.DataHash {
			oo.LogD("VerifyTxData data_hash[%s] != tx.DataHash[%s]", data_hash, tx.DataHash)
			return
		}
	}

	return data_len, true
}

func VerifyTxSign(tx *types.Transaction) (verify bool) {
	var (
		buf = tx.SignatureDataSegment()
		msg = oo.Sha256(buf)
	)
	if !wallet.VerifySignature(tx.Owner, tx.Signature, msg) {
		oo.LogD("VerifyTxSign err VerifySignature fail owner[%s] signature[%s] msg[%0x] buf[%0x]",
			tx.Owner, tx.Signature, msg, buf)
		return
	}

	if tx.ID != oo.HexEncToStringPad32(msg) {
		oo.LogD("VerifyTxSign err tx.ID and tx.Signature not match")
		return
	}

	return true
}
