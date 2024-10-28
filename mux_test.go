package mux

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/orbit-w/meteor/modules/net/transport"
	"github.com/orbit-w/mux-go/metadata"
	"github.com/stretchr/testify/assert"
	"io"
	"log"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

/*
   @Author: orbit-w
   @File: mux_test
   @2024 8月 周四 10:16
*/

const (
	Dev  = "DEV"
	Prod = "PROD"
)

func Test_MuxSend(t *testing.T) {
	host := "127.0.0.1:6800"
	Serve(t, host, true, Dev)
	defer server.Stop()

	_, vc := ClientTest(t, host, true)

	err := vc.Send([]byte("hello, server"))
	assert.NoError(t, err)
	err = vc.Send([]byte("hello, server"))
	assert.NoError(t, err)
	err = vc.CloseSend()
	assert.NoError(t, err)
	time.Sleep(time.Second)
}

// Mux 优雅退出测试
func Test_GracefulClose(t *testing.T) {
	host := "127.0.0.1:6800"
	Serve(t, host, true, Prod)
	defer server.Stop()
	multiplexer, vc := ClientTest(t, host, true)
	err := vc.CloseSend()
	assert.NoError(t, err)
	multiplexer.Close()
	time.Sleep(time.Second * 5)
}

// Mux Close 关闭测试
func Test_CloseMux(t *testing.T) {
	host := "127.0.0.1:6800"
	Serve(t, host, true, Prod)
	defer server.Stop()
	multiplexer, _ := ClientTest(t, host, true)
	time.Sleep(time.Second)
	multiplexer.Close()
	time.Sleep(time.Second * 5)
}

// Mux 虚拟链接最大数量限制在并发情况下是否生效测试
func Test_MaxVC(t *testing.T) {
	host := "127.0.0.1:6800"
	Serve(t, host, true, Dev)

	conn := transport.DialContextWithOps(context.Background(), host)
	multiplexer := NewMultiplexer(context.Background(), conn, DefaultClientConfig())

	v := atomic.Uint32{}
	wg := sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			for j := 0; j < 100; j++ {
				_, err := multiplexer.NewVirtualConn(context.Background())
				if err == nil {
					v.Add(1)
				}
			}
			wg.Done()
		}()
	}
	wg.Wait()
	fmt.Println(v.Load())
	multiplexer.Close()
}

// Mux virtual conn send and recv test
// 虚拟链接批量发送和接收测试
func Test_VirtualConnBatchSend(t *testing.T) {
	host := "127.0.0.1:6800"
	Serve(t, host, false, Dev)
	conn := transport.DialContextWithOps(context.Background(), host)
	multiplexer := NewMultiplexer(context.Background(), conn)

	vc, err := multiplexer.NewVirtualConn(context.Background())
	assert.NoError(t, err)

	var (
		s     atomic.Uint32
		r     atomic.Uint32
		total = 100000
		ntf   = make(chan struct{}, 1)
	)

	go func() {
		for {
			in, err := vc.Recv(context.Background())
			if err != nil {
				if err == io.EOF {
					log.Println("client conn read complete...")
				} else {
					log.Println("conn read cli vc failed: ", err.Error())
				}
				break
			}
			if string(in) != "hello, client" {
				panic("invalid message")
			}
			r.Add(1)
			if r.Load() == uint32(total) {
				close(ntf)
			}
		}
	}()

	for i := 0; i < total; i++ {
		if err = vc.Send([]byte("hello, server")); err != nil {
			panic(err)
		}
		s.Add(1)
	}

	<-ntf
	assert.Equal(t, s.Load(), r.Load())
	fmt.Println("Exec count: ", s.Load())
	err = vc.CloseSend()
	assert.NoError(t, err)
	time.Sleep(time.Second)
	multiplexer.Close()
}

