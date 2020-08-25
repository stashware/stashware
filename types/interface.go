package types

import (
	"stashware/oo"
)

type GossipIf interface {
	SendNewBlock(ctx *oo.ReqCtx, b *Block, BDS []byte)
	SendNewTx(ctx *oo.ReqCtx, tx *Transaction)
	CtxEvent(ev int, ctx *oo.ReqCtx)
	ReadFullBlockFromPeer(ctx *oo.ReqCtx, ws *oo.WebSock, b *Block) (err error)
}

type ChainIf interface {
	GetActiveChain() (cs *TbChain, b *Block, err error)
	GetTip() (b *Block)
	ForkChain(cs *TbChain) (err error)
	// SwitchChain(oldc *TbChain, newc *TbChain, new_ids []string) (err error)
	OnNewBlock(ctx *oo.ReqCtx, b *Block, BDS []byte, out bool) (updated bool)
	OnNewTx(ctx *oo.ReqCtx, tx *Transaction, out bool)
	GetPoa(preHash []byte, nowHeight int64) (poa []byte, err error)
	GetBDS(block *Block) (BDS []byte, err error)
	IsReady() bool
	SetReady(ready bool)
	CheckPoint(h int64, indepHash string) bool
	// TryAcceptNewBlock(ctx *oo.ReqCtx, b *Block)
}
