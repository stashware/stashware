package chain

import (
	"stashware/oo"
	"stashware/storage"
	"stashware/types"
)

//height: now height
func GetMinerPledgeNeed(height int64, addr string) (needed_swr int64, err error) {
	gs := storage.GStore
	if gs == nil {
		return
	}

	base := types.PledgeBase(height)
	nlock := types.LockBlocks(height)

	h2 := height
	h1 := h2 - nlock
	if h2 < nlock {
		h1 = 0
	}
	mined_blocks, err := gs.GetAddrMinedHeights(addr, h1, h2, types.NETWORK_MAINNET)
	if err != nil {
		oo.LogD("XXX err %v", err)
		return
	}
	needed_swr = base * (1 + int64(len(mined_blocks)))

	oo.LogD("ZZZ h=%d addr=%s ==> %d * %d = %d", height, addr, base, 1+int64(len(mined_blocks)), needed_swr)
	return
}

//height-locked_height ~ height
func GetAddrLockedBalance(height int64, addr string) (last_mined_h int64, locked_swr int64, err error) {
	gs := storage.GStore
	if gs == nil {
		err = oo.NewErrno(oo.ESERVER)
		return
	}

	h2 := height
	nlock := types.LockBlocks(height)
	h1 := h2 - nlock
	if h2 < nlock {
		h1 = 0
	}
	mined_blocks, err := gs.GetAddrMinedHeights(addr, h1, h2, types.NETWORK_MAINNET)
	if err != nil {
		oo.LogD("YYY err %v", err)
		return
	}
	if len(mined_blocks) == 0 {
		return
	}
	last_h := mined_blocks[len(mined_blocks)-1]
	last_nlock := types.LockBlocks(last_h)
	if last_h+last_nlock < height {
		return //leave lock
	}

	last_mined_h = last_h
	locked_swr, err = GetMinerPledgeNeed(last_h, addr)
	// if err == nil {
	// 	locked_swr -= types.PledgeBase(last_h)
	// }

	return
}
