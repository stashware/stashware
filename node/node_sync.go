package node

import (
	"fmt"
	"time"

	"stashware/chain"
	"stashware/oo"
	"stashware/storage"
	"stashware/types"
)

var SyncingHeight int64

func NodeStartSync() {
	var info *types.CmdInfoRsp
	// var nodes []Node
	var ninfo []string

	for {
		tipblock := *(chain.TipBlock())
		info = &types.CmdInfoRsp{
			Network: "main",
			Height:  tipblock.Height,
			Current: tipblock.IndepHash,
		}
		nodes := SelectSyncNodes(info)
		if len(nodes) == 0 {
			goto _Next
		}
		ninfo = []string{}
		for _, node := range nodes {
			peer := fmt.Sprintf("%v %s", node.Info, node.Ws.PeerAddr())
			ninfo = append(ninfo, peer)
		}
		if debug_flag > 0 {
			oo.LogD("self %v, get nodes: %v", *info, ninfo)
		}
		TryNodeSync(&nodes[0], &tipblock)
	_Next:
		nss, _ := PeerTopN(types.MIN_PEERS)
		if !chain.GChain.IsReady() && len(nss) >= 1 {
			oo.LogD("open local mining.")
			chain.GChain.SetReady(true)
		} else if chain.GChain.IsReady() && len(nss) < types.MIN_PEERS {
			//
		}
		time.Sleep(1 * time.Second)
	}
}

func TryFindForkPoint(node *Node, tipb *types.Block) (targetb *types.Block, err error) {
	ws := node.Ws
	if ws == nil {
		return
	}

	//min fork height allow.
	// h1 := tipb.Height - types.RETARGET_BLOCKS
	h1 := tipb.Height - types.BLOCKS_PER_DAY
	if h1 < 0 {
		h1 = 0
	}

	//get max height allow.
	h2 := node.Info.Height
	if h2 < tipb.Height {
		//h2 = tipb.Height; //not allow that.
		err = oo.NewError("not allow new chain %d < now chain %d", h2, tipb.Height)
		return
	}

	h2 = tipb.Height //from here

	for h1 <= h2 {
		h := (h1 + h2) / 2
		if debug_flag != int64(0) {
			oo.LogD("trying fork point %d~%d = %d", h1, h2, h)
		}

		var localb *types.TbBlock
		if localb, err = storage.GStore.ReadBlockByHeight(h); err != nil {
			//logic failed
			return
		}

		var remoteb *types.Block
		if remoteb, err = RpcReadBlockHeight(ws, h); err != nil {
			//peer error
			return
		}
		if remoteb.IndepHash == localb.IndepHash {
			if debug_flag != int64(0) {
				oo.LogD("fork point before: %d(%s)", remoteb.Height, remoteb.IndepHash)
			}
			targetb = remoteb
			h1 = h + 1
		} else {
			h2 = h - 1
		}
	}
	if targetb == nil {
		err = oo.NewError("not found fork point.")
	}

	return
}

func TryNodeSync(node *Node, tipb *types.Block) {
	ws := node.Ws
	if ws == nil {
		return
	}

	//first : find the same block hash
	targetb, err := TryFindForkPoint(node, tipb)
	if err != nil {
		PeerEvent(ws, types.NSEVENT_INTERNAL_ERROR, nil)
		oo.LogD("Failed to find fork point. peer %s, err %v", ws.PeerAddr(), err)
		return
	}

	if tipb.IndepHash != chain.TipBlock().IndepHash {
		oo.LogD("change tip block. %d %s --> %d %s skip node sync.",
			tipb.Height, tipb.IndepHash, chain.TipBlock().Height, chain.TipBlock().IndepHash)
		return
	}

	chain.GChain.SetReady(false)
	defer func() {
		chain.GChain.SetReady(true)
	}()

	if targetb.IndepHash != tipb.IndepHash {
		//need fork
		err = GNodeServer.Chain.ForkChain(targetb)
		if err != nil {
			oo.LogW("Failed to fork chain. err %v", err)
			return
		}
		oo.LogD("to do fork: %d %s ==> %d", targetb.Height, targetb.IndepHash, node.Info.Height)
	}
	if node.Ws != ws {
		return
	}

	starth := targetb.Height + 1
	stoph := node.Info.Height
	if starth > stoph {
		return
	}

	SyncingHeight = stoph //global
	if starth+types.RETARGET_BLOCKS > stoph {
		stoph = starth + types.RETARGET_BLOCKS
	}

	// bids := []string{}
	oo.LogD("syncing %s: %d--->%d", ws.PeerAddr(), starth, stoph)

	pre := targetb.IndepHash
	for height := starth; height <= stoph; height++ {
		b, err := RpcReadBlockHeight(ws, height)
		if err != nil {
			oo.LogD("Failed to read block. %d - %v. %s", height, err, ws.PeerAddr())
			PeerEvent(ws, types.NSEVENT_REQUEST_TIMEOUT, nil)
			return
		}
		//check checkpoint
		if _, failed := types.CheckCps(b.Height, b.IndepHash); failed {
			PeerEvent(ws, types.NSEVENT_BLOCK_ERROR, nil)
			return
		}
		if b.PreviousBlock != pre {
			oo.LogD("NOTFOUND preblock, peer change chain. %d %s", b.Height-1, b.PreviousBlock)
			return
		}
		oo.LogD("from %s down %d %s", ws.PeerAddr(), b.Height, b.IndepHash)
		pre = b.IndepHash
		// if _, err = storage.GStore.ReadBlockByHash(b.PreviousBlock); err != nil {
		// 	oo.LogD("NOTFOUND preblock, peer change chain. %h %s,", b.Height-1, b.PreviousBlock)
		// 	return
		// }

		if err = GNodeServer.ReadFullBlockFromPeer(ws, b); err != nil {
			oo.LogW("Failed to read full block from peer %s. err %v", ws.PeerAddr(), err)
			return
		}

		poa, err := GNodeServer.Chain.GetPoa(oo.HexDecStringPad32(b.PreviousBlock), height)
		if err != nil {
			oo.LogW("Failed to calc poa, h=%d, prehash=%s. err %v", height, b.PreviousBlock, err)
			return
		}
		if len(poa) > 0 {
			b.Poa = oo.HexEncodeToString(poa)
		}
		BDS, err := GNodeServer.Chain.GetBDS(b)
		if err != nil {
			oo.LogW("Failed to calc bds,  h=%d, prehash=%s. err %v", height, b.PreviousBlock, err)
			return
		}

		if err := GNodeServer.Chain.OnNewBlock(b, BDS, false); err != nil {
			oo.LogD("cancel sync. h=%d, peer=%v", height, ws.PeerAddr())
			if err == types.BLK_VERIFY_ERROR {
				PeerEvent(ws, types.NSEVENT_BLOCK_ERROR, nil)
			}
			return
		}
		// bids = append(bids, b.IndepHash)
	}

	// if len(bids) > 0 {
	// 	GNodeServer.Chain.SwitchChain(oldc, newc, new_ids)
	// }
}

