package multiplexers

import (
	"container/heap"
	"context"
	"fmt"
	"github.com/orbit-w/meteor/modules/net/transport"
	"github.com/orbit-w/mux-go"
	"sync"
)

/*
   @Author: orbit-w
   @File: multiplexers
   @2024 8月 周四 23:47
*/

type Multiplexers struct {
	rw         sync.RWMutex
	host       string
	size       int //mux的数量
	connsCount int //mux对应的最大虚拟连接数
	cache      MultiplexerHeap
}

func (m *Multiplexers) Dial() (IConn, error) {
	m.rw.Lock()
	defer m.rw.Unlock()

	if m.cache.Len() == 0 {
		return nil, fmt.Errorf("no available multiplexers")
	}

	// Get the multiplexer with the least number of virtual connections
	muxWrapper := heap.Pop(&m.cache).(*MultiplexerWrapper)

	if muxWrapper.virtualConns >= m.connsCount {
		// All multiplexers are at limit, create a new one
		conn := transport.DialContextWithOps(context.Background(), m.host)
		multiplexer := mux.NewMultiplexer(context.Background(), conn)
		vc, err := multiplexer.NewVirtualConn(context.Background())
		if err != nil {
			return nil, err
		}
		mw := newConnWrapper(vc, func() {
			multiplexer.Close()
		})
		return mw, nil
	}

	vc, err := muxWrapper.mux.NewVirtualConn(context.Background())
	if err != nil {
		return nil, err
	}

	mw := newConnWrapper(vc, func() {
		m.rw.Lock()
		defer m.rw.Unlock()
		muxWrapper.virtualConns--
		heap.Push(&m.cache, muxWrapper)
	})

	heap.Push(&m.cache, muxWrapper)
	return mw, nil
}
