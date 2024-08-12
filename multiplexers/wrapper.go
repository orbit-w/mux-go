package multiplexers

import (
	"context"
	"github.com/orbit-w/mux-go"
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
	cancel func()
}

func newConnWrapper(conn mux.IConn, cancel func()) IConn {
	return &ConnWrapper{
		IConn:  conn,
		cancel: cancel,
	}
}

func (c *ConnWrapper) Close() error {
	if c.IConn != nil {
		_ = c.IConn.CloseSend()
	}

	if c.cancel != nil {
		c.cancel()
	}
	return nil
}
