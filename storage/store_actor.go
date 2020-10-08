package storage

import (
	"fmt"

	"stashware/oo"
	"stashware/types"
)

type StoreActor struct {
	db_path   string
	data_path []string
}

func NewStoreActor(db_path string, data_path []string) (ret *StoreActor, err error) {
	err = initSqlx(db_path)
	if nil != err {
		err = oo.NewError("initSqlx(%s) %v", db_path, err)
		return
	}

	err2 := initTables()
	if nil != err {
		oo.LogD("NewStoreActor initTables err %v", err2)
	}

	ret = &StoreActor{
		db_path:   db_path,
		data_path: data_path,
	}

	return
}

func (this *StoreActor) WriteBlock(block *types.TbBlock) (err error) {
	sqlstr := oo.NewSqler().Table("tb_blocks").
		Insert(oo.Struct2Map(*block))
	err = DbExec(sqlstr)
	if nil != err {
		err = oo.NewError("DbExec(%s) %v", sqlstr, err)
		return
	}

	return
}
func (this *StoreActor) UpdateBlockFlag(block *types.TbBlock) (err error) {
	sqlstr := oo.NewSqler().Table("tb_blocks").
		Where("indep_hash", block.IndepHash).
		Update(map[string]interface{}{
			"network": block.Network,
		})
	err = DbExec(sqlstr)
	if nil != err {
		err = oo.NewError("DbExec(%s) %v", sqlstr, err)
		return
	}

	return
}

func (this *StoreActor) ReadBlockArrByHeight(min_h int64, max_h int64) (blocks []types.TbBlock, err error) {
	var block types.TbBlock

	sqlstr := oo.NewSqler().Table("tb_blocks").
		Where("height", ">=", min_h).
		Where("height", "<=", max_h).
		Order("height asc").
		Select(oo.Struct2Fields(block))
	err = DbSelect(sqlstr, &blocks)
	if nil != err {
		err = oo.NewError("DbSelect(%s) %v", sqlstr, err)
		return
	}

	return
}

func (this *StoreActor) ReadBlockByHeight(height int64) (block *types.TbBlock, err error) {
	var blocks []types.TbBlock
	var b types.TbBlock

	sqlstr := oo.NewSqler().Table("tb_blocks").
		Where("height", height).
		Where("network", int64(1)).
		Select(oo.Struct2Fields(b))
	err = DbSelect(sqlstr, &blocks)
	if nil != err {
		err = oo.NewError("DbSelect(%s) %v", sqlstr, err)
		return
	}
	if len(blocks) != 1 {
		err = oo.NewError("GetBlockByHeight %d returns %d count", height, len(blocks))
		return
	}
	block = &blocks[0]

	return
}

func (this *StoreActor) ReadBlockByHash(block_hash string) (block *types.TbBlock, err error) {
	block = &types.TbBlock{}

	sqlstr := oo.NewSqler().Table("tb_blocks").
		Where("indep_hash", block_hash).
		Select(oo.Struct2Fields(*block))
	err = DbGet(sqlstr, block)
	if nil != err {
		err = oo.NewError("DbGet(%s) %v", sqlstr, err)
		return
	}

	return
}

func (this *StoreActor) WriteTxs(txs []types.TbTransaction) (err error) {
	if len(txs) < 1 {
		return
	}

	var data []map[string]interface{}
	for _, tx := range txs {
		data = append(data, oo.Struct2Map(tx))
	}

	sqlstr := oo.NewSqler().Table("tb_transactions").
		InsertBatch(data)
	err = DbExec(sqlstr)
	if nil != err {
		err = oo.NewError("DbExec(%s) %v", sqlstr, err)
		return
	}

	return
}

func (this *StoreActor) UpdateTxsRefBlock(tx_ids []string, block_hash string) (err error) {
	if len(tx_ids) < 1 {
		return
	}

	var data []map[string]interface{}
	for _, tx_id := range tx_ids {
		data = append(data, map[string]interface{}{
			"id":               tx_id,
			"block_indep_hash": block_hash,
		})
	}

	sqlstr := oo.NewSqler().Table("tb_transactions").
		UpdateBatch(data, []string{"id"})
	err = DbExec(sqlstr)
	if nil != err {
		err = oo.NewError("DbExec(%s) %v", sqlstr, err)
		return
	}

	return
}

