package chain

import (
	"math"
	"math/big"

	"stashware/types"
)

var (
	target_time = int64(types.RETARGET_BLOCKS * types.TARGET_SEC)
	diff_max    = new(big.Int).Lsh(big.NewInt(1), 256)
	diff_min    = LinerDiff(31 - 14)
)

func CalcNextRequiredDifficulty399(block *types.Block) (ret *big.Int) {
	ret = block.BigDiff()

	if 0 == (block.Height+1)%types.RETARGET_BLOCKS {
		var (
			actual_time = block.Timestamp - block.LastRetarget
		)
		delta := float64(actual_time) / float64(target_time)

		if math.Abs(1-delta) < types.RETARGET_TOLERANCE {
			return
		}

		use_delta := delta
		if use_delta < types.DIFF_ADJUSTMENT_DOWN_LIMIT {
			use_delta = types.DIFF_ADJUSTMENT_DOWN_LIMIT

		} else if use_delta > types.DIFF_ADJUSTMENT_UP_LIMIT {
			use_delta = types.DIFF_ADJUSTMENT_UP_LIMIT
		}

		sub_i := new(big.Int).Sub(diff_max, ret)
		sub_f := new(big.Float).SetInt(sub_i)
		inv_f := new(big.Float).Mul(sub_f, new(big.Float).SetFloat64(use_delta))
		inv_i, _ := inv_f.Int(nil)

		ret_i := new(big.Int).Sub(diff_max, inv_i)

		if ret_i.Cmp(diff_min) < 0 {
			ret_i = diff_min

		} else if ret_i.Cmp(diff_max) > 0 {
			ret_i = diff_max
		}

		ret = ret_i
	}

	return
}

func CalcNextRequiredDifficulty(block *types.Block) (ret *big.Int) {
	ret = block.BigDiff()

	abs := func(num int64) int64 {
		if num >= 0 {
			return num
		}
		return -num
	}

	if 0 == (block.Height+1)%types.RETARGET_BLOCKS {
		var (
			actual_time = block.Timestamp - block.LastRetarget
		)
		delta := actual_time * 1e4 / target_time

		if abs(1e4-delta) < types.RETARGET_TOLERANCE_INT {
			return
		}

		use_delta := delta
		if use_delta < types.DIFF_ADJUSTMENT_DOWN_LIMIT_INT {
			use_delta = types.DIFF_ADJUSTMENT_DOWN_LIMIT_INT

		} else if use_delta > types.DIFF_ADJUSTMENT_UP_LIMIT_INT {
			use_delta = types.DIFF_ADJUSTMENT_UP_LIMIT_INT
		}

		sub_i := new(big.Int).Sub(diff_max, ret)
		mul_i := new(big.Int).Mul(sub_i, new(big.Int).SetInt64(use_delta))
		inv_i := new(big.Int).Div(mul_i, new(big.Int).SetInt64(1e4))

		ret_i := new(big.Int).Sub(diff_max, inv_i)

		if ret_i.Cmp(diff_min) < 0 {
			ret_i = diff_min

		} else if ret_i.Cmp(diff_max) > 0 {
			ret_i = diff_max
		}

		ret = ret_i
	}

	return
}

func CalcNextCumulativeDiff(last_c_diff int64, diff *big.Int) (c_diff int64) {
	c_diff = last_c_diff + InverseLinerDiff(diff)

	return
}

// impl switch_to_linear_diff
func LinerDiff(diff int64) *big.Int {
	return new(big.Int).Sub(diff_max, new(big.Int).Lsh(big.NewInt(1), uint(256-diff)))
}

func InverseLinerDiff(num *big.Int) int64 {
	for i, val := range level_diff {
		if num.Cmp(val) <= 0 {
			return int64(i - 1)
		}
	}

	return 256
}

var (
	level_diff = make([]*big.Int, 256)
)

func init() {
	for i := 0; i < 256; i++ {
		level_diff[i] = LinerDiff(int64(i))
	}
}
