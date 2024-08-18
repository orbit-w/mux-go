package multiplexers

import (
	"context"
	"fmt"
	pq "github.com/orbit-w/meteor/bases/container/priority_queue"
	"github.com/orbit-w/mux-go"
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
   @File: multiplexers_test
   @2024 8月 周日 10:18
*/

var (
	testOnce = new(sync.Once)
)

// TestMultiplexers_Close tests the Close method of the Multiplexers struct.
// It sets up a server and initializes a Multiplexers instance with multiple multiplexers.
// The test spawns multiple goroutines to simulate concurrent virtual connections.
// After a short delay, it calls the Close method to close all multiplexers and virtual connections.
// The test ensures that all connections are properly closed and counts the number of successful and received connections.
//
// TestMultiplexers_Close 测试 Multiplexers 结构体的 Close 方法。
// 它设置一个服务器并初始化一个包含多个多路复用器的 Multiplexers 实例。
// 测试生成多个 goroutine 来模拟并发虚拟连接。
// 短暂延迟后，调用 Close 方法关闭所有多路复用器和虚拟连接。
// 测试确保所有连接都正确关闭，并统计成功和接收的连接数量。
func TestMultiplexers_Close(t *testing.T) {
	host := "127.0.0.1:8888"
	Serve(t, host, false)
	muxs := NewMultiplexers("127.0.0.1:8888", 5, 10)
	wg := sync.WaitGroup{}
	wg2 := sync.WaitGroup{}
	bc := context.Background()
	count := atomic.Uint32{}
	recvCount := atomic.Uint32{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			for j := 0; j < 10000; j++ {
				conn, err := muxs.Dial()
				if err == nil {
					count.Add(1)
					wg2.Add(1)
					go func() {
						for {
							_, err := conn.Recv(bc)
							if err != nil {
								recvCount.Add(1)
								break
							}
						}
						wg2.Done()
					}()
				}
			}
			wg.Done()
		}()
	}

	wg.Add(1)
	go func() {
		time.Sleep(time.Millisecond * 100)
		muxs.Close()
		wg.Done()
	}()

	wg.Wait()
	wg2.Wait()
	fmt.Println(count.Load(), recvCount.Load())
	assert.Equal(t, count.Load(), recvCount.Load())
}

// Test_PQ tests the Close method of the connection wrapper.
// It initializes a priority queue and a read-write mutex, then spawns multiple goroutines to simulate concurrent access.
// Each goroutine creates a wrapped connection, updates the priority in the queue, and then closes the connection.
// The test ensures that the priority of the item in the queue is correctly updated and that the final priority is zero.
func Test_PQ(t *testing.T) {
	queue := pq.New[int64, mux.IMux, int]()
	rw := sync.RWMutex{}
	wg := sync.WaitGroup{}
	idx := int64(10000)
	queue.Push(idx, nil, 0)
	for i := 0; i < 10000; i++ {
		wg.Add(1)
		go func() {
			conn := wrapConn(nil, idx, func() {
				rw.Lock()
				defer rw.Unlock()
				queue.UpdatePriorityOp(idx, decrPriority)
			})
			rw.Lock()
			queue.UpdatePriorityOp(idx, incrPriority)
			rw.Unlock()
			wg.Add(1)
			go func() {
				_ = conn.Close()
				wg.Done()
			}()
			wg.Done()
		}()
	}
	wg.Wait()

	fmt.Println("done")
	item, ok := queue.Get(idx)
	assert.True(t, ok)
	assert.Equal(t, item.Priority, 0)
}

func Serve(t assert.TestingT, host string, print bool) {
	testOnce.Do(func() {
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
