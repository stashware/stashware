package node

import (
	// "github.com/sakeven/RbTree"
	"fmt"
	"net/url"
	"sort"
	"stashware/oo"
	"stashware/types"
	"strings"
	"sync"
)

type Node struct {
	Ip        string
	Score     int64
	PingTs    int64
	PeerTs    int64
	Ws        *oo.WebSock
	Info      types.CmdInfoRsp
	LastEvent int
}

var AllNodes []*Node
var nodesLock sync.Mutex

var Ws2Node sync.Map  //map[*oo.WebSock]*Score
var Ip2Conns sync.Map //map[string]*int

func nodesSort() {
	sort.Slice(AllNodes, func(i, j int) bool {
		iscore := AllNodes[i].Score
		if IsAddrTrusted(AllNodes[i].Ip) {
			iscore += types.NSEVENT_TRUSTED
		}

		jscore := AllNodes[j].Score
		if IsAddrTrusted(AllNodes[j].Ip) {
			jscore += types.NSEVENT_TRUSTED
		}
		return iscore > jscore //reverse
	})
}
func PeerConnected(ws *oo.WebSock, pinfo *types.CmdInfoRsp) {

	ip := ws.PeerAddr()
	ip = AdjustPeerAddr(ip)
	newps := &Node{
		Score: 0,
		Ws:    ws,
		Ip:    ip,
		Info:  *pinfo,
	}

	nodesLock.Lock()
	defer nodesLock.Unlock()

	if wv, ok := Ws2Node.Load(ws); ok {
		ps, _ := wv.(*Node)
		// ps.Uuid = newps.Uuid
		// ps.Score = 0
		ps.Ws = newps.Ws
		ps.Ip = newps.Ip

		ps.Info = newps.Info
	} else {
		AllNodes = append(AllNodes, newps)
		nodesSort()
		Ws2Node.Store(ws, newps)
	}

	conns := 0
	if pv, loaded := Ip2Conns.LoadOrStore(newps.Ip, &conns); loaded {
		pconn, _ := pv.(*int)
		*pconn = *pconn + 1
	}
}

func PeerDisConnected(ws *oo.WebSock) {
	ip := ws.PeerAddr()
	ip = AdjustPeerAddr(ip)

	nodesLock.Lock()
	defer nodesLock.Unlock()

	if wv, ok := Ws2Node.Load(ws); ok {
		ps, _ := wv.(*Node)
		ps.Ws = nil

		Ws2Node.Delete(ws)

		if pv, loaded := Ip2Conns.Load(ip); loaded {
			pconn, _ := pv.(*int)
			*pconn = *pconn - 1
		}
	}
}

func PeerGetUuid(ws *oo.WebSock) (uuid string) {
	if ws == nil {
		return
	}

	if wv, ok := Ws2Node.Load(ws); ok {
		ps, _ := wv.(*Node)
		uuid = ps.Info.Uuid
	}
	return
}

func IsPeerConnected(ip string) (connected bool) {
	ip = AdjustPeerAddr(ip)
	if pv, loaded := Ip2Conns.Load(ip); loaded {
		pint, ok := pv.(*int)
		if ok && *pint > 0 {
			connected = true
		}
	}
	return
}

func PeerEvent(ws *oo.WebSock, ev int, pinfo *types.CmdInfoRsp) {

	nodesLock.Lock()
	defer nodesLock.Unlock()

	if wv, ok := Ws2Node.Load(ws); ok {
		ps, _ := wv.(*Node)
		if ev != 0 {
			ps.Score += int64(ev)
			nodesSort()
		}

		if pinfo != nil {
			ps.Info = *pinfo
		}
	}
}

func PeerEventCtx(ctx *oo.ReqCtx, ev int, pinfo *types.CmdInfoRsp) {
	if nil == ctx {
		return
	}
	ws, ok := ctx.Ctx.(*oo.WebSock)
	if !ok || ws.PeerAddr() == "" {
		return
	}

	PeerEvent(ws, ev, pinfo)
	return
}

