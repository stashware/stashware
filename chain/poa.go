package chain

import (
	// "bytes"
	"encoding/binary"
	"stashware/oo"
	"stashware/storage"
	"stashware/types"
)

func (this *BlockChain) GetPoa(preHash []byte, nowHeight int64) (poa []byte, err error) {

	poaHeight := GetPoaHeight(preHash, nowHeight)
	// if poaHeight == 0 {
	// 	return //dont need poa[]
	// }
	if nil == storage.GStore {
		err = oo.NewError("storer is nil.")
		return
	}
	//read poa block
	b, err := storage.GStore.ReadBlockByHeight(poaHeight)
	if err != nil {
		oo.LogW("no poa block??? nowH=%d, hashLen=%d, poaH=%d, %v", nowHeight, len(preHash), poaHeight, err)
		return
	}

	fb, err := storage.GStore.ReadFullBlockByHash(b.IndepHash)
	if err != nil {
		oo.LogW("failed to load full poa block. h=%d, err %v", b.Height, err)
		return
	}
	soff, slen := GetPoaSeg(preHash, fb.Block())
	poa = SerialBlockSeg(fb, soff, slen)
	return
}

func GetHashInt(preHash []byte) (hashInt int64) {
	if len(preHash) > 8 {
		preHash = preHash[len(preHash)-8:]
	}
	hashInt = int64(binary.BigEndian.Uint64(preHash))
	if hashInt < 0 {
		hashInt = -hashInt
	}
	return
}
func GetPoaHeight(preHash []byte, nowHeight int64) (poaHeight int64) {
	if nowHeight == 0 {
		return
	}
	hashInt := GetHashInt(preHash)
	poaHeight = (hashInt % nowHeight)
	return
}

func GetPoaSeg(preHash []byte, poaBlock *types.Block) (soff int64, slen int64) {
	hashInt := GetHashInt(preHash)
	//align(bHead, 200)+sum(len(txs)*txHead) + block_size
	blockLen := types.BLOCK_HEAD_SIZE + int64(len(poaBlock.Txs))*types.TX_SIZE_BASE + poaBlock.BlockSize
	slen = hashInt % blockLen
	if slen < types.BLOCK_HEAD_SIZE {
		slen = types.BLOCK_HEAD_SIZE
	}
	if slen > int64(32000) {
		slen = int64(32000)
	}
	soff = (blockLen - slen)
	if soff <= 0 {
		soff = 0
	} else {
		soff = hashInt % soff
	}
	return
}

// func EncodeStruct(bh interface{}) (bin []byte) {
// 	buf := new(bytes.Buffer)
// 	if err := binary.Write(buf, binary.LittleEndian, bh); err != nil {
// 		oo.LogW("Failed to encode struct %v", err)
// 		return
// 	}
// 	bin = buf.Bytes()
// 	return
// }

// func EncodeExpand(st interface{}, size int) (bin []byte) {
// 	one := EncodeStruct(st)
// 	bin = one
// 	if len(bin) == 0 {
// 		return
// 	}

// 	for len(bin) < size {
// 		left := size - len(bin)
// 		if left > len(one) {
// 			left = len(one)
// 		}
// 		bin = append(bin, one[:left]...)
// 	}
// 	if len(bin) > size {
// 		bin = bin[:size]
// 	}
// 	return
// }

func EncodeExpand(one []byte, size int) (bin []byte) {
	// one := EncodeStruct(st)
	bin = one
	if len(bin) == 0 {
		return
	}

	for len(bin) < size {
		left := size - len(bin)
		if left > len(one) {
			left = len(one)
		}
		bin = append(bin, one[:left]...)
	}
	if len(bin) > size {
		bin = bin[:size]
	}
	return
}

func CopySeg(bin []byte, soff, slen int64) (retbin []byte, retlen int64) {
	blen := int64(len(bin))
	if soff >= blen {
		return //no data
	}
	if soff < 0 {
		soff = 0
	}

	tlen := slen
	if tlen > blen-soff {
		tlen = blen - soff
	}
	retbin = bin[soff : soff+tlen]
	retlen = slen - tlen
	return
}

func SerialBlockSeg(fb *types.FullBlock, soff int64, slen int64) (seg []byte) {
	if soff < types.BLOCK_HEAD_SIZE {
		bin := EncodeExpand(fb.TbBlock.Block().Binary(), int(types.BLOCK_HEAD_SIZE))
		seg, slen = CopySeg(bin, soff, slen)
		soff = 0
	} else {
		soff -= types.BLOCK_HEAD_SIZE
	}

	for _, tx := range fb.Txs {
		if soff < types.TX_SIZE_BASE && slen > 0 {
			bin := EncodeExpand(tx.Transaction().Binary(), int(types.TX_SIZE_BASE))
			bin, slen = CopySeg(bin, soff, slen)
			seg = append(seg, bin...)
			soff = 0
		} else {

			soff -= types.TX_SIZE_BASE
		}

		if soff < int64(len(tx.Data)) && slen > 0 {
			bin := tx.Data
			bin, slen = CopySeg(bin, soff, slen)
			seg = append(seg, bin...)
			soff = 0
		} else {
			soff -= int64(len(tx.Data))
		}
	}
	// oo.LogD("after serial seg. off-len=[%d,%d], seg len=%d", soff, slen, len(seg))
	return
}
