package multiplexers

import (
	"context"
	"errors"
	"fmt"
	pq "github.com/orbit-w/meteor/bases/container/priority_queue"
	"github.com/orbit-w/meteor/modules/net/transport"
	"github.com/orbit-w/mux-go"
	"sync"
	"sync/atomic"
)

/*
   @Author: orbit-w
   @File: multiplexers
   @2024 8月 周四 23:47
*/

type Multiplexers struct {
	state    atomic.Uint32
	rw       sync.RWMutex
	host     string
	muxIdx   int
	connIdx  atomic.Int64
	size     int //mux的数量
	maxConns int //mux对应的最大虚拟连接数
	pq       *pq.PriorityQueue[int, mux.IMux, int]
	tempMap  *connCache
}

func NewMultiplexers(host string, size, connsCount int) *Multiplexers {
	m := &Multiplexers{
		host:     host,
		size:     size,
		maxConns: connsCount,
		pq:       pq.New[int, mux.IMux, int](),
		tempMap:  newConnCache(),
	}

	for i := 0; i < size; i++ {
		ctx := context.Background()
		conn := transport.DialContextWithOps(ctx, host)
		multiplexer := mux.NewMultiplexer(ctx, conn, mux.NewClientConfig(m.maxConns))
		m.pq.Push(i, multiplexer, 0)
	}
	return m
}

// Dial 方法严格按照绑定的最小虚拟连接数优先选择多路复用器来创建虚拟连接
func (m *Multiplexers) Dial() (IConn, error) {
	m.rw.Lock()
	defer m.rw.Unlock()
	if m.pq.Empty() {
		return nil, fmt.Errorf("no available multiplexers")
	}

	// Get the multiplexer with the least number of virtual connections
	idx, multiplexer, _ := m.pq.Peek()

	vc, err := multiplexer.NewVirtualConn(context.Background())
	if err != nil {
		if !errors.Is(err, mux.ErrVirtualConnUpLimit) {
			return nil, err
		}
		return m.newTempConn()
	}

	iConn := wrapConn(vc, m.connId(), func() {
		m.rw.Lock()
		defer m.rw.Unlock()
		m.pq.UpdatePriorityOp(idx, func(s int) int {
			return s - 1
		})
	})

	m.pq.UpdatePriorityOp(idx, func(s int) int {
		return s + 1
	})
	return iConn, nil
}

// The DialEdge method does not strictly prioritize selecting the multiplexer with the fewest virtual connections to create a new virtual connection.
// By minimizing the granularity of locks, it significantly improves concurrency performance.
// However, under high concurrent requests, it may lead to an imbalance in the number of virtual connections on the preferred multiplexer.
// =========================
// DialEdge 方法不会严格按照最小虚拟连接数优先选择多路复用器来创建虚拟连接。
// 通过尽可能降低锁的粒度，极大提高了并发性能。
// 但在并发请求特别大时，可能会导致优先选择的多路复用器上的虚拟连接数不均衡。
func (m *Multiplexers) DialEdge() (IConn, error) {
	m.rw.RLock()
	if m.pq.Empty() {
		m.rw.RUnlock()
		return nil, fmt.Errorf("no available multiplexers")
	}

	// Get the multiplexer with the least number of virtual connections
	idx, multiplexer, _ := m.pq.Peek()
	m.rw.RUnlock()

	vc, err := multiplexer.NewVirtualConn(context.Background())
	if err != nil {
		if errors.Is(err, mux.ErrVirtualConnUpLimit) {
			return m.newTempConn()
		}
		return nil, err
	}

	mw := wrapConn(vc, m.connId(), func() {
		m.rw.Lock()
		defer m.rw.Unlock()
		//TODO: pq 有竞态风险，可能为nil
		m.pq.UpdatePriorityOp(idx, func(s int) int {
			return s - 1
		})
	})

	m.rw.Lock()
	m.pq.UpdatePriorityOp(idx, func(s int) int {
		return s + 1
	})
	m.rw.Unlock()
	return mw, nil
}

func (m *Multiplexers) Close() {
	if !m.state.CompareAndSwap(StateNormal, StateStopped) {
		return
	}

	m.rw.Lock()
	if m.pq != nil {
		if !m.pq.Empty() {
			for {
				_, multiplexer, exist := m.pq.Pop()
				if !exist {
					break
				}
				multiplexer.Close()
			}
		}
		m.pq = pq.New[int, mux.IMux, int]()
	}
	m.rw.Unlock()

	m.tempMap.OnClose(func(value IConn) {
		conn := value.(IConn)
		_ = conn.Close()
	})
}

func (m *Multiplexers) newTempConn() (IConn, error) {
	// All multiplexers are at limit, create a new one
	ctx := context.Background()
	conn := transport.DialContextWithOps(ctx, m.host)
	multiplexer := mux.NewMultiplexer(ctx, conn)
	vc, err := multiplexer.NewVirtualConn(ctx)
	if err != nil {
		return nil, err
	}
	idx := m.connId()
	vConn := wrapConn(vc, idx, func() {
		multiplexer.Close()
	})
	m.tempMap.Store(idx, vConn)
	return vConn, nil
}

func (m *Multiplexers) connId() int64 {
	return m.connIdx.Add(1)
}
