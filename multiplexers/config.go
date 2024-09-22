package multiplexers

/*
   @Author: orbit-w
   @File: config
   @2024 9月 周日 15:41
*/

type Config struct {
	MuxMaxConns int //每个mux最大虚拟连接数
	MuxCount    int //常驻mux数量
}

func DefaultConfig() Config {
	return Config{
		MuxMaxConns: MuxMaxConns,
		MuxCount:    MuxCount,
	}
}
