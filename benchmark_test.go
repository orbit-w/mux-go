package mux

import (
	"context"
	"github.com/orbit-w/meteor/modules/net/transport"
	"github.com/stretchr/testify/assert"
	"io"
	"log"
	"sync"
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

//func Benchmark_ConcurrencyStreamSend_Test(b *testing.B) {
//	host := "127.0.0.1:6800"
//	Serve(b, host, false)
//
//	b.RunParallel(func(pb *testing.PB) {
//		conn := transport.DialContextWithOps(context.Background(), host)
//		multiplexer := NewMultiplexer(context.Background(), conn)
//		//defer func() {
//		//	multiplexer.Close()
//		//}()
//		vc, err := multiplexer.NewVirtualConn(context.Background())
//		assert.NoError(b, err)
//		go func() {
//			for {
//				in, err := vc.Recv(context.Background())
//				if err != nil {
//					if err == io.EOF {
//						log.Println("client conn read complete...")
//					} else {
//						log.Println("conn read cli vc failed: ", err.Error())
//					}
//					break
//				}
//				if string(in) != "hello, client" {
//					panic("invalid message")
//				}
//			}
//		}()
//
//		b.ResetTimer()
//		b.StartTimer()
//		b.ReportAllocs()
//
//		for pb.Next() {
//			_ = vc.Send([]byte{1})
//		}
//	})
//
//	b.StopTimer()
//	//_ = conn.Close()
//}

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
