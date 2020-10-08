package main

import (
	"encoding/json"
	"fmt"

	"stashware/oo"
	"stashware/types"
)

func info() (ret types.CmdInfoRsp, err error) {
	msg := &oo.RpcMsg{
		Cmd: types.CMD_INFO,
	}
	err = rpcCall(msg, nil, &ret)
	if nil != err {
		return
	}

	return
}

func peers() (ret types.CmdPeersRsp, err error) {
	msg := &oo.RpcMsg{
		Cmd: types.CMD_PEERS,
	}
	err = rpcCall(msg, &struct{}{}, &ret)
	if nil != err {
		return
	}

	return
}

func get_block(current, height int64, blkid string) (ret types.CmdGetBlockRsp, err error) {
	var (
		msg = &oo.RpcMsg{
			Cmd: types.CMD_GET_BLOCK,
		}
		para = &types.CmdGetBlockReq{
			Current:     current,
			BlockHeight: height,
			BlockId:     blkid,
		}
	)
	err = rpcCall(msg, para, &ret)
	if nil != err {
		return
	}

	return
}

func get_wallet_list(current, height int64, blkid string) (ret types.CmdGetWalletListRsp, err error) {
	var (
		msg = &oo.RpcMsg{
			Cmd: types.CMD_GET_WALLET_LIST,
		}
		para = &types.CmdGetWalletListReq{
			Current:     current,
			BlockHeight: height,
			BlockId:     blkid,
		}
	)
	err = rpcCall(msg, para, &ret)
	if nil != err {
		return
	}

	return
}

func wallet_new() (ret types.CmdWalletNewRsp, err error) {
	var (
		msg = &oo.RpcMsg{
			Cmd: types.CMD_WALLET_NEW,
		}
		para = &types.CmdWalletNewReq{}
	)
	err = rpcCall(msg, para, &ret)
	if nil != err {
		return
	}

	return
}

func wallet_by_addr(addr string) (ret types.CmdWalletByAddrRsp, err error) {
	var (
		msg = &oo.RpcMsg{
			Cmd: types.CMD_WALLET_BY_ADDR,
		}
		para = &types.CmdWalletByAddrReq{
			Address: addr,
		}
	)
	err = rpcCall(msg, para, &ret)
	if nil != err {
		return
	}

	return
}

func get_addr_txs(addr string) (ret types.CmdGetAddrTxsRsp, err error) {
	var (
		msg = &oo.RpcMsg{
			Cmd: types.CMD_GET_ADDR_TXS,
		}
		para = &types.CmdGetAddrTxsReq{
			Address: addr,
		}
	)
	err = rpcCall(msg, para, &ret)
	if nil != err {
		return
	}

	return
}

func price(bytes int64, target string) (ret types.CmdPriceRsp, err error) {
	var (
		msg = &oo.RpcMsg{
			Cmd: types.CMD_PRICE,
		}
		para = &types.CmdPriceReq{
			Bytes:  bytes,
			Target: target,
		}
	)
	err = rpcCall(msg, para, &ret)
	if nil != err {
		return
	}

	return
}

func submit_tx(tx *types.Transaction) (ret types.CmdSubmitTxRsp, err error) {
	var (
		msg = &oo.RpcMsg{
			Cmd: types.CMD_SUBMIT_TX,
		}
		para = &types.CmdSubmitTxReq{
			Transaction: *tx,
		}
	)
	err = rpcCall(msg, para, &ret)
	if nil != err {
		return
	}

	return
}

func tx_pending() (ret types.CmdTxPendingRsp, err error) {
	var (
		msg = &oo.RpcMsg{
			Cmd: types.CMD_TX_PENDING,
		}
		para = &types.CmdTxPendingReq{}
	)
	err = rpcCall(msg, para, &ret)
	if nil != err {
		return
	}

	return
}

func tx_by_id(txid string) (ret types.CmdTxByIdRsp, err error) {
	var (
		msg = &oo.RpcMsg{
			Cmd: types.CMD_TX_BY_ID,
		}
		para = &types.CmdTxByIdReq{
			Txid: txid,
		}
	)
	err = rpcCall(msg, para, &ret)
	if nil != err {
		return
	}

	return
}

func tx_data(txid string) (ret types.CmdTxDataRsp, err error) {
	var (
		msg = &oo.RpcMsg{
			Cmd: types.CMD_TX_DATA,
		}
		para = &types.CmdTxDataReq{
			Txid: txid,
		}
	)
	err = rpcCall(msg, para, &ret)
	if nil != err {
		return
	}

	return
}

// ==================================================================

func JsonData(reqjson interface{}) []byte {
	if reqjson == nil {
		return []byte("")
	}

	data, err := json.MarshalIndent(reqjson, "", "  ")
	if err != nil {
		return []byte("")
	}

	return data
}

var (
	ws      *oo.WebSock
	recv_ch = make(chan *oo.RpcMsg, 1)
)

func init() {
	var err error
	ws, err = oo.InitWsClient("ws", ws_addr, "/", oo.NewChannelManager(8),
		func(c *oo.WebSock, data []byte) (b []byte, e error) {
			if len(data) > 0 {
				rsp := &oo.RpcMsg{}
				err := oo.JsonUnmarshal(data, rsp)
				if nil != err {
					oo.LogD("JsonUnmarshal(%s) err %v", data, err)
					return
				}
				recv_ch <- rsp
			}
			return
		})
	if nil != err {
		oo.LogD("InitWsClient err %v", err)
		return
	}

	go ws.StartDial(5, 0)
}

func rpcCall(req *oo.RpcMsg, para, parse interface{}) (err error) {
	if nil == req {
		err = fmt.Errorf("nil == req")
		return
	}
	if nil != para {
		req.Para = oo.JsonData(para)
	}

	err = ws.SendData(oo.JsonData(req))
	if nil != err {
		err = fmt.Errorf("SendData err %v", err)
		return
	}

	rsp := <-recv_ch

	if nil != rsp.Err {
		err = fmt.Errorf("%v", rsp.Err)
		return
	}

	if nil != parse && len(rsp.Para) > 0 {
		err = oo.JsonUnmarshal(rsp.Para, parse)
	}

	return
}
