package multiplexers

import (
	"context"
	"errors"
	"sync/atomic"

	"github.com/orbit-w/meteor/modules/net/transport"
	"github.com/orbit-w/mux-go"
)

/*
   @Author: orbit-w
   @File: multiplexers
   @2024 8月 周四 23:47
*/

type Multiplexers struct {
	state        atomic.Int32
	connIdx      atomic.Int64
	muxCount     int //mux的数量
	maxConns     int //mux对应的最大虚拟连接数
	host         string
	balancer     *Balancer
	multiplexers []mux.IMux
	tempConns    *connCache
}

func New(host string, conf *Config) *Multiplexers {
	if conf == nil {
		conf = DefaultConfig()
	}

	m := &Multiplexers{
		host:      host,
		muxCount:  conf.MuxCount,
		maxConns:  conf.MuxMaxConns,
		tempConns: newConnCache(),
		balancer:  NewBalancer(conf.MuxCount),
	}

	m.init()
	return m
}

func NewWithDefaultConf(host string) *Multiplexers {
	conf := DefaultConfig()
	m := &Multiplexers{
		host:      host,
		muxCount:  conf.MuxCount,
		maxConns:  conf.MuxMaxConns,
		tempConns: newConnCache(),
		balancer:  NewBalancer(conf.MuxCount),
	}

	m.init()
	return m
}

func (m *Multiplexers) init() {
	for i := 0; i < m.muxCount; i++ {
		ctx := context.Background()
		conn := transport.DialWithOps(ctx, m.host, transport.WithMaxIncomingPacket(MaxIncomingPacket))
		multiplexer := mux.NewMultiplexer(ctx, conn, mux.WithMaxVirtualConns(m.maxConns))
		m.multiplexers = append(m.multiplexers, multiplexer)
	}
}

func (m *Multiplexers) State() int32 {
	return m.state.Load()
}

// Dial 方法严格按照绑定的最小虚拟连接数优先选择多路复用器来创建虚拟连接
func (m *Multiplexers) Dial(ctx context.Context) (IConn, error) {
	index := m.balancer.Next()

	multiplexer := m.multiplexers[index]
	vc, err := multiplexer.NewVirtualConn(ctx)
	if err != nil {
		if !errors.Is(err, mux.ErrVirtualConnUpLimit) {
			return nil, err
		}
		return m.newTempConn()
	}

	m.balancer.Incr(index)
	conn := wrapConn(vc, func() {
		m.balancer.Decr(index)
	})

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

	for i := range m.multiplexers {
		multiplexer := m.multiplexers[i]
		multiplexer.Close()
	}

	// Iterate through the temporary map, closing each virtual connection
	m.tempConns.OnClose(func(conn IConn) {
		_ = conn.Close()
	})
}

func (m *Multiplexers) newTempConn() (IConn, error) {
	// All multiplexers are at limit, create a new one
	ctx := context.Background()
	conn := transport.DialWithOps(ctx, m.host, transport.WithMaxIncomingPacket(MaxIncomingPacket))
	multiplexer := mux.NewMultiplexer(ctx, conn)
	vc, err := multiplexer.NewVirtualConn(ctx)
	if err != nil {
		multiplexer.Close()
		return nil, err
	}
	idx := m.connId()
	vConn := wrapConn(vc, func() {
		m.tempConns.Delete(idx)
		multiplexer.Close()
	})

	if err = m.tempConns.Store(idx, vConn); err != nil {
		multiplexer.Close()
		return nil, err
	}
	return vConn, nil
}

func (m *Multiplexers) connId() int64 {
	return m.connIdx.Add(1)
}
