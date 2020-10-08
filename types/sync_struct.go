package types

import (
	"stashware/oo"
)

const WAIT_RPC_SHORT = int64(30)
const WAIT_RPC_MIDDLE = int64(60)
const WAIT_RPC_LONG = int64(600)

/*
	GET /info
	GET /peers

	GET /tx/pending				//[]txid  Return all mempool transactions.

	GET /block/current  		//blockshadow
	GET /block/{height,hash}/XXX
	GET /block/{height,hash}/XXX/{wallet_list,hash_list} // field in [wallet_list, hash_list, ...]

	GET /wallet/ADDR 			//balance, last_tx, publickey(active)

	GET /price/Bytes[/targetADDR]
	GET /TXID[.EXT] 			//just data
	GET /tx/TXID 				//confirmations, block_indep_hash, with data

	POST /block     			//header:block-hash. blockshadow[no wallet list]. short_hash_list, block_data_segment field
	POST /tx					//header:tx-id.

	-- GET /time 				//system time
	-- GET /queue 					//[]{txid, Reward, Size}. outgoing priority queue
	-- GET /wallet/ADDR/balance
	-- GET /wallet/ADDR/last_tx
	-- GET /wallet/ADDR/txs		// []txid
	-- GET /wallet/ADDR/deposits 	// []txid
	-- GET /tx/TXID/status  	//block_height, block_indep_hash, number_of_confirmations
	-- POST /peers
	-- POST /wallet 			//
	-- GET /wallet  			//create new wallet
	-- GET /tx_anchor 			//self wallet last_tx
*/

const CMD_INFO = "info"

type CmdInfoReq = struct {
	Network     string `json:"network" validate:"omitempty,required"`
	Genesis 	string `json:"genesis" validate:"omitempty,required"`
	Version     string `json:"version" validate:"omitempty,required"`
	Height      int64  `json:"height" validate:"omitempty,required"`
	SyncHeight  int64  `json:"sync_height" validate:"omitempty"`
	Current     string `json:"current" validate:"omitempty,len=64,hexadecimal"`
	Blocks      int64  `json:"blocks" validate:"omitempty,required"`
	Peers       int64  `json:"peers" validate:"omitempty,gte=0"`
	QueueLength int64  `json:"queue_length" validate:"omitempty,gte=0"`
	Uuid        string `json:"uuid" validate:"omitempty,uuid4"` //"" not valid
}

type CmdInfoRsp = CmdInfoReq

const CMD_PEERS = "peers"

type CmdPeersReq = oo.ParaNull
type CmdPeersRsp = struct {
	EndPoint []string `json:"end_point"` //["ws://3333:1981", "wss://[FF01::1101]1981"]
}

const CMD_DEBUG = "debug"

type CmdDebugReq = oo.ParaNull
type CmdDebugRsp = struct {
	Scores []string `json:"scores"`
}

const CMD_TX_PENDING = "tx_pending"

type CmdTxPendingReq = oo.ParaNull
type CmdTxPendingRsp = struct {
	Txs []string `json:"txs"`
}

const CMD_TX_BY_ID = "tx_by_id"

type CmdTxByIdReq = struct {
	Txid string `json:"txid" validate:"len=64,hexadecimal"`
}
type CmdTxByIdRsp = struct {
	Transaction
	Timestamp      int64  `json:"timestamp"`
	Confirmations  int64  `json:"confirmations"`
	BlockIndepHash string `json:"block_indep_hash"`
}

const CMD_TX_DATA = "tx_data"

type CmdTxDataReq = struct {
	Txid      string `json:"txid" validate:"len=64,hexadecimal"`
	Extension string `json:"extension,omitempty"`
}

// type CmdTxDataRsp = []byte
type CmdTxDataRsp = string // hex(data)

const CMD_GET_ADDR_TXS = "get_addr_txs"

type CmdGetAddrTxsReq struct {
	Address string `json:"address" validate:"btc_addr"`
	Height1 *int64 `json:"height1" validate:"omitempty,gte=0"`
	Height2 *int64 `json:"height2" validate:"omitempty,gtfield=Height1"`
}

type CmdGetAddrTxsRsp struct {
	Data []string `json:"data"`
}

const CMD_PRICE = "price"

type CmdPriceReq = struct {
	Bytes  int64  `json:"bytes" validate:"gte=0"`
	Target string `json:"target" validate:"omitempty,btc_addr"`
}

type CmdPriceRsp = struct {
	Amount int64 `json:"amount,omitempty"`
}

const CMD_SUBMIT_TX = "submit_tx"

type CmdSubmitTxReq = struct {
	Transaction
}
type CmdSubmitTxRsp = oo.ParaNull

//
// new wallet and return private-key, public-key, address, which are encoded by base58
//
const CMD_WALLET_NEW = "wallet_new"

type CmdWalletNewReq struct {
}

type CmdWalletNewRsp struct {
	Address    string `json:"address"`
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key"`
}

const CMD_WALLET_BY_ADDR = "wallet_by_addr"

type CmdWalletByAddrReq = struct {
	Address string `json:"address" validate:"btc_addr"`
}

type CmdWalletByAddrRsp = struct {
	Pending int64  `json:"pending"`
	Balance int64  `json:"balance"`
	LastTx  string `json:"last_tx"`
	Pos     int64  `json:"pos"`
	// PubKey  string `json:"pub_key"` //if exists : wallet actived
}

const CMD_GET_BLOCK = "get_block"

type CmdGetBlockReq = struct {
	Current     int64  `json:"current" validate:"gte=0"` //if==1, return tip of chain
	BlockId     string `json:"block_id" validate:"omitempty,len=64,hexadecimal"`
	BlockHeight int64  `json:"block_height" validate:"gte=0"`
}

type CmdGetBlockRsp = struct {
	Block
}

const CMD_GET_WALLET_LIST = "get_wallet_list"

type CmdGetWalletListReq = CmdGetBlockReq
type CmdGetWalletListRsp = struct {
	WalletList []WalletTuple `json:"wallet_list"`
}

const CMD_SUBMIT_BLOCK = "submit_block"

type CmdSubmitBlockReq = struct {
	NewBlock Block    `json:"new_block"` //no wallet_list
	HashList []string `json:"hash_list" validate:"gte=0"`
	BDS      string   `json:"bds" validate:"omitempty,hexadecimal"`
}
type CmdSubmitBlockRsp = oo.ParaNull

const CMD_MINING_INFO = "mining"

type MiningInfo = struct {
	Core     int64 `json:"core"`
	HashRate int64 `json:"hash_rate"`
}
