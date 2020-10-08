package node

import (
	"stashware/chain"
	// "stashware/miner"
	"stashware/oo"
	"stashware/storage"
	"stashware/txpool"
	"stashware/types"
	"stashware/wallet"
)

var GNodeHandlers = map[string]oo.ReqCtxHandler{
	types.CMD_INFO:            OnReqInfo,
	types.CMD_PEERS:           OnReqPeers,
	types.CMD_DEBUG:           OnReqDebug,
	types.CMD_TX_PENDING:      OnReqTxPending,
	types.CMD_TX_BY_ID:        OnReqTxById,
	types.CMD_TX_DATA:         OnReqTxData,
	types.CMD_GET_ADDR_TXS:    OnReqGetAddrTxs,
	types.CMD_PRICE:           OnReqPrice,
	types.CMD_SUBMIT_TX:       OnReqSubmitTx,
	types.CMD_WALLET_NEW:      OnReqWalletNew,
	types.CMD_WALLET_BY_ADDR:  OnReqWalletByAddr,
	types.CMD_GET_BLOCK:       OnReqGetBlock,
	types.CMD_GET_WALLET_LIST: OnReqGetWalletList,
	types.CMD_SUBMIT_BLOCK:    OnReqSubmitBlock,
	// types.CMD_TX_FIELD:       OnReqTxField,
}

func buildInfo() (ret *types.CmdInfoRsp, err error) {
	block := chain.TipBlock()
	if nil == block {
		err = oo.NewError("buildInfo chain.TipBlock() == nil")
		return
	}

	hsyncing := SyncingHeight
	if hsyncing < block.Height {
		hsyncing = block.Height
	}
	ret = &types.CmdInfoRsp{
		Network:     "main",
		Genesis:     chain.GenesisIndepHash,
		Height:      block.Height,
		SyncHeight:  hsyncing,
		Blocks:      block.Height + 1,
		Current:     block.IndepHash,
		Version:     Version,
		Peers:       int64(len(AllPeers())),
		QueueLength: txpool.Count(),
		Uuid:        Uuid,
	}

	return
}

func OnReqInfo(ctx *oo.ReqCtx, reqmsg *oo.RpcMsg) (rspmsg *oo.RpcMsg, err error) {
	reqpara, rsppara := &types.CmdInfoReq{}, &types.CmdInfoRsp{}
	if err = oo.JsonUnmarshalValidate(reqmsg.Para, reqpara); err != nil {
		oo.LogD("OnReqInfo JsonUnmarshalValidate err %v", err)
		return oo.PackReturn(reqmsg, oo.EPARAM, nil)
	}

	info := reqpara
	if nil != ctx && "" != info.Uuid && //valid handshake request
		info.Uuid != Uuid && //not self
		info.Version >= types.RequiredMinVersion && //version require
		info.Genesis == chain.GenesisIndepHash { //genesis require

		if ws, ok := ctx.Ctx.(*oo.WebSock); ok {
			PeerConnected(ws, info)
		}
	}

	rsppara, err = buildInfo()
	if nil != err {
		oo.LogD("OnReqInfo err %v", err)
		return oo.PackReturn(reqmsg, oo.ESERVER, nil)
	}

	return oo.PackReturn(reqmsg, oo.ESUCC, rsppara)
}

