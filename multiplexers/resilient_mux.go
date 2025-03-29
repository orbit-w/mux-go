package multiplexers

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/orbit-w/meteor/modules/net/transport"
	"github.com/orbit-w/mux-go"
)

// ResilientMux 管理 Multiplexer 的生命周期，当连接断开时自动重连
// 实现 mux.IMux 接口，可以替代 Multiplexer 使用
type ResilientMux struct {
	mu           sync.RWMutex
	mux          mux.IMux
	host         string
	maxConns     int
	ctx          context.Context
	cancel       context.CancelFunc
	state        atomic.Int32
	reconnecting atomic.Bool
}

const (
	ResilientMuxStateNormal = iota
	ResilientMuxStateReconnecting
	ResilientMuxStateClosed
)

// CreateResilientMux 快速创建一个连接到指定主机的 ResilientMux 实例
// 如果连接失败，会启动后台重连逻辑
// 适合需要"无感知"自动重连的场景，服务启动就连接并持续保持
func CreateResilientMux(host string, maxConns int) *ResilientMux {
	ctx, cancel := context.WithCancel(context.Background())
	rm := &ResilientMux{
		host:     host,
		maxConns: maxConns,
		ctx:      ctx,
		cancel:   cancel,
	}
	// 设置初始状态为正常
	rm.state.Store(ResilientMuxStateNormal)

	// 尝试初始连接，如果失败会自动触发重连
	err := rm.Dial()
	if err != nil {
		// 连接失败，模拟断线回调触发重连
		rm.onDisconnected(err)
	}

	return rm
}

// NewResilientMux 创建一个新的 ResilientMux 实例，并连接到指定主机
// 适合需要明确知道初始连接状态的场景，可以根据错误做特定处理
func NewResilientMux(ctx context.Context, host string, maxConns int) (*ResilientMux, error) {
	ctx, cancel := context.WithCancel(ctx)
	rm := &ResilientMux{
		host:     host,
		maxConns: maxConns,
		ctx:      ctx,
		cancel:   cancel,
	}
	// 设置初始状态为正常
	rm.state.Store(ResilientMuxStateNormal)

	// 初始化连接
	err := rm.Dial()
	if err != nil {
		cancel() // 如果连接失败，取消 context
		return nil, err
	}

	return rm, nil
}

// Dial 连接到指定的主机，创建 Multiplexer
func (rm *ResilientMux) Dial() error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if rm.state.Load() == ResilientMuxStateClosed {
		return mux.ErrCancel
	}

	// 创建新连接
	conn := transport.DialWithOps(rm.ctx, rm.host, transport.WithMaxIncomingPacket(mux.MaxIncomingPacket))

	// 创建新的 Multiplexer，并注册断开连接的回调函数
	rm.mux = mux.NewMultiplexer(rm.ctx, conn, mux.WithDisconnectedCallback(rm.onDisconnected), mux.WithMaxVirtualConns(rm.maxConns))

	// 连接成功，确保状态为正常
	rm.state.Store(ResilientMuxStateNormal)
	return nil
}

// NewVirtualConn 实现 mux.IMux 接口
// 创建一个新的虚拟连接
func (rm *ResilientMux) NewVirtualConn(ctx context.Context) (mux.IConn, error) {
	rm.mu.RLock()
	multiplexer := rm.mux
	rm.mu.RUnlock()

	if multiplexer == nil {
		return nil, mux.ErrCancel
	}

	return multiplexer.NewVirtualConn(ctx)
}

// Close 实现 mux.IMux 接口
// 关闭 ResilientMux 和其管理的资源
func (rm *ResilientMux) Close() {
	// 尝试将状态从正常或重连中变为关闭
	if rm.state.Swap(ResilientMuxStateClosed) == ResilientMuxStateClosed {
		return // 已经关闭，无需重复操作
	}

	rm.cancel()

	rm.mu.Lock()
	if rm.mux != nil {
		rm.mux.Close()
		rm.mux = nil
	}
	rm.mu.Unlock()
}

// onDisconnected 处理连接断开事件
func (rm *ResilientMux) onDisconnected(err error) {
	// 如果 ResilientMux 已关闭或正在重连，则跳过
	if rm.state.Load() == ResilientMuxStateClosed {
		return
	}

	// 标记正在重连
	if !rm.reconnecting.CompareAndSwap(false, true) {
		return
	}

	// 再次检查状态，确保在获取重连标记后状态没有变化
	if rm.state.Load() == ResilientMuxStateClosed {
		rm.reconnecting.Store(false) // 释放标记
		return
	}

	go func() {
		defer rm.reconnecting.Store(false)

		// 标记状态为重连中
		rm.state.Store(ResilientMuxStateReconnecting)

		// 记录旧的 Multiplexer，以便之后关闭
		rm.mu.Lock()
		oldMux := rm.mux
		rm.mux = nil
		rm.mu.Unlock()

		// 关闭旧的连接
		if oldMux != nil {
			oldMux.Close()
		}

		// 重连逻辑，使用指数退避策略
		backoff := 100 * time.Millisecond
		maxBackoff := 5 * time.Second

		for rm.state.Load() != ResilientMuxStateClosed {
			// 尝试重新连接
			err := rm.Dial()
			if err == nil {
				// 重连成功，恢复正常状态
				rm.state.Store(ResilientMuxStateNormal)
				return
			}

			// 如果 context 已取消，退出重连
			select {
			case <-rm.ctx.Done():
				return
			case <-time.After(backoff):
				// 增加退避时间，但不超过最大值
				backoff *= 2
				if backoff > maxBackoff {
					backoff = maxBackoff
				}
			}
		}
	}()
}

// GetState 获取当前 ResilientMux 的状态
func (rm *ResilientMux) GetState() int32 {
	return rm.state.Load()
}

func (rm *ResilientMux) IsConnected() bool {
	rm.mu.RLock()
	muxAvailable := rm.mux != nil
	currentState := rm.state.Load()
	rm.mu.RUnlock()

	return muxAvailable && currentState == ResilientMuxStateNormal
}
