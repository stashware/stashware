package miner

import (
	// "fmt"
	"github.com/satori/go.uuid"
	"math/big"
	"runtime"
	"stashware/oo"
	"stashware/types"
	"stashware/wallet"
	"sync/atomic"
	"time"
)

type MinerCfg = struct {
	Core int64 `toml:"core"`
	// Pubkey       string `toml:"pubkey"`
	Prikey       string `toml:"prikey"`
	Wallet       string `toml:"wallet"`
	MinerAddr    string `toml:"miner_addr"`
	NoCheckDelta int64  `toml:"nocheck_delta"`
	Debug        int64  `toml:"debug"`
	SelfMine     int64  `toml:"self_mine"`
}

type MinerServer struct {
	Conf MinerCfg
	// MinerPubkey   string
	// MinerAddr     string
	NoCheckDelta  int64
	MasterCloseCh chan struct{}
	SlaveCloseCh  chan struct{}
	WaitCh        chan int
	TaskCh        chan []byte
	enable        bool
	Chain         types.ChainIf
}

var GMinerServer = &MinerServer{}
var debug_flag int64

var hash_count uint64
var hash_rate uint64

func MinerInit(gconf *oo.Config) (err error) {
	if err = gconf.SessDecode("miner", &GMinerServer.Conf); err != nil {
		return
	}
	conf := &GMinerServer.Conf
	debug_flag = conf.Debug

	if len(conf.Wallet) > 0 {
		if prikey, err := oo.LoadFile(conf.Wallet); err == nil {
			conf.Prikey = string(prikey)
			oo.LogD("use wallet file : %s", conf.Wallet)
		}
	}
	if len(conf.Prikey) > 0 {
		w, err1 := wallet.LoadWallet(conf.Prikey)
		if err1 != nil {
			err = err1
			return
		}
		// GMinerServer.MinerPubkey = w.PublicKey()
		GMinerServer.Conf.MinerAddr = w.Address()
	}
	GMinerServer.NoCheckDelta = conf.NoCheckDelta

	GMinerServer.MasterCloseCh = make(chan struct{}, 1)
	GMinerServer.WaitCh = make(chan int, conf.Core+1)
	GMinerServer.TaskCh = make(chan []byte, 102400) //
	// GMinerServer.RetCh = make(chan *types.Block, 1024) //
	GMinerServer.SlaveCloseCh = make(chan struct{})

	core := int(conf.Core)
	if core < 1 {
		core = 1
	}
	if core > runtime.NumCPU() {
		core = runtime.NumCPU()
	}
	if core < runtime.NumCPU() {
		core += 1 // for rpc
	}

	runtime.GOMAXPROCS(core)

	GMinerServer.enable = true

	go UpdateHashRate()
	return
}

func UpdateHashRate() {
	tm := time.NewTicker(time.Duration(60) * time.Second)
	for range tm.C {
		hash_rate = atomic.SwapUint64(&hash_count, 0)
	}
}
func GetMiningInfo() (info *types.MiningInfo) {
	info = &types.MiningInfo{
		Core:     int64(GMinerServer.Conf.Core),
		HashRate: int64(hash_rate),
	}

	return
}
func SetChain(chain types.ChainIf) {
	GMinerServer.Chain = chain
}

func MinerGetAddr() (addr string) {
	// addr = GMinerServer.MinerAddr
	addr = GMinerServer.Conf.MinerAddr
	return
}

func ValidateNonce(BDS []byte, diff int64, height int64, nonce []byte) (BDSHash []byte, ok bool) {
	// {BDSHash, HashNonce, ExtraNonce, HashesTried} = bulk_hash_fast(Nonce, BDS, Diff)
	dInt := big.NewInt(0).Exp(big.NewInt(2), big.NewInt(256-diff), nil)
	hStr := oo.IntToBytes(height) //oo.Str2Bytes(fmt.Sprintf("%d", height))
	b := append(BDS, nonce...)
	b = append(b, hStr...)
	// BDSHash = oo.Sha384(b)
	dHash := big.NewInt(0).SetBytes(BDSHash)
	ok = (dHash.Cmp(dInt) <= 0)
	return
}

func CalcIndepHash(BDS []byte, BDSHash []byte, nonce []byte) (indep_hash []byte) {
	b := append(BDS, BDSHash...)
	b = append(b, nonce...)
	indep_hash = oo.Sha256(b)
	return
}