func OnReqPeers(ctx *oo.ReqCtx, reqmsg *oo.RpcMsg) (rspmsg *oo.RpcMsg, err error) {
	var reqpara types.CmdPeersReq

	if err = oo.JsonUnmarshalValidate(reqmsg.Para, &reqpara); err != nil {
		oo.LogD("OnReqPeers JsonUnmarshalValidate err %v", err)
		return oo.PackReturn(reqmsg, oo.EPARAM, nil)
	}

	rsppara := &types.CmdPeersRsp{
		EndPoint: AllPeers(),
	}

	return oo.PackReturn(reqmsg, oo.ESUCC, rsppara)
}
func OnReqDebug(ctx *oo.ReqCtx, reqmsg *oo.RpcMsg) (rspmsg *oo.RpcMsg, err error) {
	var reqpara types.CmdDebugReq
	if debug_flag != 1 {
		return
	}
	if err = oo.JsonUnmarshalValidate(reqmsg.Para, &reqpara); err != nil {
		oo.LogD("OnReqPeers JsonUnmarshalValidate err %v", err)
		return oo.PackReturn(reqmsg, oo.EPARAM, nil)
	}

	rsppara := &types.CmdDebugRsp{
		Scores: PrintScore(),
	}

	return oo.PackReturn(reqmsg, oo.ESUCC, rsppara)
}
func OnReqTxPending(ctx *oo.ReqCtx, reqmsg *oo.RpcMsg) (rspmsg *oo.RpcMsg, err error) {
	// reqpara, rsppara := &types.CmdTxPendingReq{}, &types.CmdTxPendingRsp{}
	// if err = oo.JsonUnmarshalValidate(reqmsg.Para, reqpara); err != nil {
	// 	return oo.PackReturn(reqmsg, oo.EPARAM, nil)
	// }

	rsppara := &types.CmdTxPendingRsp{
		Txs: txpool.Ids(),
	}

	return oo.PackReturn(reqmsg, oo.ESUCC, &rsppara)
}

func OnReqTxById(ctx *oo.ReqCtx, reqmsg *oo.RpcMsg) (rspmsg *oo.RpcMsg, err error) {
	var reqpara types.CmdTxByIdReq

	if err = oo.JsonUnmarshalValidate(reqmsg.Para, &reqpara); err != nil {
		oo.LogD("OnReqTxById JsonUnmarshalValidate err %v", err)
		return oo.PackReturn(reqmsg, oo.EPARAM, nil)
	}

	if nil == storage.GStore {
		oo.LogD("OnReqTxById storer is nil.")
		return oo.PackReturn(reqmsg, oo.ESERVER, nil)
	}

	tx, err := storage.GStore.ReadTx(reqpara.Txid)
	if nil != err {
		oo.LogD("OnReqTxById ReadTx(%s) %v", reqpara.Txid, err)
		return oo.PackReturn(reqmsg, oo.ESERVER, nil)
	}
	rsppara := &types.CmdTxByIdRsp{
		Transaction: *(tx.Transaction()),
	}

	if "" != tx.BlockIndepHash {
		block, err := chain.GetBlock(tx.BlockIndepHash)
		if nil != err {
			oo.LogD("OnReqTxById chain.GetBlock(%s) err %v", tx.BlockIndepHash, err)
			return oo.PackReturn(reqmsg, oo.ESERVER, nil)
		}

		tip := chain.TipBlock()
		if nil == tip {
			oo.LogD("OnReqTxById err chain.GetBlock() == nil")
			return oo.PackReturn(reqmsg, oo.ESERVER, nil)
		}

		rsppara.Timestamp = block.Timestamp
		rsppara.BlockIndepHash = tx.BlockIndepHash
		rsppara.Confirmations = tip.Height - block.Height + 1
	}

	return oo.PackReturn(reqmsg, oo.ESUCC, rsppara)
}

func OnReqTxData(ctx *oo.ReqCtx, reqmsg *oo.RpcMsg) (rspmsg *oo.RpcMsg, err error) {
	reqpara := &types.CmdTxDataReq{}
	if err = oo.JsonUnmarshalValidate(reqmsg.Para, reqpara); err != nil {
		oo.LogD("OnReqTxData JsonUnmarshalValidate err %v", err)
		return oo.PackReturn(reqmsg, oo.EPARAM, nil)
	}

	buf, err := storage.GStore.ReadTxData(reqpara.Txid)
	if nil != err {
		oo.LogD("OnReqTxData storage.GStore.ReadTxData(%s) %v", reqpara.Txid, err)
		return oo.PackReturn(reqmsg, oo.ESERVER, nil)
	}

	rsppara := oo.HexEncodeToString(buf)

	return oo.PackReturn(reqmsg, oo.ESUCC, &rsppara)
}

