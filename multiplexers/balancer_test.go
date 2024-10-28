package multiplexers

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"sync"
	"testing"
	"time"
)

func TestBalancer_Decr(t *testing.T) {
	ba := NewBalancer(10)
	ba.Incr(1)
	fmt.Println(ba.Load(1))
	ba.Decr(1)
	fmt.Println(ba.Load(1))
}

func TestBalancer_ConcurrentTesting(t *testing.T) {
	ba := NewBalancer(10)
	wg := sync.WaitGroup{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			for j := 0; j < 1000; j++ {
				idx := r.Intn(10)
				ba.Incr(idx)
				ba.Decr(idx)
			}
			wg.Done()
		}()
	}

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			for j := 0; j < 1000; j++ {
				_ = ba.Next()
			}
			wg.Done()
		}()
	}

	wg.Wait()
	for i := range ba.counter {
		assert.Equal(t, uint64(0), ba.Load(i))
	}
}
