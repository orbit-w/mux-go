package mux

import (
	"context"
	"fmt"
	"github.com/orbit-w/meteor/modules/net/transport"
	"github.com/stretchr/testify/assert"
	"io"
	"log"
	"sync"
	"sync/atomic"
	"testing"
)

/*
   @Author: orbit-w
   @File: benchmark_test
   @2024 8月 周四 10:05
*/

var (
	testOnce = new(sync.Once)
	server   *Server
)

func BenchmarkConnMux(b *testing.B) {
	var (
		count int
		host  = "127.0.0.1:6800"
		wg    = sync.WaitGroup{}
		buf   = make([]byte, 128*1024)
		h     = func(conn IServerConn) error {
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
				count += len(in)
				if count >= 128*1024*b.N {
					wg.Done()
					break
				}
			}
			return nil
		}
	)

	ServeWithHandler(b, host, h)
	server.SetHandler(h)

	conn := transport.DialContextWithOps(context.Background(), host, &transport.DialOption{
		MaxIncomingPacket: MaxIncomingPacket,
	})
	multiplexer := NewMultiplexer(context.Background(), conn)

	vc, err := multiplexer.NewVirtualConn(context.Background())
	assert.NoError(b, err)
	wg.Add(1)
	b.SetBytes(128 * 1024)
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = vc.Send(buf)
	}
	wg.Wait()
}

func Benchmark_ConcurrencySend_128K_Test(b *testing.B) {
	benchmarkEcho(b, 1024*128, 1)
}

func Benchmark_ConcurrencySend_64K_Test(b *testing.B) {
	benchmarkEcho(b, 65536, 5)
}

func benchmarkEcho(b *testing.B, size, num int) {
	var (
		total    = uint64(size * num * b.N)
		count    = atomic.Uint64{}
		buf      = make([]byte, size)
		complete = make(chan struct{}, 1)
	)

	server := serveWithHandler(b, Prod, func(conn IServerConn) error {
		for {
			in, err := conn.Recv(context.Background())
			if err != nil {
				if !(err == io.EOF || IsErrCanceled(err)) {
					log.Println("conn read server stream failed: ", err.Error())
				}
				break
			}
			if count.Add(uint64(len(in))) >= total {
				select {
				case complete <- struct{}{}:
				default:

				}
			}
		}
		return nil
	})
	defer server.Stop()

	// Create a new multiplexer with default configuration
	fmt.Println("Server Addr: ", server.Addr())
	conn := transport.DialContextWithOps(context.Background(), server.Addr())
	multiplexer := NewMultiplexer(context.Background(), conn)
	conns := make([]IConn, num)
	for i := 0; i < num; i++ {
		vc, err := multiplexer.NewVirtualConn(context.Background())
		if err != nil {
			b.Error(err)
			return
		}
		conns[i] = vc
	}

	defer multiplexer.Close()

	b.ReportAllocs()
	b.SetBytes(int64(size * num))
	b.ResetTimer()

	for i := range conns {
		conn := conns[i]
		go func() {
			for j := 0; j < b.N; j++ {
				if err := conn.Send(buf); err != nil {
					b.Error(err)
					return
				}
			}
		}()
	}

	<-complete
}
