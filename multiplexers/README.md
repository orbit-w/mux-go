# mux-go/multiplexers

`mux-go/multiplexers` 是一个用于管理多路复用器的 Go 语言库。它允许在M物理连接上创建N个虚拟连接，从而提高网络通信的效率和并发性能。

## 特性

- 支持创建和管理多个虚拟连接
- 提供高效的并发处理
- 通过最小化锁的粒度来提高性能
- 支持客户端和服务端模式

## 安装

使用 `go get` 命令安装：

```sh
go get github.com/orbit-w/mux-go/multiplexers
```

## 使用方法

### 创建多路复用器

```go
import (
    "context"
    "github.com/orbit-w/mux-go/multiplexers"
    "github.com/orbit-w/meteor/modules/net/transport"
)

func main() {
    host := "localhost:8080"
    size := 5
    connsCount := 10
    m := multiplexers.NewMultiplexers(host, size, connsCount)
    // 使用 m 进行虚拟连接的创建和管理
}
```

### 创建虚拟连接

```go
conn, err := m.Dial()
if err != nil {
    log.Fatalf("Failed to create virtual connection: %v", err)
}
// 使用 conn 进行数据传输
```

### 关闭多路复用器

```go
m.Close()
```

## 接口说明

### IConn 接口

`IConn` 接口定义了虚拟连接的基本操作：

- `Send(data []byte) error`：发送数据
- `Recv(ctx context.Context) ([]byte, error)`：接收数据
- `Close() error`：关闭连接

### Multiplexers 类型

`Multiplexers` 类型用于管理虚拟连接：

- `NewMultiplexers(host string, size, connsCount int) *Multiplexers`：创建新的多路复用器管理器
- `Dial() (IConn, error)`：创建新的虚拟连接
- `DialEdge() (IConn, error)`：创建新的虚拟连接，优化并发性能
- `Close()`：关闭多路复用器

## 贡献

欢迎提交问题和贡献代码。请确保在提交之前运行所有测试。

## 许可证

该项目使用 MIT 许可证。详情请参阅 LICENSE 文件。
```