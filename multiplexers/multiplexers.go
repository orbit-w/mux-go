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
	state    atomic.Int32
	rw       sync.RWMutex
	host     string
	muxIdx   int
	connIdx  atomic.Int64
	muxCount int //mux的数量
	maxConns int //mux对应的最大虚拟连接数
	pq       *pq.PriorityQueue[int, mux.IMux, int]
	tempMap  *connCache
}

func NewWithDefaultConf(host string) *Multiplexers {
	conf := DefaultConfig()
	m := &Multiplexers{
		host:     host,
		muxCount: conf.MuxCount,
		maxConns: conf.MuxMaxConns,
		pq:       pq.New[int, mux.IMux, int](),
		tempMap:  newConnCache(),
	}

	m.init()
	return m
}

func (m *Multiplexers) init() {
	for i := 0; i < m.muxCount; i++ {
		ctx := context.Background()
		conn := transport.DialContextWithOps(ctx, m.host, &transport.DialOption{
			MaxIncomingPacket: MaxIncomingPacket,
		})
		multiplexer := mux.NewMultiplexer(ctx, conn, mux.NewClientConfig(m.maxConns))
		m.pq.Push(i, multiplexer, 0)
	}
}

func (m *Multiplexers) State() int32 {
	return m.state.Load()
}

// Dial 方法严格按照绑定的最小虚拟连接数优先选择多路复用器来创建虚拟连接
func (m *Multiplexers) Dial() (IConn, error) {
	m.rw.Lock()
	defer m.rw.Unlock()

	if m.pq.Empty() {
		return nil, ErrNoAvailableMultiplexers
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

	conn := wrapConn(vc, m.connId(), func() {
		m.rw.Lock()
		defer m.rw.Unlock()
		m.pq.UpdatePriorityOp(idx, decrPriority)
	})

	m.pq.UpdatePriorityOp(idx, incrPriority)
	return conn, nil
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

	//multiplexer.NewVirtualConn and multiplexer.Close do not need to strictly ensure linear execution order. They are thread-safe and will not cause connection or data leaks.
	//They also will not block users on Recv.
	//multiplexer.NewVirtualConn 跟 multiplexer.Close 不需要严格的保证线性顺序执行，并发安全的，不会造成连接等数据泄漏。
	//也不会将用户阻塞在Recv上。
	vc, err := multiplexer.NewVirtualConn(context.Background())
	if err != nil {
		if errors.Is(err, mux.ErrVirtualConnUpLimit) {
			return m.newTempConn()
		}
		return nil, err
	}

	conn := wrapConn(vc, m.connId(), func() {
		m.rw.Lock()
		defer m.rw.Unlock()
		m.pq.UpdatePriorityOp(idx, decrPriority)
	})

	m.rw.Lock()
	m.pq.UpdatePriorityOp(idx, incrPriority)
	m.rw.Unlock()
	return conn, nil
}

// Close closes all multiplexers and virtual connections
//
// steps:
// 1. Atomically sets the state to StateClosed if it is currently StateNone.Make the Close method reentrant and enhance its robustness.
// 2. Acquires a write lock to ensure thread safety while modifying the priority queue and temporary map.
// 3. Iterates through the priority queue, closing each multiplexer and removing it from the queue.
// 4. Iterates through the temporary map, closing each virtual connection stored in it.

// Close 关闭所有多路复用器和虚拟连接
// 步骤：
//
//	1.如果当前状态是 StateNone，则以原子方式将状态设置为 StateClosed。
//	2.获取写锁以确保在修改优先队列和临时映射时的线程安全。
//	3.遍历优先队列，关闭每个多路复用器并将其从队列中移除。此处线程安全。
//	4.遍历临时映射，关闭其中存储的每个虚拟连接。
func (m *Multiplexers) Close() {
	// Atomically set the state to StateClosed if it is currently StateNone
	if !m.state.CompareAndSwap(StateNone, StateClosed) {
		return
	}

	// Acquire a write lock to ensure thread safety
	m.rw.Lock()

	// Iterate through the priority queue, closing each multiplexer
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

		m.pq.Free()
	}
	m.rw.Unlock()

	// Iterate through the temporary map, closing each virtual connection
	m.tempMap.OnClose(func(value IConn) {
		conn := value.(IConn)
		_ = conn.Close()
	})
}

func (m *Multiplexers) newTempConn() (IConn, error) {
	// All multiplexers are at limit, create a new one
	ctx := context.Background()
	conn := transport.DialContextWithOps(ctx, m.host, &transport.DialOption{
		MaxIncomingPacket: MaxIncomingPacket,
	})
	multiplexer := mux.NewMultiplexer(ctx, conn)
	vc, err := multiplexer.NewVirtualConn(ctx)
	if err != nil {
		return nil, err
	}
	idx := m.connId()
	vConn := wrapConn(vc, idx, func() {
		multiplexer.Close()
	})

	if err = m.tempMap.Store(idx, vConn); err != nil {
		return nil, err
	}
	return vConn, nil
}

func (m *Multiplexers) connId() int64 {
	return m.connIdx.Add(1)
}