func (this *StoreActor) ReadTx(tx_hash string) (tx *types.TbTransaction, err error) {
	tx = &types.TbTransaction{}

	sqlstr := oo.NewSqler().Table("tb_transactions").
		Where("id", tx_hash).
		Select(oo.Struct2Fields(*tx))
	err = DbGet(sqlstr, tx)
	if nil != err {
		if ErrNoRows == err {
			return
		}
		err = oo.NewError("DbGet(%s) %v", sqlstr, err)
		return
	}

	return
}

func (this *StoreActor) ReadTxsByBlockHash(block_hash string) (txs []types.TbTransaction, err error) {
	tx := &types.TbTransaction{}

	sqlstr := oo.NewSqler().Table("tb_transactions").
		Where("block_indep_hash", block_hash).
		Select(oo.Struct2Fields(*tx))
	err = DbSelect(sqlstr, &txs)
	if nil != err {
		err = oo.NewError("DbSelect(%s) %v", sqlstr, err)
		return
	}

	return
}

func (this *StoreActor) ReadTxsByAddress(addr string, h1, h2 *int64) (ret []types.TbTransaction, err error) {
	tx := &types.TbTransaction{}

	sqler := oo.NewSqler().Table("tb_transactions a").
		LeftJoin("tb_blocks b", "a.block_indep_hash=b.indep_hash")
	if nil != h1 {
		sqler.Where("b.height", ">=", *h1)
	}
	if nil != h2 {
		sqler.Where("b.height", "<=", *h2)
	}
	if "" != addr {
		sqler.Where(fmt.Sprintf(`(a.from_address="%s" or a.target="%s")`, addr, addr))
	}

	sqlstr := sqler.Order("ctime DESC").
		Select(oo.Struct2Fields(*tx))

	err = DbSelect(sqlstr, &ret)
	if nil != err {
		err = oo.NewError("DbSelect(%s) %v", sqlstr, err)
		return
	}

	return
}

func (this *StoreActor) WriteTxData(tx_id string, data []byte) (err error) {
	var dir string
	for _, val := range this.data_path {
		free := DiskFreeBytes(val)
		if free > int64(len(data)) {
			dir = val
			break
		}
	}

	if "" == dir {
		err = oo.NewError("no available disk space")
		return
	}

	filename := PathJoin(dir, tx_id)
	err = WriteFile(filename, data)
	if nil != err {
		err = oo.NewError("WriteFile(%s) %v", filename, err)
		return
	}

	return
}

func (this *StoreActor) ReadTxData(tx_id string) (data []byte, err error) {
	var filename string
	for _, val := range this.data_path {
		tmp := PathJoin(val, tx_id)
		exists, _ := oo.FileExists(tmp)
		if exists {
			filename = tmp
		}
	}

	if "" == filename {
		err = oo.NewError("no such file %s", tx_id)
		return
	}

	data, err = ReadFile(filename)
	if nil != err {
		err = oo.NewError("ReadFile(%s) %v", filename, err)
		return
	}

	return
}

func (this *StoreActor) UpdateWallet(address, last_tx string, balance int64) (err error) {
	if "" == address {
		err = oo.NewError("address is empty")
		return
	}
	// if "" == last_tx {
	// 	err = oo.NewError("last_tx should not be empty")
	// 	return
	// }
	if balance < 0 {
		err = oo.NewError("balance < 0")
		return
	}
	sqler := oo.NewSqler().Table("tb_wallet").
		Where("address", address)

	sqler2 := *sqler

	var wallet types.TbWallet
	sqlstr := sqler.Select(oo.Struct2Fields(wallet))
	err = DbGet(sqlstr, &wallet)
	if nil != err {
		if ErrNoRows != err {
			err = oo.NewError("DbGet(%s) %v", sqlstr, err)
			return
		}
		// add new
		data := map[string]interface{}{
			"address": address,
			"last_tx": last_tx,
			"balance": balance,
		}
		sqlstr := oo.NewSqler().Table("tb_wallet").
			Insert(data)
		err = DbExec(sqlstr)
		if nil != err {
			err = oo.NewError("DbExec(%s) %v", sqlstr, err)
			return
		}
		// oo.LogD("sql: %s |||| %d", sqlstr, balance)
		return
	}

	var (
		data = make(map[string]interface{})
	)
	if last_tx != wallet.LastTx {
		data["last_tx"] = last_tx
	}
	if balance != wallet.Balance {
		data["balance"] = balance
	}

	if len(data) > 0 {
		sqlstr := sqler2.Update(data)
		err = DbExec(sqlstr)
		if nil != err {
			err = oo.NewError("DbExec(%s) %v", sqlstr, err)
			return
		}
	}

	return
}

