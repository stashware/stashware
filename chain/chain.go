package chain

import (
	// "os"
	// "encoding/binary"
	// "math/big"
	// "sort"
	"stashware/miner"
	"stashware/oo"
	"stashware/storage"
	"stashware/txpool"
	"stashware/types"
	// "stashware/wallet"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var (
	GChain = &BlockChain{}

	GenesisIndepHash string
	new_block_mutex  sync.Mutex
)

type ChainCfg = struct {
	GenesisNonce       string   `toml:"genesis_nonce"`
	RelayTxcostPerByte int64    `toml:"relay_txcost_per_byte"`
	Debug              int64    `toml:"debug"`
	Genesis            string   `toml:"genesis"`
	CheckPoints        []string `toml:"check_points"`
}

type BlockChain struct {
	tipBlock  *types.Block
	Chain     *types.TbChain
	conf      ChainCfg
	Gossip    types.GossipIf
	syncing   int64
	do_repair bool
}

// type ChainNotifier interface {
// 	TipChange(b *types.Block)
// }

// var ChainNotifiers = []ChainNotifier{}

// func ChainAddNotifier(cn ChainNotifier) {
// 	ChainNotifiers = append(ChainNotifiers, cn)
// }

func ChainInit(gconf *oo.Config) (err error) {
	// GChain = &BlockChain{}
	if err = gconf.SessDecode("chain", &GChain.conf); err != nil {
		return
	}
	conf := &GChain.conf
	for _, one := range conf.CheckPoints {
		cps := strings.Split(one, ":")
		if len(cps) == 2 {
			height := oo.Str2Int(cps[0])
			hash := cps[1]

			if hash[:1] == "-" {
				types.AcceptCPs[height] = types.CheckPoint{
					Hash:   hash[1:],
					Reject: true,
				}
			} else {
				types.AcceptCPs[height] = types.CheckPoint{
					Hash:   hash,
					Reject: false,
				}
			}
		}
	}
	oo.LogD("check points %v", types.AcceptCPs)
	return
}
func SetGossip(gossip types.GossipIf) {
	GChain.Gossip = gossip
}

func InitGenesisBlock() {

	var block types.Block
	genesis_data, err := oo.LoadFile(GChain.conf.Genesis)
	if err != nil {
		oo.LogW("Failed to load genesis file :%s, err %v", GChain.conf.Genesis, err)
		return
	}
	if err = oo.JsonUnmarshal(genesis_data, &block); err != nil {
		oo.LogW("Failed to parse genesis file :%s, err %v", GChain.conf.Genesis, err)
		return
	}
	// var err error
	tb := block.TbBlock()
	tb.Network = int64(1)
	if err = storage.GStore.WriteBlock(tb); err == nil {
		// oo.LogD("Failed to write genesis block. err=%v", err)
		// return
	}
	ChainSaveInit(&block)

	oo.LogD("CREATE GENESIS BLOCK %s : %v", block.IndepHash, block)
}

func ChainSaveInit(b *types.Block) {
	if GChain.tipBlock != nil && GChain.tipBlock.IndepHash == b.IndepHash {
		return
	}

	//update wallet
	for _, w := range b.WalletList {
		if err := storage.GStore.UpdateWallet(w.Wallet, w.LastTx, w.Quantity); err != nil {
			oo.LogW("Failed to update wallet. %v, err:%v", w, err)
		}
	}

	GChain.StepTip(b)
}

func ChainPostInit() (err error) {
	//load data
	tb, err := storage.GStore.ReadBlockByHeight(0)
	if err != nil {
		oo.LogD("Create genesis block")
		InitGenesisBlock()

		tb, err = storage.GStore.ReadBlockByHeight(0)
		if nil != err {
			oo.LogD("ChainPostInit ReadBlockByHeight %v", err)
			return
		}
	} else {
		oo.LogD("GENESIS %s", tb.IndepHash)
	}

	if nil == tb {
		err = oo.NewError("nil == tb")
		oo.LogD("ChainPostInit err %v", err)
		return
	}
	if tb.IndepHash != "24bf2b4086c35b2977cec27d375e6638df31c26d8ec6a020bd22c7fadfcec101" {
		//not mainnet
		target_time = target_time / 50
	}
	GenesisIndepHash = tb.IndepHash

	// TryRepairDb()

	GChain.Chain, GChain.tipBlock, err = GChain.GetActiveChain()
	if err != nil || GChain.tipBlock == nil {
		oo.LogD("Not found active chain. err %v", err)
		return
	}

	oo.LogD("Read chain: %d %s", GChain.tipBlock.Height, GChain.tipBlock.IndepHash)

	if miner.GMinerServer.Conf.SelfMine == 0 {
		GChain.SetReady(false) //until one peer.
	}
	go ChainStartMine()

	return
}