func Test_Metadata(t *testing.T) {
	host := "127.0.0.1:6800"
	ServeWithHandler(t, host, func(conn IServerConn) error {
		md, _ := metadata.FromIncomingContext(conn.Context())
		fmt.Println(md)

		for {
			in, err := conn.Recv(context.Background())
			if err != nil {
				if err == io.EOF {
					log.Println("server conn read complete...")
				} else {
					log.Println("conn read server stream failed: ", err.Error())
				}
				break
			}

			fmt.Println(string(in))
			err = conn.Send([]byte("hello, client"))
			assert.NoError(t, err)
		}
		return nil
	})
	conn := transport.DialContextWithOps(context.Background(), host)
	multiplexer := NewMultiplexer(context.Background(), conn)

	id := uuid.New().String()
	fmt.Println(id)
	ctx := metadata.NewOutContext(context.Background(), map[string]any{
		"Uuid":      id,
		"AccountId": "1675987",
	})
	vc, err := multiplexer.NewVirtualConn(ctx)
	assert.NoError(t, err)

	go func() {
		for {
			in, err := vc.Recv(context.Background())
			if err != nil {
				if err == io.EOF {
					log.Println("client conn read complete...")
				} else {
					log.Println("conn read cli vc failed: ", err.Error())
				}
				break
			}
			fmt.Println(string(in))
		}
	}()

	time.Sleep(time.Second * 5)
	multiplexer.Close()
}

func ClientTest(t assert.TestingT, host string, print bool) (multiplexer IMux, vc IConn) {
	conn := transport.DialContextWithOps(context.Background(), host)
	multiplexer = NewMultiplexer(context.Background(), conn)

	var (
		err error
	)
	vc, err = multiplexer.NewVirtualConn(context.Background())
	assert.NoError(t, err)

	go func() {
		for {
			in, err := vc.Recv(context.Background())
			if err != nil {
				if err == io.EOF {
					log.Println("client conn read complete...")
				} else {
					log.Println("conn read cli vc failed: ", err.Error())
				}
				break
			}
			if print {
				fmt.Println(string(in))
			}
		}
	}()
	return
}

func Serve(t assert.TestingT, host string, print bool, stage string) {
	recvHandler := func(conn IServerConn) error {
		for {
			in, err := conn.Recv(context.Background())
			if err != nil {
				if err == io.EOF {
					log.Println("server conn read complete...")
				} else {
					log.Println("conn read server stream failed: ", err.Error())
				}
				break
			}
			if print {
				fmt.Println(string(in))
			}
			err = conn.Send([]byte("hello, client"))
			assert.NoError(t, err)
		}
		return nil
	}

	var muxServerConfig *MuxServerConfig

	switch stage {
	case "DEV":
		muxServerConfig = DevelopmentServerConfig()
	case "PROD":
		muxServerConfig = ProductionServerConfig()
	default:
		muxServerConfig = DefaultServerConfig()
	}

	testOnce.Do(func() {
		server = new(Server)
		err := server.ServeByConfig(host, recvHandler, muxServerConfig)
		assert.NoError(t, err)
	})
}

func ServeWithHandler(t assert.TestingT, host string, handler func(conn IServerConn) error) {
	testOnce.Do(func() {
		server = new(Server)
		err := server.ServeByConfig(host, handler, DevelopmentServerConfig())
		assert.NoError(t, err)
	})
}

func serveWithHandler(t assert.TestingT, stage string, recvHandler func(conn IServerConn) error) *Server {
	host := "localhost:0"
	var muxServerConfig *MuxServerConfig

	switch stage {
	case "DEV":
		muxServerConfig = DevelopmentServerConfig()
	case "PROD":
		muxServerConfig = ProductionServerConfig()
	default:
		muxServerConfig = DefaultServerConfig()
	}
	s := new(Server)
	err := s.ServeByConfig(host, recvHandler, muxServerConfig)
	assert.NoError(t, err)
	return s
}

func Test_Serve(t *testing.T) {
	s := new(Server)
	assert.NoError(t, s.Stop())
	assert.NoError(t, s.Serve("127.0.0.1:6900", func(conn IServerConn) error {
		return nil
	}))

	assert.NoError(t, s.Stop())
}
