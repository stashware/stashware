package main

import (
	"fmt"
	"log"

	"stashware/oo"
	"stashware/types"
	WALLET "stashware/wallet"
)

func restTest() {
	var err error

	// info
	var info types.CmdInfoRsp
	err = restCall("/info", nil, &info)
	if nil != err {
		log.Fatal(err)
	}
	fmt.Printf("==== ==== [%s] \n%s\n\n", "info", JsonData(info))

	// peers
	var peers types.CmdPeersRsp
	err = restCall("/peers", nil, &peers)
	if nil != err {
		log.Fatal(err)
	}
	fmt.Printf("==== ==== [%s] \n%s\n\n", "peers", JsonData(peers))

	// block
	var block types.CmdGetBlockRsp
	err = restCall("/block/height/1", nil, &block)
	if nil != err {
		log.Fatal(err)
	}
	fmt.Printf("==== ==== [%s] \n%s\n\n", "block", JsonData(block))

	err = restCall("/block/hash/"+block.IndepHash, nil, &block)
	if nil != err {
		log.Fatal(err)
	}
	fmt.Printf("==== ==== [%s] \n%s\n\n", "block", JsonData(block))

	// wallet
	var wallet types.CmdWalletNewRsp
	err = restCall("/wallet/new", nil, &wallet)
	if nil != err {
		log.Fatal(err)
	}
	fmt.Printf("==== ==== [%s] \n%s\n\n", "wallet", JsonData(wallet))

	// wallet_by_addr
	var wallet_by_addr types.CmdWalletByAddrRsp
	err = restCall("/wallet/"+w0.Address(), nil, &wallet_by_addr)
	if nil != err {
		log.Fatal(err)
	}
	fmt.Printf("==== ==== [%s] \n%s\n\n", "wallet_by_addr", JsonData(wallet_by_addr))

	// get_addr_txs
	var get_addr_txs types.CmdGetAddrTxsRsp
	err = restCall("/txs/"+w0.Address(), nil, &get_addr_txs)
	if nil != err {
		log.Fatal(err)
	}
	fmt.Printf("==== ==== [%s] \n%s\n\n", "get_addr_txs", JsonData(get_addr_txs))

	// str := "00000068656c6c6f776f726c64"
	// buf := oo.HexDecodeString(str)
	// cnt := int64(len(buf))
	cnt := 0

	// price
	var price types.CmdPriceRsp
	err = restCall(fmt.Sprintf("/price/%d/%s", cnt, wallet.Address), nil, &price)
	if nil != err {
		log.Fatal(err)
	}
	fmt.Printf("==== ==== [%s] \n%s\n\n", "price", JsonData(price))

	// submit_tx
	tx := &types.Transaction{
		LastTx:   wallet_by_addr.LastTx,
		Owner:    w0.PublicKey(),
		From:     WALLET.StringPublicKeyToAddress(w0.PublicKey()),
		Target:   wallet.Address,
		Quantity: 1e10,
		Reward:   price.Amount,
		// Data:     str,
		// DataHash: oo.HexEncToStringPad32(oo.Sha256(buf)),
		Tags: []types.Tag{types.NewTag("name", "value")},
	}
	raw_msg := tx.SignatureDataSegment()
	msg := oo.Sha256(raw_msg)
	tx.ID = oo.HexEncToStringPad32(msg)

	sign, err := WALLET.Sign(w0.PrivateKey(), msg)
	if nil != err {
		oo.LogD("wallet.Sign %v", err)
		return
	}
	tx.Signature = oo.HexEncodeToString(sign)

	var submit_tx types.CmdSubmitTxRsp
	err = restCall("/tx", JsonData(tx), &submit_tx)
	if nil != err {
		log.Fatal(err)
	}
	fmt.Printf("==== ==== [%s] \n%s\n\n", "submit_tx", JsonData(submit_tx))

	// tx_pending
	var tx_pending types.CmdTxPendingRsp
	err = restCall("/tx/pending", nil, &tx_pending)
	if nil != err {
		log.Fatal(err)
	}
	fmt.Printf("==== ==== [%s] \n%s\n\n", "tx_pending", JsonData(tx_pending))

	// tx_by_id
	var tx_by_id types.CmdTxByIdRsp
	err = restCall(fmt.Sprintf("/tx/%s", tx.ID), nil, &tx_by_id)
	if nil != err {
		log.Fatal(err)
	}
	fmt.Printf("==== ==== [%s] \n%s\n\n", "tx_by_id", JsonData(tx_by_id))

	// tx_data
	// tmp, err := oo.HttpRequest("GET", fmt.Sprintf("http://%s/%s", rest_addr, tx.ID), nil, nil)
	// if nil != err {
	// 	log.Fatal(err)
	// }
	// fmt.Printf("==== ==== [%s] \n%0x\n\n", "tx_data", tmp)
}

func restCall(uri string, body []byte, parse interface{}) error {
	url := fmt.Sprintf("http://%s%s", rest_addr, uri)

	return oo.HttpRequestAndParse("POST", url, nil, body, parse)
}