func (this *StoreActor) ReadWalletList() (ret []types.TbWallet, err error) {
	var wallet types.TbWallet

	sqlstr := oo.NewSqler().Table("tb_wallet").
		Select(oo.Struct2Fields(wallet))
	err = DbSelect(sqlstr, &ret)
	if nil != err {
		err = oo.NewError("DbSelect(%s) %v", sqlstr, err)
		return
	}

	return
}

func (this *StoreActor) ReadWalletByAddress(address string) (wallet *types.TbWallet, err error) {
	wallet = &types.TbWallet{}

	sqlstr := oo.NewSqler().Table("tb_wallet").
		Where("address", address).
		Select(oo.Struct2Fields(*wallet))
	err = DbGet(sqlstr, wallet)
	if nil != err {
		if ErrNoRows == err {
			return
		}
		err = oo.NewError("DbGet(%s) %v", sqlstr, err)
		return
	}

	return
}

func (this *StoreActor) ReadLastTxByAddress(address, tx_id string) (last_tx string, err error) {
	sqler := oo.NewSqler().Table("tb_transactions").
		Where("from_address", address)
	if "latest" != tx_id {
		var (
			row    int64
			sqlstr = fmt.Sprintf(`select row from tb_transactions where id="%s"`, tx_id)
		)
		err = DbGet(sqlstr, &row)
		if nil != err {
			if ErrNoRows != err {
				err = oo.NewError("DbGet(%s) %v", sqlstr, err)
				return
			}
		} else {
			sqler.Where(fmt.Sprintf(`row<%d`, row))
		}
	}
	sqlstr := sqler.
		Order("row DESC").
		Limit(1).
		Select(`COALESCE(id,"")`)
	err = DbGet(sqlstr, &last_tx)
	if nil != err {
		if ErrNoRows == err {
			err = nil
			return
		}
		err = oo.NewError("DbGet(%s) %v", sqlstr, err)
		return
	}

	return
}

func (this *StoreActor) ReadPendingTxs(addr string) (txs []types.TbTransaction, err error) {
	sqler := oo.NewSqler().Table("tb_pool")
	if "" != addr {
		sqler.Where(fmt.Sprintf(`(from_address="%s" or target="%s")`, addr, addr))
	}

	var (
		tx_ids []string
		tx     types.TbTransaction
	)

	sqlstr := sqler.Select("tx_id")
	err = DbSelect(sqlstr, &tx_ids)
	if nil != err {
		err = oo.NewError("DbSelect(%s) %v", sqlstr, err)
		return
	}

	if len(tx_ids) > 0 {
		sqlstr = oo.NewSqler().Table("tb_transactions").
			Where("id", "in", tx_ids).
			Select(oo.Struct2Fields(tx))
		err = DbSelect(sqlstr, &txs)
		if nil != err {
			err = oo.NewError("DbSelect(%s) %v", sqlstr, err)
			return
		}
	}

	return
}

