package mux

import (
	"context"
	"github.com/orbit-w/meteor/modules/net/network"
	"github.com/orbit-w/meteor/modules/net/transport"
	"time"
)

/*
   @Author: orbit-w
   @File: serve
   @2024 4月 周日 11:20
*/

type Server struct {
	server transport.IServer
	ctx    context.Context
	cancel context.CancelFunc
	handle func(conn IServerConn) error
}

// Serve 以默认配置启动服务
// 业务侧只需要break/return即可，不需要调用 IServerConn.Close()，系统会自动关闭虚拟链接
func (s *Server) Serve(addr string, handle func(conn IServerConn) error) error {
	conf := DefaultServerConfig()
	return s.ServeByConfig(addr, handle, conf)
}

// ServeByConfig 以指定配置启动服务
// 业务侧只需要break/return即可，不需要调用 IServerConn.Close()，系统会自动关闭虚拟链接
func (s *Server) ServeByConfig(addr string, handle func(conn IServerConn) error, conf *MuxServerConfig) error {
	s.handle = handle
	ctx, cancel := context.WithCancel(context.Background())
	s.ctx = ctx
	s.cancel = cancel
	parseServerConfig(&conf)

	tConf := conf.toTransportConfig()
	ts, err := transport.ServeByConfig("tcp", addr, func(conn transport.IConn) {
		mux := newMultiplexer(s.ctx, conn, false, s)
		mux.recvLoop()
	}, tConf)
	if err != nil {
		return err
	}
	s.server = ts
	return nil
}

func (s *Server) Stop() error {
	if s.server != nil {
		return s.server.Stop()
	}
	return nil
}

type MuxServerConfig struct {
	MaxIncomingPacket uint32
	IsGzip            bool
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	DialTimeout       time.Duration
}

func (conf *MuxServerConfig) toTransportConfig() *transport.Config {
	return &transport.Config{
		MaxIncomingPacket: conf.MaxIncomingPacket,
		IsGzip:            conf.IsGzip,
		ReadTimeout:       conf.ReadTimeout,
		WriteTimeout:      conf.WriteTimeout,
	}
}

func parseServerConfig(conf **MuxServerConfig) {
	if *conf == nil {
		*conf = DefaultServerConfig()
	}

	if (*conf).ReadTimeout <= 0 {
		(*conf).ReadTimeout = ReadTimeout
	}

	if (*conf).WriteTimeout <= 0 {
		(*conf).WriteTimeout = WriteTimeout
	}

	if (*conf).MaxIncomingPacket == 0 {
		(*conf).MaxIncomingPacket = network.MaxIncomingPacket
	}
}

func DefaultServerConfig() *MuxServerConfig {
	return &MuxServerConfig{
		MaxIncomingPacket: network.MaxIncomingPacket,
		IsGzip:            false,
		ReadTimeout:       ReadTimeout,
		DialTimeout:       DialTimeout,
		WriteTimeout:      WriteTimeout,
	}
}
