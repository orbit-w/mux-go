package multiplexers

import (
	"context"
	"github.com/orbit-w/mux-go"
	"io"
	"log"
	"sync/atomic"
	"testing"
)

/*
   @Author: orbit-w
   @File: multiplexers_benchmark_test
   @2024 8月 周日 19:15
*/

func Benchmark_ConcurrencyStreamSend_Test(b *testing.B) {
	var (
		//total    = uint64(128 * 1024 * b.N)
		count = atomic.Uint64{}
		buf   = make([]byte, 128*1024)
		host  = "127.0.0.1:8888"
	)

	ServeWithHandler(b, host, Prod, func(conn mux.IServerConn) error {
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
			count.Add(uint64(len(in)))
		}
		return nil
	})

	// Create a new multiplexer with default configuration
	muxs := NewWithDefaultConf(host)

	b.SetBytes(128 * 1024)
	b.ResetTimer()
	b.ReportAllocs()

	// Run the benchmark in parallel
	b.RunParallel(func(pb *testing.PB) {
		conn, err := muxs.Dial()
		if err != nil {
			b.Error(err)
			return
		}

		for pb.Next() {
			if err = conn.Send(buf); err != nil {
				b.Error(err)
				return
			}
		}
	})

}
