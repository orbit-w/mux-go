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
	state   atomic.Uint32
	closeCB func()
}

func wrapConn(conn mux.IConn, callback func()) IConn {
	return &ConnWrapper{
		IConn:   conn,
		closeCB: callback,
	}
}

func (c *ConnWrapper) Close() error {
	if !c.state.CompareAndSwap(StateNone, StateClosed) {
		return nil
	}

	if c.IConn != nil {
		_ = c.IConn.CloseSend()
	}

	if c.closeCB != nil {
		c.closeCB()
	}
	return nil
}
