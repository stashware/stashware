package main

import (
	"fmt"

	"stashware/oo"
	"stashware/types"

	"github.com/urfave/cli/v2"
)

func rpcCall(cctx *cli.Context, req *oo.RpcMsg, parse interface{}) (rsp *oo.RpcMsg, err error) {
	var (
		host = cctx.String("host")
		port = cctx.Int64("port")
	)
	if "" == host || port <= 0 {
		err = oo.NewError(`host[%s] port[%d]`, host, port)
		return
	}

	var (
		addr = fmt.Sprintf("%s:%d", host, port)
		stop = make(chan struct{})
	)
	ws, err := oo.InitWsClient("ws", addr, "/", oo.NewChannelManager(8),
		func(c *oo.WebSock, data []byte) (b []byte, e error) {
			if len(data) > 0 {
				rsp = &oo.RpcMsg{}
				err = oo.JsonUnmarshal(data, rsp)
				if nil != err {
					return
				}
				if nil != parse && len(rsp.Para) > 0 {
					err = oo.JsonUnmarshal(rsp.Para, parse)
				}
			}

			stop <- struct{}{}

			return
		})
	if nil != err {
		err = oo.NewError("InitWsClient err %v", err)
		return
	}
	defer ws.Close()

	go ws.StartDial(5, 0)

	err = ws.SendData(oo.JsonData(req))
	if nil != err {
		err = oo.NewError("SendData err %v", err)
		return
	}

	<-stop

	return
}

func printResult(format string, a ...interface{}) {
	fmt.Printf(format, a...)
	fmt.Println()
}

var info_cmd = &cli.Command{
	Name:  "info",
	Usage: "current blockchain info",
	Action: func(cctx *cli.Context) (err error) {
		// rsp := &types.CmdInfoRsp{}
		ret, err := rpcCall(cctx, &oo.RpcMsg{Cmd: types.CMD_INFO}, nil)
		if nil != err {
			err = oo.NewError("rpcCall %v", err)
			return
		}
		if nil != ret.Err {
			err = oo.NewError("%v", ret.Err)
			return
		}

		if len(ret.Para) > 0 {
			printResult("%s", ret.Para)
		}

		return
	},
}

var peers_cmd = &cli.Command{
	Name:  "peers",
	Usage: "current connecting peers address",
	Action: func(cctx *cli.Context) (err error) {
		ret, err := rpcCall(cctx, &oo.RpcMsg{Cmd: types.CMD_PEERS}, nil)
		if nil != err {
			err = oo.NewError("rpcCall %v", err)
			return
		}
		if nil != ret.Err {
			err = oo.NewError("%v", ret.Err)
			return
		}

		if len(ret.Para) > 0 {
			printResult("%s", ret.Para)
		}

		return
	},
}

var tx_pending_cmd = &cli.Command{
	Name:  "tx_pending",
	Usage: "tx id in memory pool",
	Action: func(cctx *cli.Context) (err error) {
		ret, err := rpcCall(cctx, &oo.RpcMsg{Cmd: types.CMD_TX_PENDING}, nil)
		if nil != err {
			err = oo.NewError("rpcCall %v", err)
			return
		}
		if nil != ret.Err {
			err = oo.NewError("%v", ret.Err)
			return
		}

		if len(ret.Para) > 0 {
			printResult("%s", ret.Para)
		}

		return
	},
}

var get_block_cmd = &cli.Command{
	Name:  "get_block",
	Usage: "get block by hash or height, default return the latest block",
	Flags: []cli.Flag{
		&cli.Int64Flag{
			Name:  "block_height",
			Value: -1,
			Usage: "(interger, optional) The block height",
		},
		&cli.StringFlag{
			Name:  "block_id",
			Usage: "(string, optional) The block hash",
		},
	},
	Action: func(cctx *cli.Context) (err error) {
		var (
			blk_id = cctx.String("block_id")
			height = cctx.Int64("block_height")
		)
		req := &types.CmdGetBlockReq{}

		if "" == blk_id && -1 == height {
			req.Current = 1

		} else if "" != blk_id {
			req.BlockId = blk_id

		} else if height >= 0 {
			req.BlockHeight = height

		} else {
			err = oo.NewError("error args")
			return
		}

		ret, err := rpcCall(cctx, &oo.RpcMsg{Cmd: types.CMD_GET_BLOCK, Para: oo.JsonData(req)}, nil)
		if nil != err {
			err = oo.NewError("rpcCall %v", err)
			return
		}
		if nil != ret.Err {
			err = oo.NewError("%v", ret.Err)
			return
		}

		if len(ret.Para) > 0 {
			printResult("%s", ret.Para)
		}

		return
	},
}

