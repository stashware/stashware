package types

import (
	"stashware/oo"
)

func (this *Block) TbBlock() (ret *TbBlock) {
	ret = &TbBlock{
		IndepHash:      this.IndepHash,
		Hash:           this.Hash,
		Height:         this.Height,
		PreviousBlock:  this.PreviousBlock,
		Nonce:          this.Nonce,
		Timestamp:      this.Timestamp,
		LastRetarget:   this.LastRetarget,
		Diff:           this.Diff,
		CumulativeDiff: this.CumulativeDiff,
		RewardAddr:     this.RewardAddr,
		RewardFee:      this.RewardFee,
		RewardPool:     this.RewardPool,
		WeaveSize:      this.WeaveSize,
		BlockSize:      this.BlockSize,
		Txs:            string(oo.JsonData(this.Txs)),
		TxRoot:         this.TxRoot,
		WalletList:     string(oo.JsonData(this.WalletList)),
		WalletListHash: this.WalletListHash,
		Network:        NETWORK_MAINNET,
	}

	return
}

func (this *Transaction) TbTransaction() (ret *TbTransaction) {
	ret = &TbTransaction{
		ID: this.ID,
		// BlockIndepHash: this.BlockIndepHash,
		LastTx:      this.LastTx,
		Owner:       this.Owner,
		FromAddress: this.From,
		Target:      this.Target,
		Quantity:    this.Quantity,
		Reward:      this.Reward,
		Tags:        string(oo.JsonData(this.Tags)),
		// this.Data,
		DataHash:  this.DataHash,
		Signature: this.Signature,
	}

	return
}

func (this *TbBlock) Block() (ret *Block) {
	ret = &Block{
		IndepHash:      this.IndepHash,
		Hash:           this.Hash,
		Height:         this.Height,
		PreviousBlock:  this.PreviousBlock,
		Nonce:          this.Nonce,
		Timestamp:      this.Timestamp,
		LastRetarget:   this.LastRetarget,
		Diff:           this.Diff,
		CumulativeDiff: this.CumulativeDiff,
		RewardAddr:     this.RewardAddr,
		RewardFee:      this.RewardFee,
		RewardPool:     this.RewardPool,
		WeaveSize:      this.WeaveSize,
		BlockSize:      this.BlockSize,
		TxRoot:         this.TxRoot,
		WalletListHash: this.WalletListHash,
	}
	if len(this.Txs) > 0 {
		_ = oo.JsonUnmarshal([]byte(this.Txs), &ret.Txs)
	}
	if len(this.WalletList) > 0 {
		_ = oo.JsonUnmarshal([]byte(this.WalletList), &ret.WalletList)
	}

	return
}

func (this *TbTransaction) Transaction() (ret *Transaction) {
	ret = &Transaction{
		ID:        this.ID,
		LastTx:    this.LastTx,
		Owner:     this.Owner,
		From:      this.FromAddress,
		Target:    this.Target,
		Quantity:  this.Quantity,
		Reward:    this.Reward,
		Signature: this.Signature,
		DataHash:  this.DataHash,
	}
	if len(this.Tags) > 0 {
		_ = oo.JsonUnmarshal([]byte(this.Tags), &ret.Tags)
	}

	return
}

func (this *FullTx) Transaction() (ret *Transaction) {
	ret = this.TbTransaction.Transaction()

	if len(this.Data) > 0 {
		ret.Data = oo.HexEncodeToString(this.Data)
	}

	return
}
