package mux

import (
	"context"
	"github.com/orbit-w/meteor/modules/mlog"
	"github.com/orbit-w/meteor/modules/net/network"
	"github.com/orbit-w/meteor/modules/net/transport"
	"github.com/spf13/viper"
	"time"
)

/*
   @Author: orbit-w
   @File: serve
   @2024 4月 周日 11:20
*/

type Server struct {
	server     transport.IServer
	ctx        context.Context
	cancel     context.CancelFunc
	handleLoop func(conn IServerConn) error
}

// Serve 以默认配置启动服务
// 业务侧只需要break/return即可，不需要调用 IServerConn.Close()，系统会自动关闭虚拟链接
func (s *Server) Serve(addr string, handleLoop func(conn IServerConn) error) error {
	conf := DefaultServerConfig()
	return s.ServeByConfig(addr, handleLoop, conf)
}

// ServeByConfig 以指定配置启动服务
// 业务侧只需要break/return即可，不需要调用 IServerConn.Close()，系统会自动关闭虚拟链接
func (s *Server) ServeByConfig(addr string, handleLoop func(conn IServerConn) error, conf *MuxServerConfig) error {
	s.handleLoop = handleLoop
	ctx, cancel := context.WithCancel(context.Background())
	s.ctx = ctx
	s.cancel = cancel
	buildServerConfig(&conf)
	conf.parse()

	//根据日志等级/文件路径设置zap日志
	mlog.SetBaseLogger(mlog.NewZapLogger())

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

func (s *Server) Addr() string {
	if s.server != nil {
		return s.server.Addr()
	}
	return ""
}

func (s *Server) SetHandler(handle func(conn IServerConn) error) {
	s.handleLoop = handle
}

type MuxServerConfig struct {
	MaxIncomingPacket uint32
	IsGzip            bool
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	DialTimeout       time.Duration
	LogDir            string
	LogLevel          string
}

func (conf *MuxServerConfig) toTransportConfig() *transport.Config {
	return &transport.Config{
		MaxIncomingPacket: conf.MaxIncomingPacket,
		IsGzip:            conf.IsGzip,
		ReadTimeout:       conf.ReadTimeout,
		WriteTimeout:      conf.WriteTimeout,
	}
}

func (conf *MuxServerConfig) parse() {
	//分析配置，设置viper全剧配置
	if conf.LogDir != "" {
		viper.Set(mlog.FlagLogDir, conf.LogDir)
	}

	viper.Set(mlog.FlagV, conf.LogLevel)
}

func buildServerConfig(conf **MuxServerConfig) {
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

	//默认等级INFO
	if (*conf).LogLevel == "" {
		(*conf).LogLevel = "INFO"
	}

}

const (
	defaultLogDir = "./logs/mux.log"
)

func DefaultServerConfig() *MuxServerConfig {
	return &MuxServerConfig{
		MaxIncomingPacket: MaxIncomingPacket,
		IsGzip:            false,
		ReadTimeout:       ReadTimeout,
		DialTimeout:       DialTimeout,
		WriteTimeout:      WriteTimeout,
		LogLevel:          "INFO",
		LogDir:            defaultLogDir,
	}
}

func ProductionServerConfig() *MuxServerConfig {
	return &MuxServerConfig{
		MaxIncomingPacket: MaxIncomingPacket,
		IsGzip:            false,
		ReadTimeout:       ReadTimeout,
		DialTimeout:       DialTimeout,
		WriteTimeout:      WriteTimeout,
		LogLevel:          "ERROR",
		LogDir:            defaultLogDir,
	}
}

func DevelopmentServerConfig() *MuxServerConfig {
	return &MuxServerConfig{
		MaxIncomingPacket: MaxIncomingPacket,
		IsGzip:            false,
		ReadTimeout:       ReadTimeout,
		DialTimeout:       DialTimeout,
		WriteTimeout:      WriteTimeout,
		LogLevel:          "DEBUG",
	}
}
