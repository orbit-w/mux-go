package multiplexers

import "sync/atomic"

type Balancer struct {
	counter []uint64
}

func NewBalancer(size int) *Balancer {
	return &Balancer{
		counter: make([]uint64, size),
	}
}

func (b *Balancer) Next() int {
	var minV uint64 = 1<<64 - 1
	var idx int
	for i, c := range b.counter {
		if c < minV {
			minV = c
			idx = i
		}
	}
	return idx
}

func (b *Balancer) Incr(idx int) {
	atomic.AddUint64(&b.counter[idx], 1)
}

func (b *Balancer) Decr(idx int) {
	atomic.AddUint64(&b.counter[idx], ^uint64(0))
}

func (b *Balancer) Load(idx int) uint64 {
	return atomic.LoadUint64(&b.counter[idx])
}