func OnReqGetAddrTxs(ctx *oo.ReqCtx, reqmsg *oo.RpcMsg) (rspmsg *oo.RpcMsg, err error) {
	reqpara, rsppara := &types.CmdGetAddrTxsReq{}, &types.CmdGetAddrTxsRsp{}
	if err = oo.JsonUnmarshalValidate(reqmsg.Para, reqpara); err != nil {
		oo.LogD("OnReqGetAddrTxs JsonUnmarshalValidate err %v", err)
		return oo.PackReturn(reqmsg, oo.EPARAM, nil)
	}

	txs, err := storage.GStore.ReadTxsByAddress(reqpara.Address, reqpara.Height1, reqpara.Height2)
	if nil != err {
		oo.LogD("OnReqGetAddrTxs storage.GStore.ReadTxsByAddress(%v) %v", reqpara, err)
		return oo.PackReturn(reqmsg, oo.ESERVER, nil)
	}

	mem_ids := txpool.Ids()
	var (
		set = make(map[string]bool)
	)
	for _, id := range mem_ids {
		set[id] = true
	}

	for _, tx := range txs {
		if "" != tx.BlockIndepHash || set[tx.ID] {
			rsppara.Data = append(rsppara.Data, tx.ID)
		}
	}

	return oo.PackReturn(reqmsg, oo.ESUCC, &rsppara)
}

func OnReqPrice(ctx *oo.ReqCtx, reqmsg *oo.RpcMsg) (rspmsg *oo.RpcMsg, err error) {
	reqpara, rsppara := &types.CmdPriceReq{}, &types.CmdPriceRsp{}
	if err = oo.JsonUnmarshalValidate(reqmsg.Para, reqpara); err != nil {
		oo.LogD("OnReqPrice JsonUnmarshalValidate err %v", err)
		return oo.PackReturn(reqmsg, oo.EPARAM, nil)
	}

	block := chain.TipBlock()
	if nil == block {
		oo.LogD("OnReqPrice err chain.TipBlock() == nil")
		return oo.PackReturn(reqmsg, oo.ESERVER, nil)
	}

	var new_target bool
	if "" != reqpara.Target {
		_, err = storage.GStore.ReadWalletByAddress(reqpara.Target)
		if nil != err {
			if storage.ErrNoRows == err {
				new_target = true
			} else {
				oo.LogD("OnReqPrice ReadWalletByAddress %v", err)
				return oo.PackReturn(reqmsg, oo.ESERVER, nil)
			}
		}
	}
	endow_fee, miner_fee := types.TxPayout(block.Height, reqpara.Bytes, new_target)

	rsppara.Amount = endow_fee + miner_fee

	return oo.PackReturn(reqmsg, oo.ESUCC, &rsppara)
}

func OnReqSubmitTx(ctx *oo.ReqCtx, reqmsg *oo.RpcMsg) (rspmsg *oo.RpcMsg, err error) {
	var reqpara types.CmdSubmitTxReq
	// var rsppara types.CmdSubmitTxRsp

	if err = oo.JsonUnmarshalValidate(reqmsg.Para, &reqpara); err != nil {
		oo.LogD("OnReqSubmitTx JsonUnmarshalValidate err %v", err)
		return oo.PackReturn(reqmsg, oo.EPARAM, nil)
	}
	PeerEventCtx(ctx, types.NSEVENT_PUSH_TX, nil)

	// if !GNodeServer.Chain.IsReady() {
	// 	oo.LogD("chain not ready.")
	// 	return oo.PackReturn(reqmsg, oo.ETIMENOTALLOW, nil)
	// }
	// block := chain.TipBlock()
	// if nil == block {
	// 	oo.LogD("OnReqSubmitTx err chain.GetBlock() == nil")
	// 	return oo.PackReturn(reqmsg, oo.ESERVER, nil)
	// }

	tx := &reqpara.Transaction

	// verify := chain.VerifySubmitTx(block.Height, tx)
	// if false == verify {
	// 	oo.LogD("OnReqSubmitTx chain.VerifySubmitTx err")
	// 	return oo.PackReturn(reqmsg, oo.ESERVER, nil)
	// }
	// tx.From = wallet.StringPublicKeyToAddress(tx.Owner)

	// oo.LogD("OnReqSubmitTx id[%s] from[%s] to[%s] amount[%.10f]",
	// 	tx.ID, tx.From, tx.Target, float64(tx.Quantity)/1e10)

	// GNodeServer.Chain.OnNewTx(ctx, tx, nil != ctx)

	err = ProcessSubmitTx(ctx, tx)
	if nil != err {
		if types.TX_VERIFY_ERROR == err {
			PeerEventCtx(ctx, types.NSEVENT_TX_ERROR, nil)
			if nil != ctx {
				ws, ok := ctx.Ctx.(*oo.WebSock)
				if ok {
					setBadTx(ws.PeerAddr(), tx.ID)
				}
			}
		}
		oo.LogD("OnReqSubmitTx ProcessSubmitTx err %v", err)
		return oo.PackReturn(reqmsg, err.Error(), nil)
	}

	return //oo.PackReturn(reqmsg, oo.ESUCC, &rsppara)
}

