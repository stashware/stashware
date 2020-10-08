package txpool

import (
	"stashware/oo"
	"stashware/storage"
	"stashware/types"
)

func TxpoolInit(gconf *oo.Config) (err error) {
	err = storage.GStore.TxPoolPrune()
	if nil != err {
		oo.LogD("TxpoolInit TxPoolPrune err %v", err)
		return
	}
	return
}

func Ids() (ret []string) {
	ret, err := storage.GStore.TxPoolGet()
	if nil != err {
		oo.LogD("TxPoolGet err %v", err)
		return
	}

	return
}

func Count() int64 {
	return int64(len(Ids()))
}

func AddTx(tx *types.Transaction) {
	err := storage.GStore.TxPoolAdd(tx.TbTransaction())
	if nil != err {
		oo.LogD("AddTx TxPoolAdd err %v", err)
		return
	}

	return
}

func GetForMine() (txs []*types.Transaction) {
	arr, err := storage.GStore.TxPoolGet()
	if nil != err {
		oo.LogD("GetForMine TxPoolGet err %v", err)
		return
	}
	for _, tx_id := range arr {
		tx, err := storage.GStore.ReadFullTx(tx_id)
		if nil != err {
			oo.LogD("GetForMine ReadFullTx err %v", err)
			return
		}
		txs = append(txs, tx.Transaction())
	}

	return
}

func RemoveTxs(txids []string) {
	err := storage.GStore.TxPoolDel(txids)
	if nil != err {
		oo.LogD("RemoveTxs TxPoolDel err %v", err)
		return
	}

	err = storage.GStore.TxPoolPrune()
	if nil != err {
		oo.LogD("RemoveTxs TxPoolPrune err %v", err)
		return
	}

	return
}
