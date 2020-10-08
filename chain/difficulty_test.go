package chain

import (
	"math"
	"math/big"
	"testing"

	"stashware/oo"
	"stashware/types"
)

func TestDifficulty(t *testing.T) {
	for i, val := range level_diff {
		t.Logf("%03d %064x", i, val.Bytes())

		if 0 != LinerDiff(int64(i)).Cmp(val) {
			t.Fatalf("%d %064x %064x", i, LinerDiff(int64(i)).Bytes(), val.Bytes())
		}
	}

	num, _ := new(big.Int).SetString("fffff80000000000000000000000000000000000000000000000000000000000", 16)
	if 21-1 != InverseLinerDiff(num) {
		t.Fatalf("%064x ==> %d", num.Bytes(), InverseLinerDiff(num))
	}
}

func TestDifficulty2(t *testing.T) {
	// str := "fffff80000000000000000000000000000000000000000000000000000000000"
	str := "fffffec6e011aa0a6b10b287fa9266c3dfc46e4d74ef1c9f9c1ca952dab2e1f4"

	t.Logf("diff = %s", str)

	num, ok := new(big.Int).SetString(str, 16)
	if !ok {
		t.Fatal(ok)
	}
	dint := InverseLinerDiff(num)
	t.Logf("dint = %d", dint)

	hash_rate := int64(math.Pow(float64(2), float64(dint)) / 150)
	t.Logf("hash_rate = %d", hash_rate)
}

func TestInverseLinerDiff(t *testing.T) {
	arr := []string{
		"JEKLtpK9e6KK13xbgi4jRAWC6cBLxXJXSCdq8aA1htPy",
		"JEKKHqshV8trS28n9ExAe7CrwiiNAqXxPMKLQDtFtpYf",
		"JEKMhJXsia2YHZsWTSd1oh9rg3uqMNBoxdJ5VkHtcRKd",
		"JEKMLbTUVGaYLypyDFRrLeYD8YwvBtSsvJhNXtAvrdGs",
		"JEKMhrZX6r1PHUmivnDE784QQHjNExrUJUFi9YtQvMNP",
		"JEKMt6Vs98J59L1eFuEXhXxZukyMb36yEHtw5orkf7LT",
		"JEKMjBDC4Afst74vXChmhFmhAv8348rbo4tLR2MgTBZR",
		"JEKMrSTnz5iPcU3EgkoPPxa7S8vYnL7pw9EedqExDh6H",
		"JEKMx8c4Ze9ZfMAe1f4ZtBsYd7cJBf13747xgoKZFQJh",
		"JEKMc2mrqahACXZL4Hm6VmLFzyDAEk1MpJXBjwUqFjWZ",
		"JEKMhgByUM959RwG7oQFXM8kZj96s594WUc8hBP6cWAp",
		"JEKMUpsur3o3KYwvaEY7i7m8ioYf8niBZDyZEcC6jhLq",
		"JEKMeptXaKZVbjRkh3eMHjPHzuhuaFwHUJvbU9ZMjXLT",
		"JEKMKQQjpc5qAWkNuENUGntqTMcENKvBFiZXB78Tg6Qy",
	}
	for _, str := range arr {
		num := new(big.Int).SetBytes(oo.Base58DecodeString(str))
		t.Logf("%s %03d", str, InverseLinerDiff(num))
	}
}

func TestCalcNextRequiredDifficulty(t *testing.T) {
	str := `{"indep_hash":"adc84ef3b5aed1d2a260438d580b3f6ee652e9f971a1103d363b22daa7c22cda","hash":"fffffdfc671f31b7176e65bf64c5d263edde4ae57885574d3918f29adf668741","height":399,"previous_block":"94b6717d3cbb9dc653719e3e394457dee18cabab4f75e28febb869bebeb6bc18","nonce":"019b47","timestamp":1598564967,"last_retarget":1598555499,"diff":"fffff7e070cfb07b265c5722de20a2a77de3c24120f000000000000000000000","cumulative_diff":7851,"reward_fee":13000000000000,"reward_addr":"3JQKrmpMymxLtWN6M4ZHBvzdFnBk8Gqf61","reward_pool":0,"weave_size":0,"block_size":0,"wallet_list":[{"wallet":"3JQKrmpMymxLtWN6M4ZHBvzdFnBk8Gqf61","quantity":494000000000000,"last_tx":""}],"wallet_list_hash":"5f9526ea84efc3ba4b8c5589b29d27ce8cd372015fbfb6ac7a27749be8330b88","txs":null,"tx_root":"","poa":""}`
	block := &types.Block{}

	err := oo.JsonUnmarshalValidate([]byte(str), block)
	if nil != err {
		t.Fatal(err)
	}

	a := oo.HexDecStringPad32("fffff5bec3a935371cd794abe86b3736d2bcd6fa907fd5655592895e00000000")
	b := CalcNextRequiredDifficulty(block).Bytes()

	t.Logf("%064x", a)
	t.Logf("%064x", b)
}
