package oo

import (
	"github.com/ngchain/go-randomx"
)

type RandxVm struct {
	cache   randomx.Cache
	dataset randomx.Dataset
	vm      randomx.VM
}

func NewRandxVm(key []byte) (ret *RandxVm, err error) {
	cache, err := randomx.AllocCache(randomx.FlagDefault, randomx.FlagJIT)
	if nil != err {
		return
	}
	randomx.InitCache(cache, key)

	dataset, err := randomx.AllocDataset(randomx.FlagDefault, randomx.FlagJIT)
	if nil != err {
		return
	}
	randomx.InitDataset(dataset, cache, 0, randomx.DatasetItemCount()) // todo: multi core acceleration

	vm, err := randomx.CreateVM(cache, dataset, randomx.FlagJIT, randomx.FlagHardAES, randomx.FlagFullMEM)
	if nil != err {
		return
	}

	ret = &RandxVm{
		cache:   cache,
		dataset: dataset,
		vm:      vm,
	}

	return
}

func (this *RandxVm) Close() {
	randomx.DestroyVM(this.vm)
	randomx.ReleaseDataset(this.dataset)
	randomx.ReleaseCache(this.cache)
}

func (this *RandxVm) Hash(buf []byte) (ret []byte) {
	return randomx.CalculateHash(this.vm, buf)
}

func RandomX(key, buf []byte) (ret []byte, err error) {
	vm, err := NewRandxVm(key)
	if nil != err {
		return
	}
	defer vm.Close()

	ret = vm.Hash(buf)

	return
}
