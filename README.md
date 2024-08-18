
# mux-go

`mux-go` 是一个用于多路复用虚拟连接的 Go 语言库。它允许在单个物理连接上创建多个虚拟连接，从而提高网络通信的效率和并发性能。

[![codecov](https://codecov.io/gh/orbit-w/mux-go/branch/master/graph/badge.svg?token=FnHuKkiGDO)](https://codecov.io/gh/orbit-w/mux-go)

## 特性

- 支持创建和管理多个虚拟连接
- 提供高效的并发处理
- 通过最小化锁的粒度来提高性能
- 支持客户端和服务端模式

## 安装

使用 `go get` 命令安装：

```sh
go get github.com/orbit-w/mux-go
```

## Benchmark
```
oos: darwin
goarch: arm64
pkg: github.com/orbit-w/mux-go
BenchmarkConnMux
BenchmarkConnMux-8   	   43083	     27426 ns/op	4779.16 MB/s	  674156 B/op	       7 allocs/op
PASS
ok  	github.com/orbit-w/mux-go	5.556s
```


## 使用方法

### 创建多路复用器

```go
import (
    "context"
    "github.com/orbit-w/mux-go"
    "github.com/orbit-w/meteor/modules/net/transport"
)

func main() {
    conn := transport.DialContextWithOps(context.Background(), "localhost:8080")
    mux := mux.NewMultiplexer(context.Background(), conn)
    // 使用 mux 进行虚拟连接的创建和管理
}
```

### 创建虚拟连接

```go
vc, err := mux.NewVirtualConn(context.Background())
if err != nil {
    log.Fatalf("Failed to create virtual connection: %v", err)
}
// 使用 vc 进行数据传输
```

### 关闭多路复用器

```go
mux.Close()
```

## 接口说明

### IConn 接口

`IConn` 接口定义了虚拟连接的基本操作：

- `Send(data []byte) error`：发送数据
- `Recv(ctx context.Context) ([]byte, error)`：接收数据
- `CloseSend() error`：关闭发送方向

### Multiplexer 类型

`Multiplexer` 类型用于管理虚拟连接：

- `NewVirtualConn(ctx context.Context) (IConn, error)`：创建新的虚拟连接
- `Close()`：关闭多路复用器


### Server demo ###

```go

package main

import (
	"context"
	"fmt"
	"github.com/orbit-w/mux-go"
	"log"
	"os"
	"os/signal"
	"syscall"
)

/*
	@Author: orbit-w
	@File: main
	@2024 8月 周四 17:19
*/

func main() {
	host := "127.0.0.1:6800"
	server := new(mux.Server)
	recvHandle := func(conn mux.IServerConn) error {
		for {
			in, err := conn.Recv(context.Background())
			if err != nil {
				log.Println("conn read stream failed: ", err.Error())
				break
			}
			fmt.Println(string(in))
			err = conn.Send([]byte("hello, client"))
		}
		return nil
	}
	err := server.Serve(host, recvHandle)
	if err != nil {
		panic(err)
	}

	// Create a channel to listen for OS signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Block until a signal is received
	sig := <-sigChan
	fmt.Printf("Received signal: %s. Shutting down...\n", sig)

	// Perform any necessary cleanup here
	// For example, you might want to gracefully close the server
	// server.Close() // Uncomment if your server has a Close method

	// Exit the program
	os.Exit(0)

}

```

### Client demo ###

```go

import (
	"context"
	"fmt"
	"github.com/orbit-w/mux-go"
	"github.com/orbit-w/meteor/modules/net/transport"
	"os"
	"os/signal"
	"syscall"
)

/*
	@Author: orbit-w
	@File: main
	@2024 8月 周四 17:25
*/

func main() {
	host := "127.0.0.1:6800"
	conn := transport.DialContextWithOps(context.Background(), host)
	mux := mux.NewMultiplexer(context.Background(), conn)

	vc, err := mux.NewVirtualConn(context.Background())
	if err != nil {
		panic(err)
	}

	err = vc.Send([]byte("hello, server"))

	err = vc.CloseSend()

	// Create a channel to listen for OS signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Block until a signal is received
	sig := <-sigChan
	fmt.Printf("Received signal: %s. Shutting down...\n", sig)

	// Perform any necessary cleanup here
	// For example, you might want to gracefully close the server
	// server.Close() // Uncomment if your server has a Close method

	// Exit the program
	os.Exit(0)
}
```