var tx_by_id_cmd = &cli.Command{
	Name:  "tx_by_id",
	Usage: "get transaction by txid",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "txid",
			Usage: "(string, required) The transaction id",
		},
	},
	Action: func(cctx *cli.Context) (err error) {
		var (
			txid = cctx.String("txid")
		)
		req := &types.CmdTxByIdReq{}

		if len(txid) <= 0 {
			err = oo.NewError("error args")
			return
		}
		req.Txid = txid

		ret, err := rpcCall(cctx, &oo.RpcMsg{Cmd: types.CMD_TX_BY_ID, Para: oo.JsonData(req)}, nil)
		if nil != err {
			err = oo.NewError("rpcCall %v", err)
			return
		}
		if nil != ret.Err {
			err = oo.NewError("%v", ret.Err)
			return
		}

		if len(ret.Para) > 0 {
			printResult("%s", ret.Para)
		}

		return
	},
}

var tx_data_cmd = &cli.Command{
	Name:  "tx_data",
	Usage: "get tx data, which show in base58 encoded",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "txid",
			Usage: "(string, required) The transaction id",
		},
		&cli.BoolFlag{
			Name:  "hex",
			Value: false,
			Usage: "(bool, optional) Show data with hex encode",
		},
	},
	Action: func(cctx *cli.Context) (err error) {
		req := &types.CmdTxDataReq{
			Txid: cctx.String("txid"),
		}
		ret, err := rpcCall(cctx, &oo.RpcMsg{Cmd: types.CMD_TX_DATA, Para: oo.JsonData(req)}, nil)
		if nil != err {
			err = oo.NewError("rpcCall %v", err)
			return
		}
		if nil != ret.Err {
			err = oo.NewError("%v", ret.Err)
			return
		}

		if len(ret.Para) > 0 {
			var str string
			oo.JsonUnmarshal(ret.Para, &str)

			if cctx.Bool("hex") {
				printResult("%s", str)

			} else {
				printResult("%s", oo.HexDecodeString(str))
			}
		}

		return
	},
}

var price_cmd = &cli.Command{
	Name:  "price",
	Usage: "get cost of saving data to blockchain",
	Flags: []cli.Flag{
		&cli.Int64Flag{
			Name:  "bytes",
			Usage: "(interger, required) Bytes of data length",
		},
		&cli.StringFlag{
			Name:  "target",
			Usage: "(string, required) The sender addres",
		},
	},
	Action: func(cctx *cli.Context) (err error) {
		req := &types.CmdPriceReq{
			Bytes:  cctx.Int64("bytes"),
			Target: cctx.String("target"),
		}
		ret, err := rpcCall(cctx, &oo.RpcMsg{Cmd: types.CMD_PRICE, Para: oo.JsonData(req)}, nil)
		if nil != err {
			err = oo.NewError("rpcCall %v", err)
			return
		}
		if nil != ret.Err {
			err = oo.NewError("%v", ret.Err)
			return
		}

		if len(ret.Para) > 0 {
			printResult("%s", ret.Para)
		}

		return
	},
}

var wallet_new_cmd = &cli.Command{
	Name:  "wallet_new",
	Usage: "new wallet and return private-key, public-key, address, which are encoded by base58",
	Flags: []cli.Flag{},
	Action: func(cctx *cli.Context) (err error) {
		req := &types.CmdWalletNewReq{}
		ret, err := rpcCall(cctx, &oo.RpcMsg{Cmd: types.CMD_WALLET_NEW, Para: oo.JsonData(req)}, nil)
		if nil != err {
			err = oo.NewError("rpcCall %v", err)
			return
		}
		if nil != ret.Err {
			err = oo.NewError("%v", ret.Err)
			return
		}

		if len(ret.Para) > 0 {
			printResult("%s", ret.Para)
		}

		return
	},
}

var wallet_by_addr_cmd = &cli.Command{
	Name:  "wallet_by_addr",
	Usage: "get address wallet info",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "address",
			Usage: "(string, required) The swr wallet address",
		},
	},
	Action: func(cctx *cli.Context) (err error) {
		var (
			address = cctx.String("address")
		)
		req := &types.CmdWalletByAddrReq{}

		if len(address) <= 0 {
			err = oo.NewError("error args")
			return
		}
		req.Address = address

		ret, err := rpcCall(cctx, &oo.RpcMsg{Cmd: types.CMD_WALLET_BY_ADDR, Para: oo.JsonData(req)}, nil)
		if nil != err {
			err = oo.NewError("rpcCall %v", err)
			return
		}
		if nil != ret.Err {
			err = oo.NewError("%v", ret.Err)
			return
		}

		if len(ret.Para) > 0 {
			printResult("%s", ret.Para)
		}

		return
	},
}

