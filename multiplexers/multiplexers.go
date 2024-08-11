package multiplexers

import (
	"github.com/orbit-w/mux-go"
	"sync"
)

/*
   @Author: orbit-w
   @File: multiplexers
   @2024 8月 周四 23:47
*/

type Multiplexers struct {
	mu         sync.RWMutex
	size       int //mux的数量
	connsCount int //mux对应的最大虚拟连接数
	cache      []*mux.Multiplexer
}
