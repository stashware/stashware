package node

import (
	"fmt"
	"stashware/oo"
	"stashware/types"
	"testing"
)

func TestTopN(t *testing.T) {

	GNodeServer.conf.InitPeers = []string{"ws://192.168.1.3:2080", "ws://192.168.1.9:2080"}
	for i := 0; i < 10; i++ {
		ip := fmt.Sprintf("ws://192.168.1.%d:2080", i)
		n := &Node{
			Ip:    ip,
			Score: int64(i * 100),
			Ws:    &oo.WebSock{},
			Info:  types.CmdInfoRsp{},
		}
		Ip2Score.Store(ip, n)
	}

	str := PrintScore()
	t.Logf("==%s==", str)
}
