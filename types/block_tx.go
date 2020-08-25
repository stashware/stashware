package types

import (
	"math/big"
	"sort"
	"strings"

	"stashware/oo"
)

type Block struct {
	IndepHash      string        `json:"indep_hash" validate:"len=64,hexadecimal"`
	Hash           string        `json:"hash" validate:"len=64,hexadecimal"` // satisfy the block's difficulty
	Height         int64         `json:"height" validate:"gte=0"`
	PreviousBlock  string        `json:"previous_block" validate:"len=64,hexadecimal"`
	Nonce          string        `json:"nonce" validate:"gte=0,hexadecimal"`
	Timestamp      int64         `json:"timestamp" validate:"gte=0"`
	LastRetarget   int64         `json:"last_retarget" validate:"gte=0"`
	Diff           string        `json:"diff" validate:"len=64,hexadecimal"`
	CumulativeDiff int64         `json:"cumulative_diff" validate:"gte=0"`
	RewardFee      int64         `json:"reward_fee" validate:"gte=0"`
	RewardAddr     string        `json:"reward_addr" validate:"btc_addr"`
	RewardPool     int64         `json:"reward_pool" validate:"gte=0"`
	WeaveSize      int64         `json:"weave_size" validate:"gte=0"` // sum of txs data size in all blocks
	BlockSize      int64         `json:"block_size" validate:"gte=0"` // sum of txs data size in this block
	WalletList     []WalletTuple `json:"wallet_list" validate:"gte=0"`
	WalletListHash string        `json:"wallet_list_hash" validate:"len=64,hexadecimal"`
	Txs            []string      `json:"txs" validate:"gte=0"` // a list of tx id
	TxRoot         string        `json:"tx_root" validate:"omitempty,len=64,hexadecimal"`
	Poa            string        `json:"poa" validate:"omitempty,hexadecimal"`
}

func (this *Block) Binary() (ret []byte) {
	ret = append(ret, oo.HexDecStringPad32(this.IndepHash)...)
	ret = append(ret, oo.HexDecStringPad32(this.Hash)...)
	ret = append(ret, oo.IntToBytes(this.Height)...)
	ret = append(ret, oo.HexDecStringPad32(this.PreviousBlock)...)
	ret = append(ret, oo.HexDecStringPad32(this.Nonce)...)
	ret = append(ret, oo.IntToBytes(this.Timestamp)...)
	ret = append(ret, oo.IntToBytes(this.LastRetarget)...)
	ret = append(ret, oo.HexDecStringPad32(this.Diff)...)
	ret = append(ret, oo.IntToBytes(this.CumulativeDiff)...)
	ret = append(ret, oo.IntToBytes(this.RewardFee)...)
	ret = append(ret, oo.Base58DecodeString(this.RewardAddr)...)
	ret = append(ret, oo.IntToBytes(this.RewardPool)...)
	ret = append(ret, oo.IntToBytes(this.WeaveSize)...)
	ret = append(ret, oo.IntToBytes(this.BlockSize)...)
	// ret = append(ret, oo.HexDecStringPad32(this.WalletList)...)
	for _, val := range this.WalletList {
		ret = append(ret, val.BinaryEncode()...)
	}
	ret = append(ret, oo.HexDecStringPad32(this.WalletListHash)...)
	// ret = append(ret, oo.HexDecStringPad32(this.Txs)...)
	for _, val := range this.Txs {
		ret = append(ret, oo.HexDecStringPad32(val)...)
	}
	ret = append(ret, oo.HexDecStringPad32(this.TxRoot)...)
	// ret = append(ret, oo.HexDecStringPad32(this.Poa)...) // none

	return
}

type Transaction struct {
	ID        string `json:"id" validate:"len=64,hexadecimal"`
	LastTx    string `json:"last_tx" validate:"omitempty,len=64,hexadecimal"`
	Owner     string `json:"owner" validate:"gt=0,hexadecimal"`
	From      string `json:"from" validate:"omitempty,btc_addr"`
	Target    string `json:"target" validate:"omitempty,btc_addr"`
	Quantity  int64  `json:"quantity" validate:"gte=0"`
	Reward    int64  `json:"reward" validate:"gt=0"`
	Tags      []Tag  `json:"tags" validate:"gte=0"`
	Data      string `json:"data" validate:"omitempty,hexadecimal"`
	DataHash  string `json:"data_hash" validate:"omitempty,len=64,hexadecimal"`
	Signature string `json:"signature" validate:"gt=0,hexadecimal"`
}

func (this *Transaction) Binary() (ret []byte) {
	ret = append(ret, oo.HexDecStringPad32(this.ID)...)
	ret = append(ret, oo.HexDecStringPad32(this.LastTx)...)
	ret = append(ret, oo.HexDecodeString(this.Owner)...)
	ret = append(ret, oo.Base58DecodeString(this.From)...)
	ret = append(ret, oo.Base58DecodeString(this.Target)...)
	ret = append(ret, oo.IntToBytes(this.Quantity)...)
	ret = append(ret, oo.IntToBytes(this.Reward)...)
	// ret = append(ret, oo.HexDecStringPad32(this.Tags)...)
	for _, val := range this.Tags {
		ret = append(ret, val.BinaryEncode()...)
	}
	ret = append(ret, oo.HexDecodeString(this.Data)...)
	ret = append(ret, oo.HexDecStringPad32(this.DataHash)...)
	ret = append(ret, oo.HexDecodeString(this.Signature)...)

	return
}

