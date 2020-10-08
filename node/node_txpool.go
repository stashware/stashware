package node

import (
	"time"

	"stashware/chain"
	"stashware/oo"
	"stashware/storage"
	"stashware/types"
)

func NodeSyncTxpool() {
	ticker := time.NewTicker(time.Second * 30)
	defer ticker.Stop()

	for range ticker.C {
		syncTxpool()
	}
}

func syncTxpool() {
	block := chain.TipBlock()
	if nil == block {
		oo.LogD("syncTxpool chain.TipBlock() == nil")
		return
	}

	var (
		top int = 5
	)
	_, wss := PeerTopN(top)

	for _, ws := range wss {
		txids, err := RpcReadTxPending(ws)
		if nil != err {
			oo.LogD("syncTxpool RpcReadTxPending err %v", err)
			continue
		}
		for _, id := range txids {
			if true == getTxSet(id) {
				continue
			}
			if true == getBadTx(ws.PeerAddr(), id) {
				continue
			}
			db_tx, err := storage.GStore.ReadTx(id)
			if nil != err {
				if storage.ErrNoRows != err {
					oo.LogD("syncTxpool ReadTx err %v", err)
					continue
				}

				tx, err := RpcReadTxAndData(ws, id)
				if nil != err {
					oo.LogD("syncTxpool RpcReadTxAndData err %v", err)
					continue
				}

				err = ProcessSubmitTx(nil, tx)
				if nil != err {
					if types.TX_VERIFY_ERROR == err {
						setBadTx(ws.PeerAddr(), id)
					}
					oo.LogD("syncTxpool ProcessSubmitTx err %v", err)
					continue
				}
			}
			_ = db_tx
		}
	}
}
