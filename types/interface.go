package types

import (
	"errors"
	"stashware/oo"
)

var (
	BLK_VERIFY_ERROR = errors.New("block_verify_error")
	TX_VERIFY_ERROR  = errors.New("tx_verify_error")
)

type GossipIf interface {
	SendNewBlock(b *Block, BDS []byte)
	SendNewTx(tx *Transaction)
	ReadFullBlockFromPeer(ws *oo.WebSock, b *Block) (err error)
}

type ChainIf interface {
	GetActiveChain() (cs *TbChain, b *Block, err error)
	GetTip() (b *Block)
	ForkChain(b *Block) (err error)
	// SwitchChain(oldc *TbChain, newc *TbChain, new_ids []string) (err error)
	OnNewBlock(b *Block, BDS []byte, out bool) (err error)
	OnNewTx(tx *Transaction, out bool)
	GetPoa(preHash []byte, nowHeight int64) (poa []byte, err error)
	GetBDS(block *Block) (BDS []byte, err error)
	IsReady() bool
	SetReady(ready bool)
	// TryAcceptNewBlock(ctx *oo.ReqCtx, b *Block)
}