func ProcessSubmitTx(ctx *oo.ReqCtx, tx *types.Transaction) (err error) {
	submit_mutex.Lock()
	defer submit_mutex.Unlock()

	if true == getTxSet(tx.ID) {
		return
	}

	block := chain.TipBlock()
	if nil == block {
		err = oo.NewError("chain.GetBlock() == nil")
		return
	}

	verify := chain.VerifySubmitTx(block.Height, tx)
	if false == verify {
		// err = oo.NewError("chain.VerifySubmitTx err")
		// return
		return types.TX_VERIFY_ERROR
	}
	tx.From = wallet.StringPublicKeyToAddress(tx.Owner)

	oo.LogD("ProcessSubmitTx id[%s] from[%s] to[%s] amount[%.10f]",
		tx.ID, tx.From, tx.Target, float64(tx.Quantity)/1e10)

	if ctx != nil {
		if ws, ok := ctx.Ctx.(*oo.WebSock); ok {
			ResMarkPeer(tx.ID, PeerGetUuid(ws))
		}
	}
	GNodeServer.Chain.OnNewTx(tx, nil != ctx)

	setTxSet(tx.ID)

	return
}

func OnReqWalletNew(ctx *oo.ReqCtx, reqmsg *oo.RpcMsg) (rspmsg *oo.RpcMsg, err error) {
	reqpara, rsppara := &types.CmdWalletNewReq{}, &types.CmdWalletNewRsp{}
	// if err = oo.JsonUnmarshalValidate(reqmsg.Para, reqpara); err != nil {
	// 	oo.LogD("OnReqWalletNew JsonUnmarshalValidate err %v", err)
	// 	return oo.PackReturn(reqmsg, oo.EPARAM, nil)
	// }
	_ = reqpara

	w, err := wallet.NewWallet()
	if nil != err {
		oo.LogD("OnReqWalletNew wallet.NewWallet err %v", err)
		return oo.PackReturn(reqmsg, oo.ESERVER, nil)
	}

	rsppara.Address = w.Address()
	rsppara.PublicKey = w.PublicKey()
	rsppara.PrivateKey = w.PrivateKey()

	return oo.PackReturn(reqmsg, oo.ESUCC, &rsppara)
}

