package oo

import (
	"testing"
)

func TestRandomX(t *testing.T) {
	var (
		key = []byte("test key 000")
		buf = []byte("This is a test")
		// 46b49051978dcce1cd534a4066035184afb16a0591b43522466e10cc2496717e
	)
	ret, err := RandomX(key, buf)
	if nil != err {
		t.Fatal(err)
	}

	t.Logf("%x\n", ret)
}
