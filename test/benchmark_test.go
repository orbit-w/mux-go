package test

import (
	"context"
	"github.com/orbit-w/meteor/modules/net/transport"
	"github.com/orbit-w/mux-go"
	"github.com/stretchr/testify/assert"
	"io"
	"log"
	"sync"
	"testing"
	"time"
)

/*
   @Author: orbit-w
   @File: benchmark_test
   @2024 8月 周四 10:05
*/

var (
	once = new(sync.Once)
)

func Benchmark_StreamSend_Test(b *testing.B) {
	host := "127.0.0.1:6800"
	Serve(b, host, false)

	conn := transport.DialContextWithOps(context.Background(), host)
	multiplexer := mux.NewMultiplexer(context.Background(), conn)

	vc, err := multiplexer.NewVirtualConn(context.Background())
	assert.NoError(b, err)

	b.Run("BenchmarkStreamSend", func(b *testing.B) {
		b.ResetTimer()
		b.StartTimer()
		defer b.StopTimer()
		for i := 0; i < b.N; i++ {
			_ = vc.Send([]byte{1})
		}
	})

	b.StopTimer()
	time.Sleep(time.Second * 5)
	//_ = conn.Close()
}

func Benchmark_ConcurrencyStreamSend_Test(b *testing.B) {
	host := "127.0.0.1:6800"
	Serve(b, host, false)

	b.RunParallel(func(pb *testing.PB) {
		conn := transport.DialContextWithOps(context.Background(), host)
		multiplexer := mux.NewMultiplexer(context.Background(), conn)
		//defer func() {
		//	multiplexer.Close()
		//}()
		vc, err := multiplexer.NewVirtualConn(context.Background())
		assert.NoError(b, err)
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
			}
		}()

		b.ResetTimer()
		b.StartTimer()
		b.ReportAllocs()

		for pb.Next() {
			_ = vc.Send([]byte{1})
		}
	})

	b.StopTimer()
	time.Sleep(time.Second * 5)
	//_ = conn.Close()
}
