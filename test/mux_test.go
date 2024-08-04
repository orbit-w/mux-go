package test

import (
	"context"
	"fmt"
	"github.com/orbit-w/meteor/modules/net/transport"
	"github.com/orbit-w/mux-go"
	"github.com/stretchr/testify/assert"
	"io"
	"log"
	"sync/atomic"
	"testing"
	"time"
)

/*
   @Author: orbit-w
   @File: mux_test
   @2024 8月 周四 10:16
*/

func Test_MuxSend(t *testing.T) {
	host := "127.0.0.1:6800"
	Serve(t, host, true)

	conn := transport.DialContextWithOps(context.Background(), host)
	multiplexer := mux.NewMultiplexer(context.Background(), conn)

	vc, err := multiplexer.NewVirtualConn(context.Background())
	assert.NoError(t, err)

	go func() {
		for {
			in, err := vc.Recv(context.Background())
			if err != nil {
				log.Println("conn read cli vc failed: ", err.Error())
				break
			}
			fmt.Println(string(in))
		}
	}()

	err = vc.Send([]byte("hello, server"))
	assert.NoError(t, err)
	err = vc.Send([]byte("hello, server"))
	assert.NoError(t, err)
	err = vc.CloseSend()
	assert.NoError(t, err)
	time.Sleep(time.Second)
}

func Test_GracefulClose(t *testing.T) {
	host := "127.0.0.1:6800"
	Serve(t, host, true)
	conn := transport.DialContextWithOps(context.Background(), host)
	multiplexer := mux.NewMultiplexer(context.Background(), conn)

	vc, err := multiplexer.NewVirtualConn(context.Background())
	assert.NoError(t, err)

	go func() {
		for {
			in, err := vc.Recv(context.Background())
			if err != nil {
				log.Println("conn read cli vc failed: ", err.Error())
				break
			}
			fmt.Println(string(in))
		}
	}()
	err = vc.CloseSend()
	assert.NoError(t, err)
	multiplexer.Close()
	time.Sleep(time.Minute * 5)
}

func Test_CloseMux(t *testing.T) {
	host := "127.0.0.1:6800"
	Serve(t, host, true)
	conn := transport.DialContextWithOps(context.Background(), host)
	multiplexer := mux.NewMultiplexer(context.Background(), conn)

	vc, err := multiplexer.NewVirtualConn(context.Background())
	assert.NoError(t, err)

	go func() {
		for {
			in, err := vc.Recv(context.Background())
			if err != nil {
				log.Println("conn read cli vc failed: ", err.Error())
				break
			}
			fmt.Println(string(in))
		}
	}()
	time.Sleep(time.Second)
	multiplexer.Close()
	time.Sleep(time.Second * 5)
}

func Test_BatchSend(t *testing.T) {
	host := "127.0.0.1:6800"
	Serve(t, host, false)
	conn := transport.DialContextWithOps(context.Background(), host)
	multiplexer := mux.NewMultiplexer(context.Background(), conn)

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

func Serve(t assert.TestingT, host string, print bool) {
	once.Do(func() {
		server := new(mux.Server)
		err := server.Serve(host, func(conn mux.IServerConn) error {
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
		})
		assert.NoError(t, err)
	})
}