func OnReqWalletByAddr(ctx *oo.ReqCtx, reqmsg *oo.RpcMsg) (rspmsg *oo.RpcMsg, err error) {
	reqpara, rsppara := &types.CmdWalletByAddrReq{}, &types.CmdWalletByAddrRsp{}

	if err = oo.JsonUnmarshalValidate(reqmsg.Para, reqpara); err != nil {
		oo.LogD("OnReqWalletByAddr JsonUnmarshalValidate err %v", err)
		return oo.PackReturn(reqmsg, oo.EPARAM, nil)
	}

	if nil == storage.GStore {
		oo.LogD("OnReqWalletByAddr storer is nil.")
		return oo.PackReturn(reqmsg, oo.ESERVER, nil)
	}

	block := chain.TipBlock()
	if nil == block {
		oo.LogD("OnReqWalletByAddr err chain.GetBlock() == nil")
		return oo.PackReturn(reqmsg, oo.ESERVER, nil)
	}
	var (
		height = block.Height
		from   = reqpara.Address
	)

	wallet, err := storage.GStore.ReadWalletByAddress(reqpara.Address)
	if nil != err {
		if storage.ErrNoRows == err {
			rsppara.Pos = types.PledgeBase(height + 1)
			return oo.PackReturn(reqmsg, oo.ESUCC, rsppara)
		}
		oo.LogD("OnReqWalletByAddr storage.GStore.ReadWalletByAddress(%s) err %v", reqpara.Address, err)
		return oo.PackReturn(reqmsg, oo.ESERVER, nil)
	}

	wallet.LastTx, err = storage.GStore.ReadLastTxByAddress(reqpara.Address, "latest")
	if nil != err {
		oo.LogD("OnReqWalletByAddr storage.GStore.ReadLastTxByAddress(%s) err %v", reqpara.Address, err)
		return oo.PackReturn(reqmsg, oo.ESERVER, nil)
	}

	_, lock_amt, err := chain.GetAddrLockedBalance(height, from)
	if nil != err {
		oo.LogD("OnReqWalletByAddr chain.GetAddrLockedBalance(%d, %s) %v", height, from, err)
		return oo.PackReturn(reqmsg, oo.ESERVER, nil)
	}
	lock_amt = lock_amt * types.MINTOKEN_PER_TOKEN

	in, out, err := chain.GetPendingAmount(reqpara.Address)
	if nil != err {
		oo.LogD("OnReqWalletByAddr chain.GetPendingAmount(%s) %v", from, err)
		return oo.PackReturn(reqmsg, oo.ESERVER, nil)
	}

	needed_swr, _ := chain.GetMinerPledgeNeed(height+1, reqpara.Address)
	rsppara = &types.CmdWalletByAddrRsp{
		Pending: in + lock_amt,
		Balance: wallet.Balance - lock_amt - out,
		LastTx:  wallet.LastTx,
		Pos:     needed_swr,
	}

	if rsppara.Balance < 0 {
		oo.LogD("OnReqWalletByAddr err %s %v", from, *rsppara)
		// return oo.PackReturn(reqmsg, oo.ESERVER, nil)
		rsppara.Balance = 0
	}

	return oo.PackReturn(reqmsg, oo.ESUCC, rsppara)
}

func OnReqGetBlock(ctx *oo.ReqCtx, reqmsg *oo.RpcMsg) (rspmsg *oo.RpcMsg, err error) {
	var reqpara types.CmdGetBlockReq

	if err = oo.JsonUnmarshalValidate(reqmsg.Para, &reqpara); err != nil {
		oo.LogD("OnReqGetBlock JsonUnmarshalValidate err %v", err)
		return oo.PackReturn(reqmsg, oo.EPARAM, nil)
	}
	blk_id := reqpara.BlockId
	if 1 == reqpara.Current {
		blk_id = chain.ChainTipHash()
	}

	rsppara := &types.CmdGetBlockRsp{}
	if "" != blk_id {
		block, err := chain.GetBlock(blk_id)
		if nil != err {
			oo.LogD("OnReqGetBlock chain.GetBlock(%s) err %v", blk_id, err)
			return oo.PackReturn(reqmsg, oo.ESERVER, nil)
		}
		rsppara.Block = *block

	} else if reqpara.BlockHeight >= 0 {
		block, err := chain.GetHeightBlock(reqpara.BlockHeight)
		if nil != err {
			oo.LogD("OnReqGetBlock chain.GetHeightBlock(%d) err %v", reqpara.BlockHeight, err)
			return oo.PackReturn(reqmsg, oo.ESERVER, nil)
		}
		rsppara.Block = *block
	}

	return oo.PackReturn(reqmsg, oo.ESUCC, rsppara)
}

