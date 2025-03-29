package mux

/*
   @Author: orbit-w
   @File: config
   @2024 7月 周日 19:23
*/

type Opt func(config *MuxClientConfig)

type MuxClientConfig struct {
	MaxVirtualConns      int //最大流数
	DisconnectedCallback func(err error)
}

const (
	maxVirtualConns = 200
)

func DefaultClientConfig() *MuxClientConfig {
	return &MuxClientConfig{
		MaxVirtualConns: maxVirtualConns,
	}
}

func WithDisconnectedCallback(callback func(err error)) Opt {
	return func(config *MuxClientConfig) {
		config.DisconnectedCallback = callback
	}
}

func WithMaxVirtualConns(maxVirtualConns int) Opt {
	return func(config *MuxClientConfig) {
		config.MaxVirtualConns = maxVirtualConns
	}
}

func parseConfig(config *MuxClientConfig, opts ...Opt) {
	for _, opt := range opts {
		opt(config)
	}

	if config.MaxVirtualConns <= 0 {
		config.MaxVirtualConns = maxVirtualConns
	}
}
