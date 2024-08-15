package mux

/*
   @Author: orbit-w
   @File: config
   @2024 7月 周日 19:23
*/

type MuxClientConfig struct {
	MaxVirtualConns int //最大流数
}

const (
	maxVirtualConns = 200
)

func DefaultClientConfig() MuxClientConfig {
	return MuxClientConfig{
		MaxVirtualConns: maxVirtualConns,
	}
}

func NewClientConfig(maxVirtualConns int) MuxClientConfig {
	return MuxClientConfig{
		MaxVirtualConns: maxVirtualConns,
	}
}

func parseConfig(params ...MuxClientConfig) MuxClientConfig {
	if len(params) == 0 {
		return DefaultClientConfig()
	}

	conf := params[0]
	if conf.MaxVirtualConns <= 0 {
		conf.MaxVirtualConns = maxVirtualConns
	}
	return conf
}
