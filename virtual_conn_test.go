package mux

import (
	"context"
	"github.com/orbit-w/meteor/modules/net/transport"
	"github.com/stretchr/testify/assert"
	"testing"
)

/*
   @Author: orbit-w
   @File: virtual_conn_test
   @2024 8月 周日 17:06
*/

func TestVirtualConn_Send(t *testing.T) {
	host := "127.0.0.1:6800"
	Serve(t, host, false)
	conn := transport.DialContextWithOps(context.Background(), host)
	multiplexer := NewMultiplexer(context.Background(), conn)

	ivc, err := multiplexer.NewVirtualConn(context.Background())
	assert.NoError(t, err)
	vc := ivc.(*VirtualConn)
	assert.NoError(t, vc.send([]byte("Hello"), true))
	assert.Error(t, vc.send([]byte("Hello"), true))
	assert.Error(t, vc.send([]byte("Hello"), false))
	multiplexer.Close()
}

func TestVirtualConn_CloseSend(t *testing.T) {
	host := "127.0.0.1:6800"
	Serve(t, host, false)
	conn := transport.DialContextWithOps(context.Background(), host)
	multiplexer := NewMultiplexer(context.Background(), conn)

	ivc, err := multiplexer.NewVirtualConn(context.Background())
	assert.NoError(t, err)
	vc := ivc.(*VirtualConn)
	vc.mux.isClient = false
	assert.NoError(t, vc.CloseSend())
}
