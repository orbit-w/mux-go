package mux

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

/*
   @Author: orbit-w
   @File: conns_test
   @2024 8月 周日 16:47
*/

func TestVirtualConns_ConcurrentTesting(t *testing.T) {
	mgr := newConns(10)
	wg := sync.WaitGroup{}
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			for j := 0; j < 500; j++ {
				_ = mgr.Reg(mgr.Id(), &VirtualConn{})
			}
			wg.Done()
		}()
	}

	wg.Wait()
	id := mgr.Id()
	assert.Error(t, mgr.Reg(id, &VirtualConn{}))
	fmt.Println(mgr.Len())

	mgr.OnClose(func(stream *VirtualConn) {

	})

	mgr.OnClose(func(stream *VirtualConn) {

	})

	assert.ErrorAs(t, mgr.Reg(mgr.Id(), &VirtualConn{}), &ErrCancel)

}

func TestVirtualConns_Del(t *testing.T) {
	mgr := newConns(10)
	id := mgr.Id()
	assert.NoError(t, mgr.Reg(id, &VirtualConn{}))
	mgr.Del(id)
	assert.False(t, mgr.Exist(id))
}
