package types

import (
	"math"
)

const NETWORK_TESTNET = 0
const NETWORK_MAINNET = 1

const TOKEN_NAME = "SWR"
const MINTOKEN_NAME = "Stashy"
const MINTOKEN_PER_TOKEN = int64(1e10)

const TOKEN_TOTAL = int64(6e7)
const MINTOKEN_TOTAL = float64(TOKEN_TOTAL) * float64(MINTOKEN_PER_TOKEN)

const TARGET_SEC = 150

const RETARGET_BLOCKS = 50

const BLOCKS_PER_HOUR = int64(3600 / TARGET_SEC)
const BLOCKS_PER_DAY = (BLOCKS_PER_HOUR * 24)
const BLOCKS_PER_WEEK = (BLOCKS_PER_DAY * 7)
const BLOCKS_PER_MONTH = (BLOCKS_PER_DAY * 30)
const BLOCKS_PER_YEAR = (BLOCKS_PER_DAY * 365)
const BLOCKS_PER_4YEAR = (BLOCKS_PER_YEAR * 4)

const STAGE_1_H = int64(15 * BLOCKS_PER_DAY)
const STAGE_1_RINF = int64(1300 * MINTOKEN_PER_TOKEN)
const STAGE_2_H = int64(100000)        // (173*day + 15*day)
const STAGE_2_PLEDGE_BASE = int64(500) //500*X*(ln(h)/(ln(h)+1))
const STAGE_2_LOCKB = int64(10 * BLOCKS_PER_DAY)
const STAGE_3_PLEDGE_BASE = int64(800) //
const STAGE_3_LOCKB = int64(15 * BLOCKS_PER_DAY)

const BLOCK_HEAD_SIZE = int64(1000)
const BLOCK_TX_COUNT_LIMIT = 1000
const TX_SIZE_BASE = int64(3210)                   //
const TX_DATA_SIZE_LIMIT = int64(10 * 1024 * 1024) //10M

const MEMPOOL_DATA_SIZE_LIMIT = int64(1 << 29) //512M

const COST_GBY_2020 = int64(0.002 * float64(MINTOKEN_PER_TOKEN))
const DAY_2020_LEFT = int64(100)
const DECAY_ANNUAL_GBY = float64(0.995)

const WALLET_GEN_FEE = int64(0.2 * float64(MINTOKEN_PER_TOKEN)) //-->endow. reward=total_size*cost*0.4
const TX_ENDOW_RATE = float64(2.0)                              //(txbase+size)*PeCostByte * 2
const TX_REWARD_RATE = float64(0.4)                             //(txbase+size)*PeCostByte * 0.4

const TX_SEND_WITHOUT_ASKING_SIZE = 1000

const (
	RETARGET_TOLERANCE         = float64(0.1)
	DIFF_ADJUSTMENT_DOWN_LIMIT = float64(0.25)
	DIFF_ADJUSTMENT_UP_LIMIT   = float64(2)

	RETARGET_TOLERANCE_INT         = 1000  // *1e4
	DIFF_ADJUSTMENT_DOWN_LIMIT_INT = 2500  // *1e4
	DIFF_ADJUSTMENT_UP_LIMIT_INT   = 20000 // *1e4
)

const (
	JOIN_CLOCK_TOLERANCE              = 15 // seconds
	MAX_BLOCK_PROPAGATION_TIME        = 60 // seconds
	CLOCK_DRIFT_MAX                   = 5  // seconds
	MINING_TIMESTAMP_REFRESH_INTERVAL = 10 // seconds
)

const (
	// Mining reward as a proportion of tx cost.
	MINING_REWARD_MULTIPLIER = float64(0.2)
)

const (
	MIN_TRUSTED_NODES = 6
	MAX_TRUSTED_NODES = 12

	MIN_PEERS = 1
	MAX_PEERS = 32
)

const (
	DEFAULT_SYNC_PORT = 3080
	// DEFAULT_GATEWAY_PORT    = 2080
	// DEFAULT_WEB_WALLET_PORT = 1080
)

func PecostGb(h int64) (pe_cost int64, nyear int64, cost_gby_now int64) {
	nyear = int64(math.Max((float64(h+BLOCKS_PER_YEAR-DAY_2020_LEFT) / float64(BLOCKS_PER_YEAR)), 0))
	cost_gby_now = int64(float64(COST_GBY_2020) * math.Pow(float64(DECAY_ANNUAL_GBY), float64(nyear)))
	pe_cost = int64(float64(cost_gby_now) / (1 - DECAY_ANNUAL_GBY))
	// fmt.Printf("h=%d, y=%d, g=%.10f, c=%d\n", h, nyear, cost_gby_now, pe_cost)
	return
}
func TxPayout(h int64, dataBytes int64, newTarget bool) (endowFee int64, minerFee int64) {
	txBytes := TX_SIZE_BASE + dataBytes
	pecost, _, _ := PecostGb(h)
	pecost = pecost * txBytes / (1 << 30)
	endowFee = int64(float64(pecost) * TX_ENDOW_RATE)
	minerFee = int64(float64(pecost) * TX_REWARD_RATE)

	if newTarget {
		// endowFee += WALLET_GEN_FEE
		minerFee += WALLET_GEN_FEE
	}
	return
}

func Rinf(h int64) (rinf int64) {
	if h <= STAGE_1_H {
		rinf = STAGE_1_RINF
		return
	}
	GTotal := TOKEN_TOTAL - STAGE_1_H*(STAGE_1_RINF/MINTOKEN_PER_TOKEN) // in token
	rinf_token := float64(GTotal) * math.Log(2) * math.Pow(2, -float64((h-STAGE_1_H)/BLOCKS_PER_4YEAR)) / float64(BLOCKS_PER_4YEAR)
	rinf = int64(rinf_token * float64(MINTOKEN_PER_TOKEN))
	return
}
func PledgeBase(h int64) (pledge_base int64) {
	if h <= STAGE_1_H {
		pledge_base = 0
	} else if h <= STAGE_2_H {
		pledge_base = int64(float64(STAGE_2_PLEDGE_BASE) * math.Log(float64(h)) / (math.Log(float64(h)) + float64(1)))
	} else {
		pledge_base = int64(float64(STAGE_3_PLEDGE_BASE) * math.Log(float64(h)) / (math.Log(float64(h)) + float64(1)))
	}
	return
}
func LockBlocks(h int64) (lock_blocks int64) {
	if h <= STAGE_1_H {
		lock_blocks = 0
	} else if h <= STAGE_2_H {
		// if h >= STAGE_1_H+STAGE_2_LOCKB { //max lock
		lock_blocks = STAGE_2_LOCKB
		// } else {
		// 	lock_blocks = h - STAGE_1_H
		// }
	} else {
		// if h >= STAGE_1_H+STAGE_3_LOCKB {
		lock_blocks = STAGE_3_LOCKB
		// } else {
		// 	lock_blocks = h - STAGE_1_H
		// }
	}
	return
}

// func CostGby(timesteamp) float
// func PeCostGby(timesteamp) { CostGby * (1-0.995)}

// const COST_PER_BYTE =
