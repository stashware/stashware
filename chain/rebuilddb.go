package chain

import (
	"stashware/oo"
	"stashware/storage"
	"stashware/types"
)

// func TryVerifyAndRepairDb() (err error) {
// 	return
// }

// func VerifyTbBlocks() (idx int64, err error) {
//	cs, tip, err := GetActiveChain()
// 	for i := int64(0); i <= cs.Height; i += types.RETARGET_BLOCKS {
// 		max := i + types.RETARGET_BLOCKS
// 		if max > cs.Height {
// 			max = cs.Height
// 		}
// 		abs, err1 := storage.GStore.ReadActiveBlocks(i, cs.Height)
// 		if err1 != nil {
// 			err = oo.NewError("read activeblock %d - %d err %v", i, cs.Height, err1)
// 			return
// 		}
// 		if int64(len(abs)) != cs.Height-i+1 {
// 			err = oo.NewError("activeblock(%d, %d) %d != %d", i, cs.Height, len(abs), cs.Height-i+1)
// 			return
// 		}
// 		for j := i; j <= max; j++ {
// 			if abs[j-i].Height != i {
// 				err = oo.NewError("activeblock err(%d)", j)
// 				return
// 			}
// 			idx = j
// 		}
// 	}

// 	return
// }

func RebuildDb() (err error) {
	oo.LogD("REBUILDING DB ...")
	GChain.do_repair = true
	defer func() {
		GChain.do_repair = false
	}()

	if err = storage.GStore.ClearTips(); err != nil {
		oo.LogW("Failed to clear tips: %v", err)
		return
	}

	InitGenesisBlock()

	var pre *types.TbBlock
	if pre, err = storage.GStore.ReadBlockByHeight(0); err != nil {
		oo.LogD("no genesis block")
		return
	}
	preBlock := *pre

_For:
	for i := int64(1); ; /**/ i += types.BLOCKS_PER_DAY {
		imax := i + types.BLOCKS_PER_DAY - int64(1)
		oo.LogD("try rebuild %d ~ %d ...", i, imax)

		var tbs []types.TbBlock
		if tbs, err = storage.GStore.ReadBlockArrByHeight(i, imax); err != nil {
			oo.LogD("no block %d~%d", i, imax)
			err = nil
			break _For
		}
		for _, tb := range tbs {
			if preBlock.IndepHash != tb.PreviousBlock {
				oo.LogD("pre %d %s, now b %d pre %s", preBlock.Height, preBlock.IndepHash, tb.Height, tb.PreviousBlock)
				break _For
			}
			b := tb.Block()

			//checkpoint
			if _, failed := types.CheckCps(b.Height, b.IndepHash); failed {
				oo.LogD("Failed check point. %d %s", b.Height, b.IndepHash)
				break _For
			}

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
			if BDS, err = GChain.GetBDS(b); err != nil {
				oo.LogW("Failed to get BDS, h=%d, err %v", i, err)
				return
			}

			if err := GChain.OnNewBlock(b, BDS, false); err != nil {
				oo.LogD("Failed to add block %d %s", i, b.IndepHash)
				//err = nil
				break _For
			}
			preBlock = tb
		}
		if int64(len(tbs)) < types.BLOCKS_PER_DAY {
			// oo.LogD("block h=%d count=%d != 1", i, len(tbs))
			//err = nil
			break _For
		}
	}
	oo.LogD("REBUILDING DONE ==> %d %s", preBlock.Height, preBlock.IndepHash)
	return
}
