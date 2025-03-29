package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/orbit-w/meteor/modules/net/transport"
	"github.com/orbit-w/mux-go"
	"github.com/spf13/viper"
)

/*
   @Author: orbit-w
   @File: main
   @2024 8月 周四 17:25
*/

func main() {
	parseConfig()

	host := "127.0.0.1:6800"
	conn := transport.DialWithOps(context.Background(), host)
	multiplexer := mux.NewMultiplexer(context.Background(), conn)

	vc, err := multiplexer.NewVirtualConn(context.Background())
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

func parseConfig() {
	// 设置配置文件名称（不带扩展名）
	viper.SetConfigName("config")

	// 设置配置文件类型
	viper.SetConfigType("toml")

	// 添加配置文件路径
	viper.AddConfigPath(".") // 当前目录

	// 尝试读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Sprintf("Error reading config file: %v", err))
	}
}
