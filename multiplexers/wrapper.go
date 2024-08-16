package multiplexers

import (
	"context"
	"github.com/orbit-w/mux-go"
	"sync/atomic"
)

/*
   @Author: orbit-w
   @File: wrapper
   @2024 8月 周日 18:15
*/

type IConn interface {
	Send(data []byte) error
	Recv(ctx context.Context) ([]byte, error)
	Close() error
}

type ConnWrapper struct {
	mux.IConn
	state  atomic.Uint32
	idx    int64
	cancel func()
}

func wrapConn(conn mux.IConn, _idx int64, cancel func()) IConn {
	return &ConnWrapper{
		idx:    _idx,
		IConn:  conn,
		cancel: cancel,
	}
}

func (c *ConnWrapper) Close() error {
	if !c.state.CompareAndSwap(StateNormal, StateStopped) {
		return nil
	}

	if c.IConn != nil {
		_ = c.IConn.CloseSend()
	}

	if c.cancel != nil {
		c.cancel()
	}
	return nil
}
