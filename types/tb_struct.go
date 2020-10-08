package types

// tb_blocks
type TbBlock struct {
	IndepHash      string `db:"indep_hash"`
	Hash           string `db:"hash"`
	Height         int64  `db:"height"`
	PreviousBlock  string `db:"previous_block"`
	Nonce          string `db:"nonce"`
	Timestamp      int64  `db:"timestamp"`
	LastRetarget   int64  `db:"last_retarget"`
	Diff           string `db:"diff"`
	CumulativeDiff int64  `db:"cumulative_diff"`
	RewardAddr     string `db:"reward_addr"`
	RewardFee      int64  `db:"reward_fee"`
	RewardPool     int64  `db:"reward_pool"`
	WeaveSize      int64  `db:"weave_size"`
	BlockSize      int64  `db:"block_size"`
	Txs            string `db:"txs"`
	TxRoot         string `db:"tx_root"`
	WalletList     string `db:"wallet_list"`
	WalletListHash string `db:"wallet_list_hash"`
	Network        int64  `db:"network"`
}

// tb_transactions
type TbTransaction struct {
	Row            int64  `db:"row" sqler:"skips"`
	Ctime          string `db:"ctime" sqler:"skips"`
	ID             string `db:"id"`
	BlockIndepHash string `db:"block_indep_hash"`
	LastTx         string `db:"last_tx"`
	Owner          string `db:"owner"`
	FromAddress    string `db:"from_address"`
	Target         string `db:"target"`
	Quantity       int64  `db:"quantity"`
	Signature      string `db:"signature"`
	Reward         int64  `db:"reward"`
	Tags           string `db:"tags"`
	DataHash       string `db:"data_hash"`
}

// tb_wallet
type TbWallet struct {
	Id      int64  `db:"id" sqler:"skips"`
	Address string `db:"address"`
	LastTx  string `db:"last_tx"`
	Balance int64  `db:"balance"`
}

// tb_chain
type TbChain struct {
	IndepHash string `db:"indep_hash"`
	Height    int64  `db:"height"`
	HashList  string `db:"hash_list"` // json([]string)
	IsActive  int64  `db:"is_active"`
}

// tb_pool
type TbPool struct {
	Id          int64  `db:"id" sqler:"skips"`
	TxId        string `db:"tx_id"`
	FromAddress string `db:"from_address"`
	Target      string `db:"target"`
	LastTx      string `db:"last_tx"`
	Ctime       string `db:"ctime" sqler:"skips"`
}

//
// wrapper
//
type FullBlock struct {
	*TbBlock
	Txs []FullTx
}

type FullTx struct {
	*TbTransaction
	Data []byte
}

type ActiveBlock = struct {
	Height    int64  `db:"height"`
	IndepHash string `db:"indep_hash"`
}
