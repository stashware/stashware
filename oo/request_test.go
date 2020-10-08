package oo

import (
	"errors"
	"testing"
)

func TestPackReturn(t *testing.T) {
	msg := RpcMsg{
		Cmd: "tx_by_id",
		Typ: "rsp",
		Chk: ".496",
		Err: &Error{
			Eno: "ESERVER",
		},
	}
	js, err := JsonMarshal(msg)
	t.Logf("js: %s, err: %v\n", js, err)
	//.       {"cmd":"tx_by_id","typ":"rsp","chk":".496","err":{"eno":"ESERVER","err":{}}}

	t.Logf("%#v", errors.New("ssssss").Error())
	// js2 := `{"cmd":"tx_by_id","typ":"rsp","chk":".496","err":{"eno":"ESERVER"}}`
	msg2 := RpcMsg{}
	t.Logf("msg2: %v, err %v\n", msg2, err)
	err = JsonUnmarshal([]byte(js), &msg2)
	t.Logf("msg2: %v, err %v\n", msg2, err)
}
