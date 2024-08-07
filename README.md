# Multiplexer
mux-go is a multiplexing library for Golang.
It relies on an underlying connection to provide reliability and ordering.
The initial purpose of the design was to solve the problem of how to reuse a physical link with a virtual link when multiple clients are connected to the server in the game server,
thereby reducing the resource consumption of the server.

mux-go 是一个基于golang 语言的简单的多路复用库，它依赖于底层连接提供可靠性和排序。
是N条虚拟链接复用一条物理链路的设计，初衷是解决在游戏服务器中多个客户端连接到服务器时如何复用物理链路的问题，从而减少服务器的资源消耗。

## How it works

### Server ###

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

### Client ###

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