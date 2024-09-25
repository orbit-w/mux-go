package multiplexers

import (
	"context"
	"fmt"
	"github.com/orbit-w/mux-go"
	"io"
	"log"
	"runtime"
	"sync/atomic"
	"testing"
)

/*
   @Author: orbit-w
   @File: multiplexers_benchmark_test
   @2024 8月 周日 19:15
*/

func Benchmark_ConcurrencySend_128K_Test(b *testing.B) {
	benchmarkEcho(b, 1024*128, 5)
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

	server := serveWithHandler(b, Prod, func(conn mux.IServerConn) error {
		for {
			in, err := conn.Recv(context.Background())
			if err != nil {
				if !(err == io.EOF || mux.IsErrCanceled(err)) {
					log.Println("conn read server stream failed: ", err.Error())
				}
				break
			}
			t := count.Add(uint64(len(in)))
			if t >= total {
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
	mus := NewWithDefaultConf(server.Addr())
	conns := make([]IConn, num)
	for i := 0; i < num; i++ {
		conn, err := mus.Dial()
		if err != nil {
			b.Error(err)
			return
		}
		conns[i] = conn
	}

	defer mus.Close()

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
	runtime.GC()
}