func (this *StoreActor) ReadChainTips() (ret []types.TbChain, err error) {
	var chain types.TbChain

	sqlstr := oo.NewSqler().Table("tb_chain").
		Select(oo.Struct2Fields(chain))
	err = DbSelect(sqlstr, &ret)
	if nil != err {
		err = oo.NewError("DbSelect(%s) %v", sqlstr, err)
		return
	}

	return
}
func (this *StoreActor) ReadActiveBlocks(min_h int64, max_h int64) (ret []types.ActiveBlock, err error) {
	sqlstr := oo.NewSqler().Table("tb_blocks").
		Where("height", ">=", min_h).
		Where("height", "<=", max_h).
		Where("network", 1).
		Order("height asc").
		Select("height, indep_hash")
	err = DbSelect(sqlstr, &ret)
	if nil != err {
		err = oo.NewError("DbSelect(%s) %v", sqlstr, err)
		return
	}
	return
}
func (this *StoreActor) ClearTips() (err error) {
	//remove all tips
	sqlstr := "DELETE FROM tb_wallet"
	err = DbExec(sqlstr)
	if err != nil {
		err = oo.NewError("DbExec(%s) %v", sqlstr, err)
		return
	}

	sqlstr = "DELETE FROM tb_blocks where network=0"
	if err = DbExec(sqlstr); err != nil {
		err = oo.NewError("DbExec(%s) %v", sqlstr, err)
		return
	}
	sqlstr = "DELETE FROM tb_transactions where block_indep_hash=''"
	if err = DbExec(sqlstr); err != nil {
		err = oo.NewError("DbExec(%s) %v", sqlstr, err)
		return
	}

	//update
	sqlstr = oo.NewSqler().Table("tb_blocks").
		Update(map[string]interface{}{
			"network": 0,
		})
	err = DbExec(sqlstr)
	if err != nil {
		err = oo.NewError("DbExec(%s) %v", sqlstr, err)
		return
	}

	sqlstr = oo.NewSqler().Table("tb_transactions").
		Update(map[string]interface{}{
			"block_indep_hash": "",
		})
	err = DbExec(sqlstr)
	if err != nil {
		err = oo.NewError("DbExec(%s) %v", sqlstr, err)
		return
	}
	return
}

func (this *StoreActor) UpdateChainTips(tips []types.TbChain) (err error) {
	if len(tips) < 1 {
		return
	}

	sqlstr := "DELETE FROM tb_chain"
	err = DbExec(sqlstr)
	if nil != err {
		err = oo.NewError("DbExec(%s) %v", sqlstr, err)
		return
	}

	var data []map[string]interface{}
	for _, tip := range tips {
		data = append(data, oo.Struct2Map(tip))
	}

	sqlstr = oo.NewSqler().Table("tb_chain").
		InsertBatch(data)
	err = DbExec(sqlstr)
	if nil != err {
		err = oo.NewError("DbExec(%s) %v", sqlstr, err)
		return
	}

	// var (
	// 	hash1 []string
	// 	hash2 []string
	// )
	// for _, val := range tips {
	// 	if val.IsActive > 0 {
	// 		hash1 = append(hash1, val.IndepHash)
	// 	} else {
	// 		hash2 = append(hash2, val.IndepHash)
	// 	}
	// }
	// if len(hash1) > 0 {
	// 	sqlstr = oo.NewSqler().Table("tb_blocks").
	// 		Where("indep_hash", "in", hash1).
	// 		Update(map[string]interface{}{
	// 			"network": 1,
	// 		})
	// 	err = DbExec(sqlstr)
	// 	if nil != err {
	// 		err = oo.NewError("DbExec(%s) %v", sqlstr, err)
	// 		return
	// 	}
	// }
	// if len(hash2) > 0 {
	// 	sqlstr = oo.NewSqler().Table("tb_blocks").
	// 		Where("indep_hash", "in", hash2).
	// 		Update(map[string]interface{}{
	// 			"network": 0,
	// 		})
	// 	err = DbExec(sqlstr)
	// 	if nil != err {
	// 		err = oo.NewError("DbExec(%s) %v", sqlstr, err)
	// 		return
	// 	}
	// }

	return
}

func (this *StoreActor) ReadFullBlockByHeight(height int64) (blocks []types.FullBlock, err error) {
	var arr []string

	sqlstr := oo.NewSqler().Table("tb_blocks").
		Where("height", height).
		Select("indep_hash")
	err = DbSelect(sqlstr, &arr)
	if nil != err {
		err = oo.NewError("DbSelect(%s) %v", sqlstr, err)
		return
	}

	for _, block_hash := range arr {
		var block *types.FullBlock
		block, err = this.ReadFullBlockByHash(block_hash)
		if nil != err {
			return
		}
		blocks = append(blocks, *block)
	}

	return
}