func PeerTopN(n int) (peers []Node, wss []*oo.WebSock) {
	nodesLock.Lock()
	defer nodesLock.Unlock()

	for i := 0; i < len(AllNodes); i++ {
		ps := AllNodes[i]
		if ps.Ws == nil {
			continue
		}
		if len(peers) >= n {
			break
		}
		peers = append(peers, *ps)
		wss = append(wss, ps.Ws)
	}
	return
}

//

func SelectSyncNodes(info *types.CmdInfoRsp) (nodes []Node) {
	defer func() {
		if errs := recover(); errs != nil {
			oo.LogW("recover: %v", errs)
		}
	}() //val.Info == nil

	nodes, _ = PeerTopN(types.MAX_TRUSTED_NODES)

	// Ip2Score.Range(func(k, v interface{}) bool {
	// 	ns, _ := v.(*Node)
	// 	if ns.Ws != nil && ns.Info != nil && info.Height < ns.Info.Height {
	// 		nodes = append(nodes, ns)
	// 	}
	// 	return true
	// })
	if len(nodes) <= 0 {
		return
	}

	todels := map[string]bool{}
	n := 0
	for i, val := range nodes {
		if val.Info.Height <= info.Height {
			continue
		}

		hash := val.Info.Current
		if _, ok := todels[hash]; !ok {
			todels[hash] = true
			if n != i {
				nodes[n] = nodes[i]
			}
			n++
		}
	}
	if n < len(nodes) {
		nodes = nodes[:n]
	}

	//dont sort, keep random
	// sort.Slice(nodes, func(i, j int) bool {
	// 	return nodes[i].Info.Height > nodes[j].Info.Height //reverse
	// })
	return
}

func SelectTrustedNodes() (nodes []Node) {
	nodes, _ = PeerTopN(types.MAX_TRUSTED_NODES)
	if len(nodes) <= types.MIN_TRUSTED_NODES {
		return
	}
	n := len(nodes) - 1
	for i := n; i > types.MIN_TRUSTED_NODES; i-- {
		if nodes[i].Score < 0 {
			n--
		}
	}
	nodes = nodes[:n+1]

	return
}

func AdjustPeerAddr(peerAddr string) (newAddr string) {
	if !strings.HasPrefix(peerAddr, "ws://") {
		peerAddr = "ws://" + peerAddr
	}
	vurl, err := url.Parse(peerAddr)
	if err != nil {
		oo.LogD("Failed to parse peer %v", peerAddr)
		return
	}
	// newAddr = fmt.Sprintf("ws://%s:%d", vurl.Hostname(), types.DEFAULT_SYNC_PORT)
	newAddr = fmt.Sprintf("ws://%s:%d", vurl.Hostname(), GNodeServer.ListenPort)
	return
}
func AdjuestPeerAddrs(peers []string) (newPeers []string) {
	emap := map[string]struct{}{}
	for _, peer := range peers {
		addr := AdjustPeerAddr(peer)
		if _, ok := emap[addr]; !ok {
			newPeers = append(newPeers, addr)
			emap[addr] = struct{}{}
		}
	}
	return
}

func AllPeers() (ret []string) {

	nodes, _ := PeerTopN(types.MAX_PEERS)
	for i := 0; i < len(nodes); i++ {
		ret = append(ret, nodes[i].Ip)
	}

	return
}

func SelectSubset(peers []string) (newPeers []string) {

	for _, val := range peers {
		if ok := IsPeerConnected(val); !ok {
			newPeers = append(newPeers, val)
		}
	}

	return
}

func PrintScore() (strs []string) {
	nodes, _ := PeerTopN(100)
	for _, n := range nodes {
		strs = append(strs, fmt.Sprintf("%v, tr %v", n, IsAddrTrusted(n.Ip)))
	}
	return
}

var Id2SourceId sync.Map //map[id]uuid;  1

func ResMarkPeer(resid string, uuid string) {
	if len(uuid) > 0 {
		oo.LogD("ResMarkPeer %s ==> %s", resid, uuid)
		Id2SourceId.Store(resid, uuid)
	}
}

func ResFromPeer(txid string) (uuid string) {
	if uuid1, ok := Id2SourceId.Load(txid); ok {
		uuid = uuid1.(string)
	}
	return
}
