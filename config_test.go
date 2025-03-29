package mux

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

/*
   @Author: orbit-w
   @File: config_test
   @2024 8月 周日 17:00
*/

func Test_parseConfig(t *testing.T) {
	var conf *MuxServerConfig
	buildServerConfig(&conf)
	fmt.Println(conf.toTransportConfig())

	conf.ReadTimeout = -time.Second * 5
	conf.WriteTimeout = -time.Second * 5
	conf.DialTimeout = -time.Second * 5
	conf.MaxIncomingPacket = 0
	buildServerConfig(&conf)
	fmt.Println(conf.MaxIncomingPacket)
	fmt.Println(conf.ReadTimeout)
	fmt.Println(conf.WriteTimeout)

	c := DefaultClientConfig()
	WithMaxVirtualConns(-200)(c)
	parseConfig(c)
	fmt.Println(c.MaxVirtualConns)
}

func Test_misc(t *testing.T) {
	err := errors.New("context canceled")
	assert.True(t, IsErrCanceled(err))
	fmt.Println(newStreamBufSetErr(err).Error())
	fmt.Println(newDecodeErr(err).Error())
}