func (this *NodeServer) ReadFullBlockFromPeer(ws *oo.WebSock, b *types.Block) (err error) {
	// if b.WalletList, err = RpcGetWalletList(ws, b.IndepHash); err != nil {
	// 	oo.LogW("failed to read wallet_list. %d- %v", b.Height, err)
	// 	return
	// }

	var (
		txs []*types.Transaction
		set = make(map[string]bool) // txid -> write_to_db
	)

	for _, txid := range b.Txs {
		var (
			full_tx *types.FullTx
			tx      *types.Transaction
		)
		full_tx, err = chain.GetFullTx(txid)
		if nil != err {
			if storage.ErrNoRows != err {
				err = oo.NewError("GetFullTx(%s) err %v", txid, err)
				return
			}
		} else {
			txs = append(txs, full_tx.Transaction())
			continue
		}

		tx, err = RpcReadTxAndData(ws, txid)
		if nil != err {
			err = oo.NewError("RpcReadTxAndData(%s) err %v", txid, err)
			return
		}
		txs = append(txs, tx)

		set[txid] = true
	}

	if false == chain.VerifyBlockTxs(b.Height, txs) {
		err = oo.NewError("VerifyBlockTxs failed")
		return
	}

	var tbtxs []types.TbTransaction
	for _, tx := range txs {
		if set[tx.ID] {
			if "" != tx.DataHash {
				err = storage.GStore.WriteTxData(tx.ID, oo.HexDecodeString(tx.Data))
				if nil != err {
					err = oo.NewError("WriteTxData(%s) err %v", tx.ID, err)
					return
				}
			}
			tbtx := tx.TbTransaction()
			tbtxs = append(tbtxs, *tbtx)
		}
	}
	if len(tbtxs) > 0 {
		err = storage.GStore.WriteTxs(tbtxs)
		if nil != err {
			err = oo.NewError("WriteTxs err %v", err)
			return
		}
	}

	oo.LogD("read full block from peer %s", oo.JsonData(b))
	// if !VerifyBlock(b) {
	// 	err = oo.NewError("verify block")
	// }
	return
}

func (this *NodeServer) SendNewBlock(b *types.Block, BDS []byte) {
	var pushReq types.CmdSubmitBlockReq
	pushReq.NewBlock = *b
	pushReq.BDS = oo.HexEncodeToString(BDS)
	pushReq.HashList = []string{}
	reqmsg := oo.PackRequest(types.CMD_SUBMIT_BLOCK, pushReq, nil)
	// msgb := oo.JsonData(reqmsg)
	// n, err := this.chmgr.PushAllChannelMsg(msgb, getSendFilter(ctx))
	n, err := PushPeers(b.IndepHash, reqmsg)
	oo.LogD("send block %d to n=%d peer. err=%v", b.Height, n, err)
}

func (this *NodeServer) SendNewTx(tx *types.Transaction) {
	var pushReq types.CmdSubmitTxReq
	pushReq.Transaction = *tx
	reqmsg := oo.PackRequest(types.CMD_SUBMIT_TX, pushReq, nil)
	// msgb := oo.JsonData(reqmsg)
	// n, err := this.chmgr.PushAllChannelMsg(msgb, getSendFilter(ctx))
	n, err := PushPeers(tx.ID, reqmsg)
	oo.LogD("send tx %s to n=%d peer. err=%v", tx.ID, n, err)
}

func PushPeers(resid string, msg *oo.RpcMsg) (n int, err error) {
	uuid := ResFromPeer(resid)

	data, err := oo.JsonMarshal(msg)
	nodes, _ := PeerTopN(types.MAX_TRUSTED_NODES)
	for i := 0; i < len(nodes); i++ {
		pn := &nodes[i]
		if pn.Info.Uuid == uuid {
			oo.LogD("pushskip %s==>%v", resid, *pn)
			continue
		}
		err = pn.Ws.SendData(data)
		if err == nil {
			n++
		}
	}
	return
}
