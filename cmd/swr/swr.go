package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"stashware/chain"
	"stashware/miner"
	"stashware/node"
	"stashware/oo"
	"stashware/rest"
	"stashware/storage"
	"stashware/txpool"
	"stashware/types"
	"stashware/web_wallet"

	"strings"
)

var GServerName string
var GitVersion string

// -vVdD -c cfgfile -l logpath
type CmdArgs = struct {
	CfgFiles  []string
	LogPath   string
	Debug     int64
	Daemon    bool
	Version   bool
	Help      bool
	Mine      int64
	Genesis   bool
	TrimBlock int64
	Rebuild   bool

	Parsed bool
}

var GCmdArgs CmdArgs
var GConfig *oo.Config
var GStore *storage.StoreActor

func ParseCmdArgs(ver string) (ca *CmdArgs) {
	ca = &GCmdArgs
	if ca.Parsed {
		return
	}

	cfgfile := string("")
	flag.StringVar(&cfgfile, "c", "./swr.conf", "config file path")
	flag.BoolVar(&ca.Daemon, "D", false, "deamon application")
	flag.StringVar(&ca.LogPath, "l", "./", "log files path")
	flag.BoolVar(&ca.Version, "v", false, "print version")
	flag.BoolVar(&ca.Version, "V", false, "print version")
	flag.Int64Var(&ca.Mine, "m", 0, "mine start flag")
	flag.BoolVar(&ca.Genesis, "g", false, "genesis block")
	flag.BoolVar(&ca.Rebuild, "r", false, "rebuild db")
	flag.Int64Var(&ca.Debug, "d", 0, "debug")
	flag.Int64Var(&ca.TrimBlock, "t", -1, "trim block to height")
	flag.BoolVar(&ca.Help, "h", false, "help")
	flag.Parse()

	if ca.Help || ca.Version {
		fmt.Printf("Version: %s\n", ver)
		if ca.Help {
			flag.PrintDefaults()
		}
		os.Exit(0)
	}

	if ca.Daemon {
		ca.Daemon = false
		args := flag.Args()
		cmd := exec.Command(os.Args[0], args...)
		cmd.Start()
		fmt.Printf("[PID]%v\n", cmd.Process.Pid)
		os.Exit(0)
	}

	ca.CfgFiles = strings.Split(cfgfile, ",")
	ca.Parsed = true

	return
}

type Mod = struct {
	name        string
	initfn      func(*oo.Config) error
	post_initfn func() error
}

var AllMod = []Mod{
	{"storage", storage.StorageInit, nil},
	{"chain", chain.ChainInit, chain.ChainPostInit},
	{"node", node.NodeInit, node.NodePostInit},
	{"rest", rest.RestInit, rest.RestPostInit},
	{"mempool", txpool.TxpoolInit, nil},
	{"web_wallet", web_wallet.WebWalletInit, web_wallet.WebWalletPostInit},
	// {"miner", miner.MinerInit, nil},
}

func main() {
	var err error
	ParseCmdArgs(GitVersion)
	GServerName = strings.Split(filepath.Base(os.Args[0]), ".")[0]

	err = os.Chdir(filepath.Dir(os.Args[0]))
	if nil != err {
		fmt.Printf("chdir err %v", err)
		return
	}

	//default 4 core
	oo.GoRunProc(4)

	if GConfig, err = oo.InitConfig(GCmdArgs.CfgFiles, nil); err != nil {
		fmt.Printf("Failed to parse config. err=%s\n", err)
		return
	}
	var global_cfg types.GlobalCfg
	if err = GConfig.SessDecode("global", &global_cfg); err != nil {
		fmt.Printf("Config file error. err=%s\n", err)
		return
	}
	if GCmdArgs.Debug == 0 {
		GCmdArgs.Debug = global_cfg.Debug
	}

	logdir := filepath.Join(global_cfg.DataDir, "logs")
	os.MkdirAll(logdir, os.ModePerm)

	for _, dir1 := range global_cfg.TxsDirs {
		os.MkdirAll(dir1, os.ModePerm)
	}

	if GCmdArgs.Debug == 0 {
		oo.InitLog(logdir, GServerName, GitVersion, nil, 60)
	}

	//common mod
	if GCmdArgs.Mine > 0 {
		AllMod = append(AllMod, Mod{"miner", miner.MinerInit, nil})
	}

	for _, mod := range AllMod {
		oo.LogD("init %s", mod.name)
		if err = mod.initfn(GConfig); err != nil {
			oo.LogW("Failed to init %s, err=%v", mod.name, err)
			return
		}
	}

	//link interface
	chain.SetGossip(node.GNodeServer)
	miner.SetChain(chain.GChain)
	node.SetChain(chain.GChain)
	node.SetVersion(GitVersion)

	if GCmdArgs.Genesis {
		fmt.Println("Begin genesis...")
		miner.CreateGenesisBlock()
		fmt.Println("End genesis")
		os.Exit(0)
	}

	if GCmdArgs.TrimBlock >= 0 {
		if err = chain.TrimBlock(GCmdArgs.TrimBlock); err != nil {
			oo.LogD("Failed to trimblock, err %v", err)
			return
		}
	}
	if GCmdArgs.Rebuild {
		if err = chain.RebuildDb(); err != nil {
			oo.LogW("Failed to rebuild db. err %v", err)
			return
		}
	}

	for _, mod := range AllMod {
		if mod.post_initfn == nil {
			continue
		}
		if err = mod.post_initfn(); err != nil {
			oo.LogW("Failed to post init %s, err=%v", mod.name, err)
			return
		}
	}

	fmt.Printf("start stashware %s.\n", GitVersion)

	oo.LogD("start stashware %s.", GitVersion)

	//
	oo.WaitExitSignal()
}