func (this *BlockChain) GetActiveChain() (cs *types.TbChain, b *types.Block, err error) {
	css, err := storage.GStore.ReadChainTips()
	if err != nil {
		oo.LogW("Failed to read chain info, err=%v", err)
		return
	}
	for _, c1 := range css {
		if c1.IsActive == int64(1) {
			var tb *types.TbBlock
			tb, err = storage.GStore.ReadBlockByHash(c1.IndepHash)
			if err != nil {
				oo.LogD("failed to load tip hash: %s, err=%v", c1.IndepHash, err)
				return
			}
			b = tb.Block()
			cs = &c1
			return
		}
	}

	err = oo.NewError("not found active chain")
	return
}

func ChainCreateBlockTemplate(tip *types.Block) (tblock *types.Block) {
	tblock = &types.Block{}
	tblock.PreviousBlock = tip.IndepHash
	tblock.RewardAddr = miner.MinerGetAddr()

	tblock.Timestamp = time.Now().Unix()
	if tblock.Timestamp <= tip.Timestamp {
		tblock.Timestamp = tip.Timestamp + 1
	}

	tblock.Height = tip.Height + 1
	tblock.Diff = oo.HexEncToStringPad32(CalcNextRequiredDifficulty(tip).Bytes())

	//
	// after setting-diff
	//
	tblock.CumulativeDiff = CalcNextCumulativeDiff(tip.CumulativeDiff, tblock.BigDiff())
	//
	// after setting-timestamp
	//
	tblock.LastRetarget = CalcLastRetarget(tblock, tip.LastRetarget)

	txs := txpool.GetForMine()
	for _, tx := range txs {
		tblock.Txs = append(tblock.Txs, tx.ID)
		tblock.BlockSize += int64(len(tx.RawData()))
	}
	tblock.WeaveSize = tip.WeaveSize + tblock.BlockSize

	tblock.TxRoot = CalcTxRootString(tblock)

	var err error
	if tblock.WalletList, err = CalcWalletList(tblock, txs); err != nil {
		oo.LogW("logic failed. err=%v", err)
		return
	}

	miner_reward, new_endowpool, err := CalcRewardPool(tblock, tip.RewardPool)
	if err != nil {
		oo.LogW("logic failed. err=%v txs %v", err, tblock.Txs)
		return
	}
	tblock.RewardPool = new_endowpool
	tblock.RewardFee = miner_reward

	for i, w := range tblock.WalletList {
		if w.Wallet == tblock.RewardAddr {
			tblock.WalletList[i].Quantity += miner_reward
		}
	}
	//
	// after setting-wallet_list
	//
	tblock.WalletListHash = oo.HexEncToStringPad32(tblock.HashWalletList())

	oo.LogD("template: %s +%d, endow +%d, wallet: %v", tblock.RewardAddr, miner_reward, new_endowpool, tblock.WalletList)
	if GChain.conf.Debug == int64(1) {
		oo.LogD("template prev [%d %s, diff %x, lastretarget %d, ts %d", tip.Height, tip.IndepHash,
			tip.BigDiff().Bytes(), tip.LastRetarget, tip.Timestamp)
	}
	return
}

