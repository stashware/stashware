package types

import (
	"stashware/oo"
)

const (
	RequiredMinVersion = "v0.8.0"
)
const ( //fork height
	Fork_2000_0920 = int64(6999) ///
)

type CheckPoint = struct {
	Hash   string
	Reject bool
}

type CheckPointsMap = map[int64]CheckPoint

var AcceptCPs = CheckPointsMap{
	7000: CheckPoint{
		Hash:   "b71b8e3afeebc796e7a5d8d778f47b27b87ad06d419151ff32bb6671b2cd7598",
		Reject: true,
	},
}

func CheckCps(h int64, indepHash string) (wl bool, failed bool) {
	if cp, ok := AcceptCPs[h]; ok {
		if !cp.Reject && cp.Hash != indepHash { //force that hash
			oo.LogD("CHECK POINT %d %s, skip %s", h, cp.Hash, indepHash)
			failed = true
			return
		}
		if cp.Hash == indepHash && cp.Reject { //force not that hash
			oo.LogD("CHECK POINT %d, reject %s", h, indepHash)
			failed = true
			return
		}
		if cp.Hash == indepHash && !cp.Reject { //equ hash, accept
			oo.LogD("CHECK POINT %d, accept %s", h, indepHash)
			wl = true
			return
		}
		//accept
	}
	//accept
	return
}