func OnReqGetWalletList(ctx *oo.ReqCtx, reqmsg *oo.RpcMsg) (rspmsg *oo.RpcMsg, err error) {
	var reqpara types.CmdGetWalletListReq

	if err = oo.JsonUnmarshalValidate(reqmsg.Para, &reqpara); err != nil {
		oo.LogD("OnReqGetWalletList JsonUnmarshalValidate err %v", err)
		return oo.PackReturn(reqmsg, oo.EPARAM, nil)
	}

	blk_id := reqpara.BlockId
	if 1 == reqpara.Current {
		blk_id = chain.ChainTipHash()
	}

	rsppara := &types.CmdGetWalletListRsp{}
	if "" != blk_id {
		block, err := chain.GetBlock(blk_id)
		if nil != err {
			oo.LogD("OnReqGetWalletList chain.GetBlock(%s) err %v", blk_id, err)
			return oo.PackReturn(reqmsg, oo.ESERVER, nil)
		}
		rsppara.WalletList = block.WalletList

	} else if reqpara.BlockHeight >= 0 {
		block, err := chain.GetHeightBlock(reqpara.BlockHeight)
		if nil != err {
			oo.LogD("OnReqGetWalletList chain.GetHeightBlock(%d) err %v", reqpara.BlockHeight, err)
			return oo.PackReturn(reqmsg, oo.ESERVER, nil)
		}
		rsppara.WalletList = block.WalletList
	}

	return oo.PackReturn(reqmsg, oo.ESUCC, rsppara)
}

func OnReqSubmitBlock(ctx *oo.ReqCtx, reqmsg *oo.RpcMsg) (rspmsg *oo.RpcMsg, err error) {
	var reqpara types.CmdSubmitBlockReq
	// var rsppara types.CmdSubmitBlockRsp

	if err = oo.JsonUnmarshalValidate(reqmsg.Para, &reqpara); err != nil {
		oo.LogD("OnReqSubmitBlock JsonUnmarshalValidate err %v", err)
		return oo.PackReturn(reqmsg, oo.EPARAM, nil)
	}
	ws, ok := ctx.Ctx.(*oo.WebSock)
	if !ok {
		return
	}
	PeerEvent(ws, types.NSEVENT_PUSH_BLOCK, nil)

	//checkbds, get all tx&data from peer
	// if len(reqpara.BDS) == 0 {
	// 	return oo.PackReturn(reqmsg, oo.EPARAM, nil)
	// }
	if !chain.GChain.IsReady() {
		return
	}

	b := &reqpara.NewBlock
	tip := GNodeServer.Chain.GetTip()

	if tip != nil && b.Height > tip.Height+1 {
		oo.LogD("Skip block: (%s) %d > %d+1", ws.PeerAddr(), b.Height, tip.Height)
		return
	}
	if b.PreviousBlock != tip.IndepHash {
		if b.IndepHash != tip.IndepHash {
			oo.LogD("Skip not tip block: %s %s, tip %s", b.IndepHash, b.PreviousBlock, tip.IndepHash)
		}
		return
	}

	if err = GNodeServer.ReadFullBlockFromPeer(ws, b); err != nil {
		oo.LogD("failed to read full block %d. err %v", b.Height, err)
		return
	}

	if !chain.GChain.IsReady() {
		return
	}
	BDS := []byte(reqpara.BDS)
	// if len(BDS) == 0 { //force gen by local
	poa, err := GNodeServer.Chain.GetPoa(oo.HexDecStringPad32(b.PreviousBlock), b.Height)
	if err != nil {
		oo.LogD("Failed to get poa %d. %v", b.Height, err)
		return oo.PackReturn(reqmsg, oo.ESERVER, nil)
	}
	if len(poa) > 0 {
		b.Poa = oo.HexEncodeToString(poa)
	}
	BDS, err = GNodeServer.Chain.GetBDS(b)
	if err != nil {
		oo.LogD("Failed to get bds %d. %v", b.Height, err)
		return oo.PackReturn(reqmsg, oo.ESERVER, nil)
	}
	// }
	// if ctx != nil {
	// 	if ws, ok := ctx.Ctx.(*oo.WebSock); ok {
	ResMarkPeer(reqpara.NewBlock.IndepHash, PeerGetUuid(ws))
	// 	}
	// }
	err = GNodeServer.Chain.OnNewBlock(&reqpara.NewBlock, BDS, true)
	if err == types.BLK_VERIFY_ERROR {
		PeerEvent(ws, types.NSEVENT_BLOCK_ERROR, nil)
	}

	return //oo.PackReturn(reqmsg, oo.ESUCC, nil)
}