func ChainStartMine() {

	addr := miner.MinerGetAddr()
	if len(addr) == 0 {
		oo.LogD("no miner addr")
		return
	}
	oo.LogD("miner addr : %s", addr)

	tiph := "" //GChain.tipBlock.IndepHash
	for {
		this_tip := *GChain.tipBlock
		if GChain.IsReady() && tiph != this_tip.IndepHash {

			tip := &this_tip
			tiph = tip.IndepHash

			need_swr, err := GetMinerPledgeNeed(tip.Height, addr)
			if err != nil {
				oo.LogD("Failed to get Miner need. err=%v", err)
				continue
			}

			if need_swr > 0 {
				w, err := storage.GStore.ReadWalletByAddress(addr)
				if err != nil {
					oo.LogD("h %d, addr not active: %s", tip.Height, addr)
					continue
				}
				if w.Balance < need_swr*types.MINTOKEN_PER_TOKEN {
					oo.LogD("h %d, balance %d < need %d, skip mined", tip.Height, w.Balance/types.MINTOKEN_PER_TOKEN, need_swr)
					continue
				}
			}

			newb := ChainCreateBlockTemplate(tip)

			poa, err := GChain.GetPoa(oo.HexDecStringPad32(tiph), tip.Height+1)
			if err != nil {
				oo.LogD("Failed to get h=%d poa. err %v", tip.Height, err)
				continue
			}
			if len(poa) == 0 {
				oo.LogW("POA null !!! preh=%d", tip.Height)
			}
			if len(poa) > 0 {
				newb.Poa = oo.HexEncodeToString(poa)
			}

			BDS, _ := GChain.GetBDS(newb)
			miner.MineRestart(newb, BDS)
		} else {
			//<-time.After(1 * time.Second)
			time.Sleep(1 * time.Second)
		}
	}

}

func ChainTipHash() (block_id string) {

	if GChain.tipBlock != nil {
		block_id = GChain.tipBlock.IndepHash
	}
	return
}

func TipBlock() (block *types.Block) {
	return GChain.tipBlock
}

func (this *BlockChain) GetTip() (b *types.Block) {
	b = GChain.tipBlock
	return
}
func (this *BlockChain) IsReady() bool {

	syncing := atomic.LoadInt64(&this.syncing)
	// oo.LogD("Sying :%v", syncing)
	return syncing == 0
}
func (this *BlockChain) SetReady(ready bool) {

	syncing := int64(1)
	if ready {
		syncing = 0
	}
	// oo.LogD("SET Sying :%v", syncing)
	atomic.StoreInt64(&this.syncing, syncing)
}

//must save block before here
func (this *BlockChain) StepTip(b *types.Block) (err error) {
	tb := b.TbBlock()
	tb.Network = types.NETWORK_MAINNET
	if err = storage.GStore.UpdateBlockFlag(tb); err != nil {
		return
	}

	css := []types.TbChain{
		types.TbChain{
			Height:    tb.Height,
			IndepHash: tb.IndepHash,
			IsActive:  int64(1),
		},
	}
	if err = storage.GStore.UpdateChainTips(css); err != nil {
		//oo.LogW("Failed to update chaintips. err %v", err)
		return
	}

	cs, tip, err := GChain.GetActiveChain()
	if err != nil {
		return
	}
	GChain.Chain, GChain.tipBlock = cs, tip

	if !GChain.do_repair {
		oo.LogD("Change TIP %d %s", cs.Height, cs.IndepHash)
	}
	return
}

func (this *BlockChain) OnNewBlock(b *types.Block, BDS []byte, out bool) (err error) {

	var in_white_list bool
	var failed bool
	if in_white_list, failed = types.CheckCps(b.Height, b.IndepHash); failed {
		return
	}
	new_block_mutex.Lock()
	defer new_block_mutex.Unlock()

	if b.PreviousBlock != TipBlock().IndepHash {
		oo.LogD("tip has change. cancel add block. %d %s", b.Height, b.IndepHash)
		return
	}

	//1. verify
	verify := VerifyBlock(b)
	if !verify {
		if !in_white_list {
			err = types.BLK_VERIFY_ERROR
			oo.LogD("OnNewBlock VerifyBlock err %s", oo.JsonData(b))
			return
		}
		oo.LogD("verify failed, but white list %s", oo.JsonData(b))
	}
	tb := b.TbBlock()
	tb.Network = types.NETWORK_MAINNET

	//2. save it
	err = storage.GStore.WriteBlock(tb)
	if err != nil {
		// oo.LogD("Block exists %d %s", b.Height, b.IndepHash)
		// return
	}

	err = storage.GStore.UpdateTxsRefBlock(b.Txs, b.IndepHash)
	if err != nil {
		oo.LogD("OnNewBlock UpdateTxsRefBlock err %v", err)
		return
	}
	for _, w := range b.WalletList {
		if err = storage.GStore.UpdateWallet(w.Wallet, w.LastTx, w.Quantity); err != nil {
			oo.LogW("Failed to update wallet. %v, err:%v", w, err)
			return
		}
	}
	if GChain.StepTip(b) != nil {
		oo.LogW("Failed to update tipBlock %d %s, err %v", b.Height, b.IndepHash, err)
		return
	}

	//remove txpool txs
	txpool.RemoveTxs(b.Txs)

	if out {
		b.Poa = ""
		BDS = []byte("")

		//send to friend
		GChain.Gossip.SendNewBlock(b, BDS)
	}

	// oo.LogD("succeed to change CHAIN TIP %d %s", b.Height, b.IndepHash)
	// }

	return
}

