package node

import (
	"net/url"
	"stashware/chain"
	"stashware/oo"
	"stashware/types"
	"strings"
	"time"
)

func NodeStartPeers() {
	conf := &GNodeServer.conf

	n, err := TryConnectPeers(conf.InitPeers, true)
	if err != nil {
		oo.LogD("Failtd to connect trusted peers, err %v", err)
	} else {
		oo.LogD("succeed connect %d peers", n)
	}

	for {
		var all_peers []string
		nodes, _ := PeerTopN(types.MAX_PEERS)
		if len(nodes) == types.MAX_PEERS {
			goto _Next
		}
		nodes = SelectTrustedNodes()

		all_peers = []string{}
		for _, pnode := range nodes {
			if pnode.Ws == nil {
				continue
			}
			peers, err := RpcPeers(pnode.Ws)
			if err != nil {
				oo.LogD("failed to get peers. %v err %v", pnode.Ip, err)
				continue
			}
			peers = AdjuestPeerAddrs(peers)

			all_peers = append(all_peers, peers...)
		}
		//
		all_peers = oo.StringsUniq(all_peers, nil)
		//
		all_peers = RemoveRejectPeers(all_peers)
		//
		// all_peers = SelectSubset(all_peers)

		oo.LogD("try connect peers %v", all_peers)
		if len(all_peers) > 0 {
			n, err = TryConnectPeers(all_peers, false)
			if err != nil {
				oo.LogD("Failtd to connect more peers, err %v", err)
			} else {
				oo.LogD("try peers %d", n)
			}
		}

	_Next:
		time.Sleep(time.Duration(conf.PeerSec) * time.Second)
	}
}

func IsAddrReject(peerAddr string) (reject bool) {
	if !strings.HasPrefix(peerAddr, "ws://") {
		peerAddr = "ws://" + peerAddr
	}
	vurl, err := url.Parse(peerAddr)
	if err != nil {
		oo.LogD("Failed to parse peer %v", peerAddr)
		return
	}

	if oo.InArray(vurl.Host, GNodeServer.conf.RejectIps) {
		reject = true
	}
	return
}
func IsAddrTrusted(peerAddr string) (trusted bool) {
	if !strings.HasPrefix(peerAddr, "ws://") {
		peerAddr = "ws://" + peerAddr
	}
	// vurl, err := url.Parse(peerAddr)
	// if err != nil {
	// 	oo.LogD("Failed to parse peer %v", peerAddr)
	// 	return
	// }
	if oo.InArray(peerAddr, GNodeServer.conf.InitPeers) {
		trusted = true
	}
	return
}

func RejectAddr(peerAddr string) {
	if !strings.HasPrefix(peerAddr, "ws://") {
		peerAddr = "ws://" + peerAddr
	}
	vurl, err := url.Parse(peerAddr)
	if err != nil {
		oo.LogD("Failed to parse peer %v", peerAddr)
		return
	}

	if !oo.InArray(vurl.Host, GNodeServer.conf.RejectIps) {
		GNodeServer.conf.RejectIps = append(GNodeServer.conf.RejectIps, vurl.Host)
	}
}

func RemoveRejectPeers(peers []string) (new_peers []string) {

	for _, one := range peers {
		if !IsAddrReject(one) {
			new_peers = append(new_peers, one)
		}
	}

	return
}

func TryConnectPeers(peers []string, trusted bool) (n_try int, err error) {
	//
	peers = SelectSubset(peers)

	//try reconnect 3 times
	max_reconnect := int64(0)
	if trusted {
		max_reconnect = int64(-1)
	}
	//
	for _, one := range peers {
		if err = ConnectPeer(one, max_reconnect); err != nil {
			oo.LogD("err %v", err)
		}
	}
	n_try = len(peers)
	return
}

func ConnectPeer(peerAddr string, max_reconnect int64) (err error) {

	var vurl *url.URL
	var ws *oo.WebSock

	if vurl, err = url.Parse(peerAddr); err != nil {
		oo.LogD("Failed to parse peer %v", peerAddr)
		return
	}
	if IsPeerConnected(vurl.Host) {
		err = oo.NewError("cancel double connect. %v", vurl.Host)
		return
	}

	if ws, err = oo.InitWsClient(vurl.Scheme, vurl.Host, vurl.Path, GNodeServer.chmgr, oo.RecvRpcFn); err != nil {
		oo.LogD("Failed to connect peer %v", peerAddr)
		return
	}
	oo.LogD("prepare connect %v", peerAddr)
	ws.SetOpenHandler(LinkOpen)
	ws.SetIntervalHandler(AutoPing)

	//connecting : set_state --> 1. connected: set ws, 2. disconnect : clear ws
	// ws.Data = NodeEvent(types.NSEVENT_CONNECTING, ws)
	//registe
	go ws.StartDial(60, max_reconnect)
	return
}

func LinkOpen(ws *oo.WebSock) (err error) {
	peerAddr := ws.PeerAddr()
	if IsAddrReject(peerAddr) {
		oo.LogD("reject this peer. %s", peerAddr)
		err = oo.NewError("reject this peer. %s", peerAddr)
		return
	}

	ws.OpenCoroutineFlag()
	ws.SetCloseHandler(func(c *oo.WebSock) {
		if debug_flag > 0 {
			oo.LogD("Peer close : %v.", c.ConnInfo())
		}

		PeerDisConnected(c)
	})
	// ws.SetIntervalHandler(AutoPing)
	ws.HandleFuncMap(GNodeHandlers)
	// ws.HandleFunc(types.CMD_INFO, OnReqInfo)
	if debug_flag > 0 {
		oo.LogD("Peer open: %v", ws.ConnInfo())
	}
	return
}

func AutoPing(ws *oo.WebSock, ms int64) {
	if !ws.IsReady() {
		return
	}

	ts, ok := ws.Data.(int64)
	if !ok {
		ws.Data = int64(0)
		ts = 0
	}

	if ts+GNodeServer.conf.PingSec < ms/1e3 {
		ws.Data = ms / 1e3

		if IsAddrReject(ws.PeerAddr()) {
			return
		}

		info, err := buildInfo()
		if nil != err {
			oo.LogD("AutoPing err %v", err)
			return
		}

		ping_ret, err := RpcInfo(ws, info)
		if err != nil {
			oo.LogD("request %s info err %v", ws.PeerAddr(), err)
			// ws.Close()
			return
		}

		if ping_ret.Uuid == Uuid ||
			ping_ret.Version < types.RequiredMinVersion ||
			ping_ret.Genesis != chain.GenesisIndepHash {

			oo.LogD("reject %s, uuid %s, ver %s, genesis %s", ws.PeerAddr(),
				ping_ret.Uuid, ping_ret.Version, ping_ret.Genesis)
			RejectAddr(ws.PeerAddr())
			return
		}

		PeerConnected(ws, ping_ret)
	}
}