func GetSolution(block *types.Block, BDS []byte) {

	if !GMinerServer.Chain.IsReady() {
		return
	}

	tip := GMinerServer.Chain.GetTip()
	if tip == nil || tip.IndepHash == block.PreviousBlock {

		oo.LogD("get a solution: %s", oo.JsonData(block))

		if GMinerServer.Conf.Debug > 0 {
			oo.LogD("to add new block: %s", block.IndepHash)
		}
		GMinerServer.Chain.OnNewBlock(block, BDS, true)
	}
}

func MineThread(temp *types.Block, BDS []byte, iThread int) {
	key := oo.HexDecStringPad32(temp.PreviousBlock)

	vm, err := oo.NewRandxVm(key)
	if nil != err {
		oo.LogD("MineThread NewRandxVm err %v", err)
	}
	defer vm.Close()

	var (
		dInt = temp.BigDiff()
		ext2 = oo.IntToBytes(temp.Height)
	)

	valid := func(nce []byte) (ret []byte, ok bool) {
		buf := append(BDS, append(nce, ext2...)...)
		ret = vm.Hash(buf)

		ok = big.NewInt(0).SetBytes(ret).Cmp(dInt) > 0

		atomic.AddUint64(&hash_count, 1)
		return
	}

	for {
		select {
		case nonce := <-GMinerServer.TaskCh:
			// if hash, ok := ValidateNonce(BDS, temp.Diff, temp.Height, nonce); ok {
			if hash, ok := valid(nonce); ok {
				retBlock := *temp
				retBlock.Nonce = oo.HexEncodeToString(nonce)
				retBlock.Hash = oo.HexEncToStringPad32(hash)
				retBlock.IndepHash = oo.HexEncToStringPad32(CalcIndepHash(BDS, hash, nonce))
				// notifyCh <- retBlock
				GetSolution(&retBlock, BDS)
				return
			}
		case <-GMinerServer.SlaveCloseCh:
			return
		}
	}
}

func MineGenNonce() {
	nMax := big.NewInt(2)
	nMax.Exp(nMax, big.NewInt(256), nil)

	n := big.NewInt(0).SetBytes(uuid.NewV4().Bytes())
	for ; n.Cmp(nMax) <= 0; n.Add(n, big.NewInt(1)) {
		select {
		case GMinerServer.TaskCh <- n.Bytes():
			//continue
		case <-GMinerServer.SlaveCloseCh:
			return
			// case time.After(100 * time.Millisecond):
			//donothing
		}
	}
}

func MineStop() {
	close(GMinerServer.SlaveCloseCh)
	GMinerServer.MasterCloseCh <- struct{}{}
	// close(GMinerServer.MasterCloseCh)
}

func MineRestart(newb *types.Block, BDS []byte) (err error) {
	if !GMinerServer.enable || len(BDS) == 0 {
		err = oo.NewError("miner mod not start.")
		return
	}
	oo.LogD("mining %d start ...", newb.Height)

	defer func() {}()
	//stop first
	select {
	case <-GMinerServer.SlaveCloseCh:
	default:
		close(GMinerServer.SlaveCloseCh)
	}

	GMinerServer.MasterCloseCh <- struct{}{}
	GMinerServer.SlaveCloseCh = make(chan struct{})

	go func() {
		if GMinerServer.Conf.Debug > 0 {
			oo.LogD("thread %d start", 0)
		}
		MineGenNonce()
		GMinerServer.WaitCh <- 0
	}()

	for i := 0; i < int(GMinerServer.Conf.Core); i++ {
		go func(i int) {
			if GMinerServer.Conf.Debug > 0 {
				oo.LogD("thread %d start", i+1)
			}
			MineThread(newb, BDS, i+1)
			GMinerServer.WaitCh <- (i + 1)
		}(i)
	}

	//wait solution / wait stop
	for i := 0; i < int(GMinerServer.Conf.Core)+1; i++ {
		n := <-GMinerServer.WaitCh
		if GMinerServer.Conf.Debug > 0 {
			oo.LogD("thread %d stop", n)
		}
		if i == 0 {
			close(GMinerServer.SlaveCloseCh)
		}
	}

	//discard all task and ret
	for i := 0; i == 0; {
		select {
		case <-GMinerServer.TaskCh:
		default:
			i = 1
		}
	}

	<-GMinerServer.MasterCloseCh

	oo.LogD("mining %d stop ...", newb.Height)

	return
}
