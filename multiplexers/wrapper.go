package multiplexers

import "github.com/orbit-w/mux-go"

/*
   @Author: orbit-w
   @File: wrapper
   @2024 8月 周日 18:15
*/

type ConnWrapper struct {
	mux.IConn
	cancel func()
}

func (c *ConnWrapper) Close() {
	if c.IConn != nil {
		_ = c.IConn.CloseSend()
	}

	if c.cancel != nil {
		c.cancel()
	}
}