var submit_block_cmd = &cli.Command{
	Name:  "submit_block",
	Usage: "submit new block to swr blockchain",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "data",
			Usage: "(string, required) The block json encode data",
		},
		&cli.StringFlag{
			Name:  "bds",
			Usage: "(string, required) The block data segment, base58 encode",
		},
	},
	Action: func(cctx *cli.Context) (err error) {
		var (
			data = cctx.String("data")
			bds  = cctx.String("bds")
		)
		req := &types.CmdSubmitBlockReq{}

		if len(data) <= 0 || len(bds) <= 0 {
			err = oo.NewError("error args")
			return
		}
		req.BDS = bds

		var block types.Block
		err = oo.JsonUnmarshal([]byte(data), &block)
		if nil != err {
			err = oo.NewError("data json decode %v", err)
			return
		}
		req.NewBlock = block

		ret, err := rpcCall(cctx, &oo.RpcMsg{Cmd: types.CMD_SUBMIT_BLOCK, Para: oo.JsonData(req)}, nil)
		if nil != err {
			err = oo.NewError("rpcCall %v", err)
			return
		}
		if nil != ret.Err {
			err = oo.NewError("%v", ret.Err)
			return
		}

		if len(ret.Para) > 0 {
			printResult("%s", ret.Para)
		}

		return
	},
}

var submit_tx_cmd = &cli.Command{
	Name:  "submit_tx",
	Usage: "send new tx to swr blockchain",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "data",
			Usage: "(string, required) The transaction json encode data",
		},
	},
	Action: func(cctx *cli.Context) (err error) {
		var (
			data = cctx.String("data")
		)
		req := &types.CmdSubmitTxReq{}

		if len(data) <= 0 {
			err = oo.NewError("error args")
			return
		}

		var tx types.Transaction
		err = oo.JsonUnmarshal([]byte(data), &tx)
		if nil != err {
			err = oo.NewError("data json decode %v", err)
			return
		}
		req.Transaction = tx

		ret, err := rpcCall(cctx, &oo.RpcMsg{Cmd: types.CMD_SUBMIT_TX, Para: oo.JsonData(req)}, nil)
		if nil != err {
			err = oo.NewError("rpcCall %v", err)
			return
		}
		if nil != ret.Err {
			err = oo.NewError("%v", ret.Err)
			return
		}

		if len(ret.Para) > 0 {
			printResult("%s", ret.Para)
		}

		return
	},
}

var get_wallet_list_cmd = &cli.Command{
	Name:  "get_wallet_list",
	Usage: "get the wallet list in block",
	Flags: []cli.Flag{
		&cli.Int64Flag{
			Name:  "block_height",
			Value: -1,
			Usage: "(interger, optional) The block height",
		},
		&cli.StringFlag{
			Name:  "block_id",
			Usage: "(string, optional) The block hash",
		},
	},
	Action: func(cctx *cli.Context) (err error) {
		var (
			blk_id = cctx.String("block_id")
			height = cctx.Int64("block_height")
		)
		req := &types.CmdGetWalletListReq{}

		if "" == blk_id && -1 == height {
			req.Current = 1

		} else if "" != blk_id {
			req.BlockId = blk_id

		} else if height >= 0 {
			req.BlockHeight = height

		} else {
			err = oo.NewError("error args")
			return
		}

		ret, err := rpcCall(cctx, &oo.RpcMsg{Cmd: types.CMD_GET_WALLET_LIST, Para: oo.JsonData(req)}, nil)
		if nil != err {
			err = oo.NewError("rpcCall %v", err)
			return
		}
		if nil != ret.Err {
			err = oo.NewError("%v", ret.Err)
			return
		}

		if len(ret.Para) > 0 {
			printResult("%s", ret.Para)
		}

		return
	},
}

var get_addr_txs_cmd = &cli.Command{
	Name:  "get_addr_txs",
	Usage: "get address txs id in [height1, height2]",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "address",
			Usage: "(string, required) The wallet address",
		},
		&cli.Int64Flag{
			Name:  "height1",
			Value: -1,
			Usage: "(interger, optional) The block height",
		},
		&cli.Int64Flag{
			Name:  "height2",
			Value: -1,
			Usage: "(interger, optional) The block height",
		},
	},
	Action: func(cctx *cli.Context) (err error) {
		var (
			address = cctx.String("address")
			height1 = cctx.Int64("height1")
			height2 = cctx.Int64("height2")
		)
		req := &types.CmdGetAddrTxsReq{
			Address: address,
		}
		if -1 != height1 {
			req.Height1 = &height1
		}
		if -1 != height2 {
			req.Height2 = &height2
		}

		ret, err := rpcCall(cctx, &oo.RpcMsg{Cmd: types.CMD_GET_ADDR_TXS, Para: oo.JsonData(req)}, nil)
		if nil != err {
			err = oo.NewError("rpcCall %v", err)
			return
		}
		if nil != ret.Err {
			err = oo.NewError("%v", ret.Err)
			return
		}

		printResult("%s", ret.Para)

		return
	},
}
