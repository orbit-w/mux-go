package mux

import (
	"sync"
	"sync/atomic"
)

/*
   @Author: orbit-w
   @File: streamers
   @2024 7月 周日 17:56
*/

type VirtualConns struct {
	idx   atomic.Int64
	max   int
	err   error
	rw    sync.RWMutex
	conns map[int64]*VirtualConn
}

func newConns(max int) *VirtualConns {
	return &VirtualConns{
		max:   max,
		rw:    sync.RWMutex{},
		conns: make(map[int64]*VirtualConn),
	}
}

func (ins *VirtualConns) Id() int64 {
	return ins.idx.Add(1)
}

func (ins *VirtualConns) Get(id int64) (*VirtualConn, bool) {
	ins.rw.RLock()
	s, ok := ins.conns[id]
	ins.rw.RUnlock()
	return s, ok
}

func (ins *VirtualConns) Exist(id int64) (exist bool) {
	ins.rw.RLock()
	_, exist = ins.conns[id]
	ins.rw.RUnlock()
	return
}

func (ins *VirtualConns) Len() int {
	ins.rw.RLock()
	defer ins.rw.RUnlock()
	return len(ins.conns)
}

func (ins *VirtualConns) Reg(id int64, s *VirtualConn) error {
	ins.rw.Lock()
	defer ins.rw.Unlock()
	if ins.err != nil {
		return ins.err
	}
	if ins.max != 0 && len(ins.conns) >= ins.max {
		return ErrVirtualConnUpLimit
	}
	ins.conns[id] = s
	return nil
}

func (ins *VirtualConns) Del(id int64) {
	ins.rw.Lock()
	delete(ins.conns, id)
	ins.rw.Unlock()
}

func (ins *VirtualConns) GetAndDel(id int64) (*VirtualConn, bool) {
	ins.rw.Lock()
	s, exist := ins.conns[id]
	if exist {
		delete(ins.conns, id)
	}
	ins.rw.Unlock()
	return s, exist
}

func (ins *VirtualConns) Range(iter func(stream *VirtualConn)) {
	ins.rw.RLock()
	for k := range ins.conns {
		stream := ins.conns[k]
		iter(stream)
	}
	ins.rw.RUnlock()
}

func (ins *VirtualConns) OnClose(onClose func(stream *VirtualConn)) {
	ins.rw.Lock()
	defer ins.rw.Unlock()
	ins.err = ErrCancel
	for k := range ins.conns {
		stream := ins.conns[k]
		onClose(stream)
	}
	ins.conns = make(map[int64]*VirtualConn)
}
