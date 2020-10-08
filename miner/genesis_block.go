package miner

import (
	// "math/big"
	"fmt"
	"stashware/oo"
	"stashware/types"
	"time"
	_ "time"
)

type GenChain struct {
}

func (*GenChain) GetActiveChain() (cs *types.TbChain, b *types.Block, err error) {
	return
}
func (*GenChain) GetTip() (b *types.Block) {
	return
}
func (*GenChain) OnNewBlock(b *types.Block, BDS []byte, out bool) (err error) {
	fmt.Printf("%s\n", oo.JsonDataIndent(*b))
	fmt.Println("")

	return
}
func (*GenChain) OnNewTx(tx *types.Transaction, out bool) {

}
func (*GenChain) GetBDS(block *types.Block) (BDS []byte, err error) {
	return
}
func (*GenChain) GetPoa(preHash []byte, nowHeight int64) (poa []byte, err error) {
	return
}
func (*GenChain) IsReady() bool {
	return true
}
func (*GenChain) SetReady(ready bool) {
	return
}
func (*GenChain) ForkChain(b *types.Block) (err error) {
	return
}

// func (*GenChain) SwitchChain(oldc *types.TbChain, newc *types.TbChain, new_ids []string) (err error) {
// 	return
// }

// func (*GenChain) TryAcceptNewBlock(ctx *oo.ReqCtx, b *types.Block) {
// 	return
// }
func CreateGenesisBlock() (block *types.Block) {
	//construct block

	// var block types.Block
	block = &types.Block{}
	block.IndepHash = "" //--
	block.Hash = ""      //--
	// block.HashList = []string{} //--
	// block.HashListMerkle = ""   //--
	block.Height = 0
	block.PreviousBlock = ""
	block.Nonce = "0"
	block.Timestamp = time.Now().Unix()
	block.LastRetarget = block.Timestamp                                            //
	block.Diff = "fffff80000000000000000000000000000000000000000000000000000000000" //
	block.CumulativeDiff = 21                                                       //
	block.RewardAddr = "3AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"                         //
	block.RewardPool = 0                                                            //
	block.WeaveSize = 0                                                             //
	block.BlockSize = 0
	// block.Tags = []types.Tag{}
	block.WalletList = []types.WalletTuple{}
	block.WalletListHash = ""
	block.TxRoot = ""
	block.Txs = []string{}

	// var newb *types.Block
	// ch := make(chan *types.Block, 10)

	/*go */
	bds, _ := GMinerServer.Chain.GetBDS(block)

	var chain GenChain
	SetChain(&chain)
	MineRestart(block, bds)
	// for i := 1; i > 0; {
	// 	select {
	// 	case newb = <-ch:
	// 		i = 0
	// 	case <-time.After(1 * time.Second):
	// 		fmt.Printf("%d\n", i)
	// 	}
	// }
	// fmt.Printf("%#v\n", newb)
	//mine nonce （from config）

	//save block to db
	//save chain to db
	return
}