func (this *StoreActor) ReadFullBlockByHash(block_hash string) (block *types.FullBlock, err error) {
	blk, err := this.ReadBlockByHash(block_hash)
	if nil != err {
		return
	}
	block = &types.FullBlock{
		TbBlock: blk,
	}

	txs, err := this.ReadTxsByBlockHash(blk.IndepHash)
	if nil != err {
		return
	}

	for i, tx := range txs {
		full_tx := types.FullTx{
			TbTransaction: &txs[i],
		}
		if "" != tx.DataHash {
			var buf []byte
			buf, err = this.ReadTxData(tx.ID)
			if nil != err {
				return
			}
			full_tx.Data = buf
		}
		block.Txs = append(block.Txs, full_tx)
	}

	return
}

func (this *StoreActor) ReadFullTx(tx_hash string) (tx *types.FullTx, err error) {
	t, err := this.ReadTx(tx_hash)
	if nil != err {
		return
	}

	tx = &types.FullTx{
		TbTransaction: t,
	}

	if "" != tx.DataHash {
		tx.Data, err = this.ReadTxData(tx_hash)
		if nil != err {
			return
		}
	}

	return
}

func (this *StoreActor) GetAddrMinedHeights(addr string, h1, h2 int64, network int) (hs []int64, err error) {
	sqlstr := oo.NewSqler().Table("tb_blocks").
		Where("reward_addr", addr).
		Where("height", ">", h1).
		Where("height", "<", h2).
		Where("network", network).
		Select("height")
	err = DbSelect(sqlstr, &hs)
	if nil != err {
		return
	}

	return
}

// tb_pool

func (this *StoreActor) TxPoolAdd(tx *types.TbTransaction) (err error) {
	data := types.TbPool{
		TxId:        tx.ID,
		FromAddress: tx.FromAddress,
		Target:      tx.Target,
		LastTx:      tx.LastTx,
	}
	sqlstr := oo.NewSqler().Table("tb_pool").
		Insert(oo.Struct2Map(data))
	err = DbExec(sqlstr)
	if nil != err {
		err = oo.NewError("DbExec(%s) %v", sqlstr, err)
		return
	}

	return
}

func (this *StoreActor) TxPoolDel(tx_ids []string) (err error) {
	if len(tx_ids) <= 0 {
		return
	}
	sqlstr := oo.NewSqler().Table("tb_pool").
		Where("tx_id", "in", tx_ids).
		Delete()
	err = DbExec(sqlstr)
	if nil != err {
		err = oo.NewError("DbExec(%s) %v", sqlstr, err)
		return
	}

	return
}

func (this *StoreActor) TxPoolGet() (tx_ids []string, err error) {
	sqlstr := oo.NewSqler().Table("tb_pool").
		Order("id asc").
		Select("tx_id")
	err = DbSelect(sqlstr, &tx_ids)
	if nil != err {
		err = oo.NewError("DbSelect(%s) %v", sqlstr, err)
		return
	}

	return
}

func (this *StoreActor) TxPoolGetLastTx(addr string) (last_tx string, err error) {
	sqlstr := oo.NewSqler().Table("tb_pool").
		Where("from_address", addr).
		Order("id desc").
		Limit(1).
		Select("tx_id")
	err = DbGet(sqlstr, &last_tx)
	if nil != err {
		if ErrNoRows == err {
			return
		}
		err = oo.NewError("DbGet(%s) %v", sqlstr, err)
		return
	}

	return
}

func (this *StoreActor) TxPoolPrune() (err error) {
	ids, err := this.TxPoolGet()
	if nil != err {
		return
	}

	if 0 == len(ids) {
		return
	}

	var (
		tx  *types.TbTransaction
		txs []*types.TbTransaction
	)
	for _, id := range ids {
		tx, err = this.ReadTx(id)
		if nil != err {
			return
		}
		txs = append(txs, tx)
	}

	var del []string
	is_del := func(str string) bool {
		for _, val := range del {
			if str == val {
				return true
			}
		}
		return false
	}

	for _, tx := range txs {
		sqlstr := oo.NewSqler().Table("tb_transactions").
			Where("from_address", tx.FromAddress).
			Where("last_tx", tx.LastTx).
			Where("block_indep_hash != ''").
			Count()
		var count int64
		count, err = DbGetInt64(sqlstr)
		if nil != err {
			return
		}
		if count > 0 {
			del = append(del, tx.ID)
			continue
		}

		if is_del(tx.LastTx) {
			del = append(del, tx.ID)
			continue
		}
	}

	if len(del) > 0 {
		err = this.TxPoolDel(del)
	}

	return
}
