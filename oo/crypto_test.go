package oo

import (
	"testing"
)

func TestBase58(t *testing.T) {
	t.Logf("==== %s\n", Base58EncodeString(nil))

	var buf []byte
	t.Logf("==== %s\n", Base58EncodeString(buf))

	t.Logf("==== %x\n", Base58DecodeString(""))
	t.Logf("==== %x\n", Base58DecodeString("F2MzfHX7L14m9JeuoBhGiESkr2fq7NHR7Wz2V6S8B3x5"))
}

func TestIntToBytes(t *testing.T) {
	var num int64 = 1e10
	t.Logf("%d %x\n", num, IntToBytes(num))
}

func TestStr2Bytes(t *testing.T) {
	str := "041225a6bc8308cffb24b4fd08626b97fd1dbe0589a93b9f39f47d8a97d57bef"

	// t.Logf("%0x\n", Str2Bytes(str)) // unavailable
	buf1 := HexDecodeString(str)
	t.Logf("[%d] : %0x\n", len(buf1), buf1)

	buf2 := JsonData(&str)
	t.Logf("JsonData [%d]: %0x\n", len(buf2), buf2)

	msg1 := &RpcMsg{
		Para: buf1,
	}
	t.Logf("==== %s\n", JsonData(msg1))

	msg2 := &RpcMsg{
		Para: buf2,
	}
	t.Logf("==== %s\n", JsonData(msg2))
}

func TestEmptyHash(t *testing.T) {
	// buf := Sha256("")
	// t.Logf("%064x\n", buf)
	// t.Logf("%s\n", Base58EncodeString(buf))
}

func TestHex(t *testing.T) {
	buf := HexDecStringPad32("")
	t.Logf("==== %x", buf)

	str := HexEncToStringPad32(nil)
	t.Logf("==== %s", str)
}
