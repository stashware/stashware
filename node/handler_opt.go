package node

import (
	// "errors"
	"sync"
)

var (
	submit_mutex sync.Mutex
	txset        sync.Map // id => (has process)
	badtx        sync.Map // peer-addr_id => (has process)
)

func setTxSet(id string) {
	txset.Store(id, 1)
}

func getTxSet(id string) (ok bool) {
	_, ok = txset.Load(id)

	return
}

func setBadTx(addr, id string) {
	badtx.Store(addr+"_"+id, 1)
}

func getBadTx(addr, id string) (ok bool) {
	_, ok = badtx.Load(addr + "_" + id)

	return
}
