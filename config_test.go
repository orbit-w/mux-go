package mux

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

/*
   @Author: orbit-w
   @File: config_test
   @2024 8月 周日 17:00
*/

func Test_parseConfig(t *testing.T) {
	var conf *MuxServerConfig
	parseServerConfig(&conf)
	fmt.Println(conf.toTransportConfig())

	conf.ReadTimeout = -time.Second * 5
	conf.WriteTimeout = -time.Second * 5
	conf.DialTimeout = -time.Second * 5
	conf.MaxIncomingPacket = 0
	parseServerConfig(&conf)
	fmt.Println(conf.MaxIncomingPacket)
	fmt.Println(conf.ReadTimeout)
	fmt.Println(conf.WriteTimeout)

	c := NewClientConfig(-200)
	c = parseConfig(c)
	fmt.Println(c.MaxVirtualConns)
}

func Test_misc(t *testing.T) {
	err := errors.New("context canceled")
	assert.True(t, IsErrCanceled(err))
	fmt.Println(NewStreamBufSetErr(err).Error())
	fmt.Println(NewDecodeErr(err).Error())
}