// type Poa struct {
// 	Option   int64  `json:"option"`
// 	TxPath   string `json:"tx_path"`
// 	DataPath string `json:"data_path"`
// 	Chunk    string `json:"chunk"`
// }

type Tag struct {
	Name  string `json:"name" validate:"gte=0"`
	Value string `json:"value" validate:"gte=0"`
}

func NewTag(name, value string) Tag {
	return Tag{
		Name:  name,
		Value: value,
	}
}

func (this *Tag) BinaryEncode() (ret []byte) {
	if "" != this.Name {
		ret = append(ret, oo.Str2Bytes(this.Name)...)
	}
	if "" != this.Value {
		ret = append(ret, oo.Str2Bytes(this.Value)...)
	}

	return
}

type WalletTuple struct {
	Wallet   string `json:"wallet" validate:"btc_addr"`
	Quantity int64  `json:"quantity" validate:"gte=0"`
	LastTx   string `json:"last_tx" validate:"omitempty,len=64,hexadecimal"`
}

func (this *WalletTuple) BinaryEncode() (ret []byte) {
	ret = append(ret, oo.Base58DecodeString(this.Wallet)...)
	ret = append(ret, oo.IntToBytes(this.Quantity)...)
	ret = append(ret, oo.HexDecStringPad32OrNil(this.LastTx)...)

	return
}

type WalletTuples []WalletTuple

func (this WalletTuples) Len() int { return len(this) }

func (this WalletTuples) Swap(i, j int) { this[i], this[j] = this[j], this[i] }

func (this WalletTuples) Less(i, j int) bool {
	return strings.Compare(this[i].Wallet, this[j].Wallet) > 0
}

func SortWalletTuples(arr []WalletTuple) {
	sort.Sort(WalletTuples(arr))
}

func (this *Block) BigDiff() *big.Int {
	num, _ := new(big.Int).SetString(this.Diff, 16)

	return num
}

func (this *Block) BigHash() *big.Int {
	num, _ := new(big.Int).SetString(this.Hash, 16)

	return num
}

func (this *Block) DataSegment() (ret []byte) {
	generate_base := func() []byte {
		var base []byte
		base = append(base, oo.IntToBytes(this.Height)...)

		hex_arr := []string{
			this.PreviousBlock,
			this.TxRoot,
		}
		for _, val := range this.Txs {
			hex_arr = append(hex_arr, val)
		}

		var buf []byte
		for _, str := range hex_arr {
			buf = oo.HexDecStringPad32(str)
			base = append(base, buf...)
		}
		base = append(base, oo.IntToBytes(this.BlockSize)...)
		base = append(base, oo.IntToBytes(this.WeaveSize)...)

		buf = oo.Base58DecodeString(this.RewardAddr)
		base = append(base, buf...)

		// poa_to_list
		base = append(base, oo.HexDecodeString(this.Poa)...)

		return oo.Sha256(base)
	}

	var buf []byte
	buf = generate_base()
	buf = append(buf, oo.IntToBytes(this.Timestamp)...)
	buf = append(buf, oo.IntToBytes(this.LastRetarget)...)
	buf = append(buf, oo.HexDecStringPad32(this.Diff)...)
	buf = append(buf, oo.IntToBytes(this.CumulativeDiff)...)
	buf = append(buf, oo.IntToBytes(this.RewardPool)...)
	buf = append(buf, this.HashWalletList()...)

	ret = oo.Sha256(buf)

	return
}

func (this *Block) HashWalletList() []byte {
	if len(this.WalletList) <= 0 {
		return nil
	}

	var buf []byte
	for _, val := range this.WalletList {
		buf = append(buf, val.BinaryEncode()...)
	}

	return oo.Sha256(buf)
}

func (this *Transaction) SignatureDataSegment() (ret []byte) {
	ret = append(ret, oo.HexDecodeString(this.Owner)...)
	ret = append(ret, oo.Base58DecodeString(this.Target)...)
	ret = append(ret, oo.HexDecStringPad32OrNil(this.DataHash)...)
	// ret = append(ret, this.RawData()...)
	ret = append(ret, oo.IntToBytes(this.Quantity)...)
	ret = append(ret, oo.IntToBytes(this.Reward)...)
	ret = append(ret, oo.HexDecStringPad32OrNil(this.LastTx)...)

	for _, val := range this.Tags {
		tmp := val.BinaryEncode()
		if len(tmp) > 0 {
			ret = append(ret, tmp...)
		}
	}

	return
}

func (this *Transaction) RawData() []byte {
	return oo.HexDecodeString(this.Data)
}
