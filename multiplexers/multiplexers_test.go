package multiplexers

import (
	"fmt"
	pq "github.com/orbit-w/meteor/bases/container/priority_queue"
	"github.com/orbit-w/mux-go"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

/*
   @Author: orbit-w
   @File: multiplexers_test
   @2024 8月 周日 10:18
*/

func TestConnWrapper_Close(t *testing.T) {
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
