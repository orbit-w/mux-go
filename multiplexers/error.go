package multiplexers

import "errors"

/*
   @Author: orbit-w
   @File: error
   @2024 8月 周六 01:12
*/

var (
	ErrMultiplexersStopped     = errors.New("multiplexers is stopped")
	ErrNoAvailableMultiplexers = errors.New("no available multiplexers")
)
