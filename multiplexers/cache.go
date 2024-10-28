package multiplexers

import (
	"github.com/orbit-w/mux-go"
	"sync"
)

type connCache struct {
	rw    sync.RWMutex
	err   error
	conns map[int64]IConn
}

func newConnCache() *connCache {
	return &connCache{
		conns: make(map[int64]IConn),
	}
}

func (cc *connCache) Store(id int64, conn IConn) error {
	cc.rw.Lock()
	defer cc.rw.Unlock()
	if cc.err != nil {
		return cc.err
	}
	cc.conns[id] = conn
	return nil
}

func (cc *connCache) Delete(id int64) {
	cc.rw.Lock()
	defer cc.rw.Unlock()
	if cc.err != nil {
		return
	}
	delete(cc.conns, id)
}

func (cc *connCache) OnClose(iter func(conn IConn)) {
	cc.rw.RLock()
	if cc.err != nil {
		cc.rw.RUnlock()
		return
	}
	cc.err = mux.ErrCancel
	t := make([]IConn, 0, len(cc.conns))
	for k := range cc.conns {
		conn := cc.conns[k]
		t = append(t, conn)
	}
	cc.conns = nil
	cc.rw.RUnlock()

	for i := range t {
		conn := t[i]
		iter(conn)
	}
}
