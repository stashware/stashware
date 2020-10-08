package node

import (
	"stashware/oo"
	"stashware/types"
)

func RpcReadBlockHeight(ws *oo.WebSock, height int64) (b *types.Block, err error) {
	var req types.CmdGetBlockReq
	var rsp types.CmdGetBlockRsp
	req.BlockHeight = height
	rpcmsg := oo.PackRequest(types.CMD_GET_BLOCK, req, nil)
	if err = ws.RpcRequest(rpcmsg, types.WAIT_RPC_SHORT, &rsp); err != nil {
		// oo.LogW("Failed to read block. %d - %s", req.BlockHeight)
		return
	}
	b = &rsp.Block
	return
}

func RpcReadTxById(ws *oo.WebSock, txid string) (tx *types.Transaction, err error) {
	var txreq types.CmdTxByIdReq
	var txrsp types.CmdTxByIdRsp

	txreq.Txid = txid
	rpcmsg := oo.PackRequest(types.CMD_TX_BY_ID, txreq, nil)
	if err = ws.RpcRequest(rpcmsg, types.WAIT_RPC_SHORT, &txrsp); err != nil {
		return
	}
	tx = &txrsp.Transaction
	return
}

func RpcReadTxData(ws *oo.WebSock, txid string) (data string, err error) {
	var txdatareq types.CmdTxDataReq
	var txdatarsp types.CmdTxDataRsp
	txdatareq.Txid = txid
	rpcmsg := oo.PackRequest(types.CMD_TX_DATA, txdatareq, nil)
	if err = ws.RpcRequest(rpcmsg, types.WAIT_RPC_LONG, &txdatarsp); err != nil {
		return
	}
	data = txdatarsp
	return
}

func RpcReadTxPending(ws *oo.WebSock) (ret []string, err error) {
	var (
		req_para = &types.CmdTxPendingReq{}
		rsp_para = &types.CmdTxPendingRsp{}
	)
	rpcmsg := oo.PackRequest(types.CMD_TX_PENDING, req_para, nil)

	err = ws.RpcRequest(rpcmsg, types.WAIT_RPC_SHORT, rsp_para)
	if nil != err {
		return
	}
	ret = rsp_para.Txs

	return
}

func RpcReadTxAndData(ws *oo.WebSock, txid string) (tx *types.Transaction, err error) {
	var (
		req_para = &types.CmdTxByIdReq{
			Txid: txid,
		}
		rsp_para = &types.CmdTxByIdRsp{}
	)
	rpcmsg := oo.PackRequest(types.CMD_TX_BY_ID, req_para, nil)

	if true == getTxSet(txid) {
		return
	}
	err = ws.RpcRequest(rpcmsg, types.WAIT_RPC_SHORT, rsp_para)
	if nil != err {
		PeerEvent(ws, types.NSEVENT_REQUEST_TIMEOUT, nil)
		return
	}
	tx = &rsp_para.Transaction

	if "" != tx.DataHash {
		var (
			req_para = &types.CmdTxDataReq{
				Txid: txid,
			}
			rsp_para types.CmdTxDataRsp
		)
		rpcmsg := oo.PackRequest(types.CMD_TX_DATA, req_para, nil)

		if true == getTxSet(txid) {
			return
		}
		err = ws.RpcRequest(rpcmsg, types.WAIT_RPC_LONG, &rsp_para)
		if nil != err {
			PeerEvent(ws, types.NSEVENT_REQUEST_TIMEOUT, nil)
			return
		}
		tx.Data = rsp_para
	}

	return
}

func RpcGetWalletList(ws *oo.WebSock, block_id string) (wlist []types.WalletTuple, err error) {
	var wallets_req types.CmdGetWalletListReq
	var wallets_rsp types.CmdGetWalletListRsp

	wallets_req.BlockId = block_id
	rpcmsg := oo.PackRequest(types.CMD_GET_WALLET_LIST, wallets_req, nil)
	if err = ws.RpcRequest(rpcmsg, types.WAIT_RPC_MIDDLE, &wallets_rsp); err != nil {
		return
	}
	wlist = wallets_rsp.WalletList
	return
}

func RpcInfo(ws *oo.WebSock, info *types.CmdInfoRsp) (info_rsp *types.CmdInfoRsp, err error) {
	rpcmsg := oo.PackRequest(types.CMD_INFO, info, nil)

	info_rsp = &types.CmdInfoRsp{}
	if err = ws.RpcRequest(rpcmsg, types.WAIT_RPC_SHORT, info_rsp); err != nil {
		// oo.LogD("request %s err %v", rpcmsg.Cmd, err) //too must error because old version Error
		return
	}
	return
}

func RpcPeers(ws *oo.WebSock) (peers []string, err error) {
	rpcmsg := oo.PackRequest(types.CMD_PEERS, nil, nil)

	rsp := types.CmdPeersRsp{}
	if err = ws.RpcRequest(rpcmsg, types.WAIT_RPC_SHORT, &rsp); err != nil {
		oo.LogD("request %s err %v", rpcmsg.Cmd, err)
		return
	}
	peers = rsp.EndPoint
	return
}
