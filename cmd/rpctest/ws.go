package main

import (
	"fmt"

	"stashware/oo"
	"stashware/types"
	"stashware/wallet"
)

func wsTest() {
	r1, err := info()
	if nil != err {
		oo.LogD("rpcCall %v", err)
		return
	}
	fmt.Printf("==== ==== [%s] \n%s\n\n", "info", JsonData(r1))

	r2, err := peers()
	if nil != err {
		oo.LogD("rpcCall %v", err)
		return
	}
	fmt.Printf("==== ==== [%s] \n%s\n\n", "peers", JsonData(r2))

	// OnReqGetBlock
	// OnReqGetWalletList
	{
		r30, err := get_block(1, 0, "")
		if nil != err {
			oo.LogD("rpcCall %v", err)
			return
		}
		fmt.Printf("==== ==== [%s] \n%s\n\n", "get_block", JsonData(r30))

		r31, err := get_wallet_list(1, 0, "")
		if nil != err {
			oo.LogD("rpcCall %v", err)
			return
		}
		fmt.Printf("==== ==== [%s] \n%s\n\n", "get_wallet_list", JsonData(r31))

		// r4, err := get_block(0, r30.Height, "")
		// if nil != err {
		// 	oo.LogD("rpcCall %v", err)
		// 	return
		// }
		// fmt.Printf("==== ==== [%s] \n%s\n\n", "get_block", JsonData(r4))

		// r5, err := get_block(0, 0, r30.IndepHash)
		// if nil != err {
		// 	oo.LogD("rpcCall %v", err)
		// 	return
		// }
		// fmt.Printf("==== ==== [%s] \n%s\n\n", "get_block", JsonData(r5))
	}

	// OnReqWalletNew
	r3, err := wallet_new()
	if nil != err {
		oo.LogD("rpcCall %v", err)
		return
	}
	fmt.Printf("==== ==== [%s] \n%s\n\n", "wallet_new", JsonData(r3))

	// OnReqWalletByAddr
	r4, err := wallet_by_addr(w0.Address())
	if nil != err {
		oo.LogD("rpcCall %v", err)
		return
	}
	fmt.Printf("==== ==== [%s] \n%s\n\n", "wallet_by_addr", JsonData(r4))

	// OnReqGetAddrTxs
	r5, err := get_addr_txs(w0.Address())
	if nil != err {
		oo.LogD("rpcCall %v", err)
		return
	}
	fmt.Printf("==== ==== [%s] \n%s\n\n", "get_addr_txs", JsonData(r5))

	// OnReqPrice
	str := "00000068656c6c6f776f726c64"
	buf := oo.HexDecodeString(str)
	cnt := int64(len(buf))

	r6, err := price(cnt, r3.Address)
	if nil != err {
		oo.LogD("rpcCall %v", err)
		return
	}
	fmt.Printf("==== ==== [%s] \n%s\n\n", "price", JsonData(r6))

	// OnReqSubmitTx
	var tx *types.Transaction
	{
		tx = &types.Transaction{
			LastTx:   r4.LastTx,
			Owner:    w0.PublicKey(),
			From:     wallet.StringPublicKeyToAddress(w0.PublicKey()),
			Target:   r3.Address,
			Quantity: 1e10,
			Reward:   r6.Amount,
			Data:     str,
			DataHash: oo.HexEncToStringPad32(oo.Sha256(buf)),
			Tags:     []types.Tag{types.NewTag("name", "value")},
		}
		raw_msg := tx.SignatureDataSegment()
		msg := oo.Sha256(raw_msg)
		tx.ID = oo.HexEncToStringPad32(msg)

		sign, err := wallet.Sign(w0.PrivateKey(), msg)
		if nil != err {
			oo.LogD("wallet.Sign %v", err)
			return
		}
		tx.Signature = oo.HexEncodeToString(sign)

		r7, err := submit_tx(tx)
		if nil != err {
			oo.LogD("rpcCall %v", err)
			return
		}
		fmt.Printf("==== ==== [%s] \n%s\n%x\n\n", "submit_tx", JsonData(r7), raw_msg)
	}

	// OnReqTxPending
	r8, err := tx_pending()
	if nil != err {
		oo.LogD("rpcCall %v", err)
		return
	}
	fmt.Printf("==== ==== [%s] \n%s\n\n", "tx_pending", JsonData(r8))

	// OnReqTxById
	r9, err := tx_by_id(tx.ID)
	if nil != err {
		oo.LogD("rpcCall %v", err)
		return
	}
	fmt.Printf("==== ==== [%s] \n%s\n\n", "tx_by_id", JsonData(r9))

	// OnReqTxData
	r10, err := tx_data(tx.ID)
	if nil != err {
		oo.LogD("rpcCall %v", err)
		return
	}
	fmt.Printf("==== ==== [%s] \n%s\n\n", "tx_data", JsonData(r10))

	// todo OnReqSubmitBlock
}
