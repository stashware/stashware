package storage

import (
	"fmt"
	"path/filepath"
	"stashware/oo"
	"stashware/types"
)

var GStore *StoreActor

func StorageInit(gconf *oo.Config) (err error) {
	//decode global_config
	var global_cfg types.GlobalCfg
	if err = gconf.SessDecode("global", &global_cfg); err != nil {
		fmt.Printf("Config file error. err=%s\n", err)
		return
	}
	dbpath := filepath.Join(global_cfg.DataDir, "stashware.db")
	GStore, err = NewStoreActor(dbpath, global_cfg.TxsDirs)
	if err != nil {
		oo.LogW("Failed to loaddb. err=%v", err)
		return
	}
	return
}
