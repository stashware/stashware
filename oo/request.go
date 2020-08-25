package oo

import (
	// "encoding/base64"
	// "errors"
	"encoding/json"
	// "fmt"
	// "regexp"
	// "sort"
	// "strconv"
	"strings"
	// "sync"
	// "time"
)

const RpcTypeReq = "req"
const RpcTypeResp = "rsp"
const RpcTypePush = "psh"
const RpcTypeErr = "err"

type ReqCtx = struct {
	Ctx    interface{} //ws, ns
	Cmd    string
	Sess   *Session
	CoSche bool
}

type Session = struct {
	Uid    uint64 `json:"uid,omitempty"`
	Key    string `json:"key,omitempty"`
	Gwid   string `json:"gwid,omitempty"`
	Connid uint64 `json:"connid,omitempty"`
	Ip     string `json:"ip,omitempty"` //for ipv4/ipv6
}

type RpcMsg = struct {
	Cmd  string          `json:"cmd,omitempty"`
	Typ  string          `json:"typ,omitempty"`
	Chk  string          `json:"chk,omitempty"`
	Sess *Session        `json:"sess,omitempty"`
	Err  *Error          `json:"err,omitempty"`
	Para json.RawMessage `json:"para,omitempty"`
}

type ParaNull = struct {
}

type ReqCtxHandler = func(*ReqCtx, *RpcMsg) (*RpcMsg, error)

func MakeReqCtx(c interface{}, cmd string, sess *Session) (ctx *ReqCtx) {
	ctx = &ReqCtx{
		Ctx:  c,
		Cmd:  cmd,
		Sess: sess,
	}
	return
}

func PackRequest(cmd string, para interface{}, sess *Session) *RpcMsg {
	return &RpcMsg{Cmd: cmd, Typ: RpcTypeReq, Para: JsonData(para), Sess: sess}
}

func PackReturn(reqmsg *RpcMsg, eno string, rsp interface{}) (rspmsg *RpcMsg, err error) {
	rspmsg = reqmsg
	rspmsg.Typ = RpcTypeResp
	if eno == ESUCC {
		rspmsg.Para = JsonData(rsp)
	} else {
		rspmsg.Para = nil
		rspmsg.Err = NewErrno(eno)
		// err = rspmsg.Err
	}
	return
}

func MakeSubjs(subjs []string) (m map[string]bool) {
	m = make(map[string]bool)

	for _, cmd := range subjs {
		cmd = strings.Split(cmd, ":")[0]
		if Str2Bytes(cmd)[0] == '-' {
			m[cmd[1:]] = false
		} else {
			m[cmd] = true
		}
	}
	return
}
func CheckSubj(m map[string]bool, subj string) bool {
	if len(m) == 0 {
		return false
	}
	if c, ok := m[subj]; ok {
		return c
	}
	subjs := strings.Split(subj, ".")
	for i := len(subjs) - 1; i > 0; i-- {
		subjs[i] = "*"
		subj = strings.Join(subjs, ".")
		if c, ok := m[subj]; ok {
			return c
		}
	}
	return false
}