func (this *BlockChain) OnNewTx(tx *types.Transaction, out bool) {
	//try to save
	if len(tx.Data) > 0 {
		if err := storage.GStore.WriteTxData(tx.ID, tx.RawData()); err != nil {
			oo.LogD("OnNewTx WriteTxData(%s) err %v", tx.ID, err)
			return
		}
	}

	tbtx := *(tx.TbTransaction())
	err := storage.GStore.WriteTxs([]types.TbTransaction{tbtx})
	if err != nil {
		oo.LogD("OnNewTx WriteTxs(%s) err %v", tbtx.ID, err)
		return
	}

	// if not exists , try to add to mempool
	txpool.AddTx(tx)

	if out {
		// try to send to friend
		GChain.Gossip.SendNewTx(tx)
	}
}

func GetBlock(blk_hash string) (ret *types.Block, err error) {
	if nil == storage.GStore {
		err = oo.NewError("storer is nil.")
		return
	}

	block, err := storage.GStore.ReadBlockByHash(blk_hash)
	if nil != err {
		err = oo.NewError("ReadBlockByHash(%s) %v", blk_hash, err)
		return
	}
	ret = block.Block()

	return
}

//
// should return the main-chain block of height
//
func GetHeightBlock(height int64) (ret *types.Block, err error) {
	if nil == storage.GStore {
		err = oo.NewError("storer is nil.")
		return
	}

	tb, err := storage.GStore.ReadBlockByHeight(height)
	if nil != err {
		err = oo.NewError("ReadBlockByHeight(%d) %v", height, err)
		return
	}
	ret = tb.Block()

	return
}

func GetTransaction(tx_id string) (ret *types.Transaction, err error) {
	if nil == storage.GStore {
		err = oo.NewError("storer is nil.")
		return
	}

	tx, err := storage.GStore.ReadTx(tx_id)
	if nil != err {
		err = oo.NewError("ReadTx(%s) %v", tx_id, err)
		return
	}
	ret = tx.Transaction()

	return
}

func GetTxData(tx_id string) (data []byte, err error) {
	if nil == storage.GStore {
		err = oo.NewError("storer is nil.")
		return
	}

	data, err = storage.GStore.ReadTxData(tx_id)
	if nil != err {
		err = oo.NewError("ReadTxData(%s) %v", tx_id, err)
		return
	}

	return
}

func GetFullTx(tx_id string) (ret *types.FullTx, err error) {
	if nil == storage.GStore {
		err = oo.NewError("storer is nil.")
		return
	}

	ret, err = storage.GStore.ReadFullTx(tx_id)
	if nil != err {
		return
	}

	return
}

func GetPendingAmount(addr string) (in, out int64, err error) {
	txs, err := storage.GStore.ReadPendingTxs(addr)
	if nil != err {
		err = oo.NewError("ReadPendingTxs(%s) %v", addr, err)
		return
	}
	for _, val := range txs {
		if val.Target == addr {
			in += val.Quantity //
		}
		if val.FromAddress == addr {
			out += val.Quantity //
		}
	}

	return
}

func (*BlockChain) GetBDS(block *types.Block) (BDS []byte, err error) {
	return block.DataSegment(), nil
}
