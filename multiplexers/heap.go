package multiplexers

import "github.com/orbit-w/mux-go"

/*
   @Author: orbit-w
   @File: heap
   @2024 8月 周一 23:51
*/

type MultiplexerWrapper struct {
	mux          *mux.Multiplexer
	virtualConns int
}

type MultiplexerHeap []*MultiplexerWrapper

func (h MultiplexerHeap) Len() int { return len(h) }
func (h MultiplexerHeap) Less(i, j int) bool {
	return h[i].virtualConns < h[j].virtualConns
}
func (h MultiplexerHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }

func (h *MultiplexerHeap) Push(x interface{}) {
	*h = append(*h, x.(*MultiplexerWrapper))
}

func (h *MultiplexerHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}
