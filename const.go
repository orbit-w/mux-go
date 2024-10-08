package mux

import "time"

/*
   @Author: orbit-w
   @File: const
   @2024 4月 周一 22:09
*/

const (
	MaxIncomingPacket = 262144 //256kb
	ReadTimeout       = time.Second * 60
	WriteTimeout      = time.Second * 5

	DialTimeout = time.Second * 15
)

const (
	StateMuxRunning = iota
	StateMuxStopped
)

const (
	ConnActive uint32 = iota
	ConnWriteDone
)

const (
	MessageRaw = iota
	MessageStart
	MessageFin
)
