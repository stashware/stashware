package oo

import (
	"errors"
	"sync"
	"sync/atomic"
)

type ChannelManager struct {
	rpcSequence uint64
	channelMap  sync.Map
	chlen       uint64
	DebugFlag   int64
}

//channel pair
type Channel struct {
	Data     interface{}
	sequence uint64
	respChan chan interface{}
	closed   chan struct{}
	Chmgr    *ChannelManager
}

func (ch *Channel) GetSeq() uint64 {
	return ch.sequence
}
func (ch *Channel) GetSeqType() int {
	return int(ch.sequence >> 56)
}

func (ch *Channel) Close() {
	ch.Chmgr.DelChannel(ch)
}

func (ch *Channel) IsClosed() <-chan struct{} {
	return ch.closed
}

func (ch *Channel) PushMsg(m interface{}) (err error) {
	defer func() {
		if errs := recover(); errs != nil {
			LogW("Failed recover: %v", errs)
		}
	}()
	// if ch.Chmgr.DebugFlag > 0 {
	// 	LogD("ch  %d PushMsg %s", ch.sequence, m)
	// }
	select {
	case ch.respChan <- m:
		return nil
	case <-ch.closed:
		return errors.New("chan closed")
	default:
		return errors.New("chan full")
	}
}

func (ch *Channel) RecvChan() <-chan interface{} {
	return ch.respChan
}

//-----
var g_seq uint64

func (cm *ChannelManager) NewChannel(seq_high8 ...int) *Channel {
	chlen := uint64(64)
	pseq := &g_seq
	if cm != nil {
		chlen = cm.chlen
		pseq = &cm.rpcSequence
	}
	ch := &Channel{respChan: make(chan interface{}, chlen), closed: make(chan struct{})}
	ch.sequence = atomic.AddUint64(pseq, 1)
	if len(seq_high8) != 0 {
		ch.sequence += uint64(seq_high8[0]&0xFF) << 56
	}
	ch.Chmgr = cm
	if cm != nil {
		cm.channelMap.Store(ch.sequence, ch)
	}

	return ch
}
func (cm *ChannelManager) DelChannel(ch *Channel) {
	if cm != nil {
		cm.channelMap.Delete(ch.sequence)
	}
	close(ch.closed)
	close(ch.respChan)
}
func (cm *ChannelManager) GetChannel(seq uint64) (*Channel, error) {
	if v, ok := cm.channelMap.Load(seq); ok {
		ch, _ := v.(*Channel)
		return ch, nil
	}

	return nil, errors.New("not found")
}
func (cm *ChannelManager) PushChannelMsg(seq uint64, m interface{}) error {
	if nil == cm {
		return errors.New("no manager")
	}
	if v, ok := cm.channelMap.Load(seq); ok {
		ch, _ := v.(*Channel)
		return ch.PushMsg(m)
	} else {
		return errors.New("not found")
	}
}
func (cm *ChannelManager) PushAllChannelMsg(m interface{}, fn func(ch *Channel) bool) (n int, err error) {
	if nil == cm {
		err = errors.New("no manager")
		return
	}
	cm.channelMap.Range(func(key, value interface{}) bool {
		ch, _ := value.(*Channel)
		if fn == nil || fn(ch) {
			if ch.PushMsg(m) == nil {
				n++
			}
		}

		return true
	})
	return
}
func (cm *ChannelManager) PushOneChannelMsg(m interface{}, fn func(ch *Channel) bool) (err error) {
	if nil == cm {
		return errors.New("no manager")
	}

	cm.channelMap.Range(func(key, value interface{}) bool {
		ch, _ := value.(*Channel)
		if fn(ch) {
			err = ch.PushMsg(m)
			return false
		}

		return true
	})
	return
}

func NewChannelManager(chlen uint64) *ChannelManager {
	cm := &ChannelManager{1, sync.Map{}, chlen, 0}
	// if GChannelManager == nil {
	// 	GChannelManager = cm
	// }
	return cm
}
