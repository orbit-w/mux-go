package mux

import (
	"context"
	"fmt"
	"github.com/orbit-w/meteor/bases/misc/utils"
	"github.com/orbit-w/meteor/bases/net/packet"
	"github.com/orbit-w/meteor/modules/net/transport"
	"github.com/orbit-w/mux-go/metadata"
	"io"
	"log"
	"runtime/debug"
	"sync/atomic"
)

/*
   @Author: orbit-w
   @File: client
   @2024 7月 周日 19:12
*/

type IMux interface {
	NewVirtualConn(ctx context.Context) (IConn, error)
	Close()
}

type Multiplexer struct {
	isClient     bool
	state        atomic.Uint32
	conn         transport.IConn
	codec        *Codec
	virtualConns *VirtualConns
	ctx          context.Context
	cancel       context.CancelFunc

	conf   MuxClientConfig //client side config
	server *Server         //server side
}

func NewMultiplexer(f context.Context, conn transport.IConn, ops ...MuxClientConfig) IMux {
	conf := parseConfig(ops...)
	mux := newCliMultiplexer(f, conn, conf)
	go mux.recvLoop()
	return mux
}

func newMultiplexer(f context.Context, conn transport.IConn, isClient bool, server *Server) *Multiplexer {
	ctx, cancel := context.WithCancel(f)
	mux := &Multiplexer{
		isClient:     isClient,
		conn:         conn,
		virtualConns: newConns(0),
		ctx:          ctx,
		cancel:       cancel,
		codec:        new(Codec),
		server:       server,
	}
	return mux
}

func newCliMultiplexer(f context.Context, conn transport.IConn, conf MuxClientConfig) *Multiplexer {
	ctx, cancel := context.WithCancel(f)
	mux := &Multiplexer{
		isClient:     true,
		conn:         conn,
		virtualConns: newConns(conf.MaxVirtualConns),
		ctx:          ctx,
		cancel:       cancel,
		codec:        new(Codec),
		conf:         conf,
	}
	return mux
}

func (mux *Multiplexer) NewVirtualConn(ctx context.Context) (IConn, error) {
	md, _ := metadata.FromOutContext(ctx)
	data, err := metadata.Marshal(md)
	if err != nil {
		return nil, err
	}

	id := mux.virtualConns.Id()
	vc := virtualConn(ctx, id, mux.conn, mux)

	if !mux.virtualConns.Reg(id, vc) {
		return nil, ErrVirtualConnUpLimit
	}

	fp := mux.codec.Encode(&Msg{
		Type: MessageStart,
		Id:   id,
		Data: data,
	})

	defer packet.Return(fp)

	if err = vc.conn.Send(fp.Data()); err != nil {
		mux.virtualConns.Del(id)
		return nil, NewStreamBufSetErr(err)
	}
	return vc, nil
}

func (mux *Multiplexer) Close() {
	if mux.state.CompareAndSwap(StateMuxNormal, StateMuxStopped) {
		if mux.conn != nil {
			_ = mux.conn.Close()
		}
	}
}

func (mux *Multiplexer) recvLoop() {
	var (
		in  []byte
		err error
		ctx = context.Background()
	)

	defer func() {
		mux.state.Store(StateMuxStopped)
		if mux.conn != nil {
			_ = mux.conn.Close()
		}

		closeErr := ErrCancel
		if err != nil {
			if !(err == io.EOF || IsErrCanceled(err)) {
				closeErr = err
				log.Println(fmt.Errorf("conn disconnected: %s", err.Error()))
			}
		}
		mux.virtualConns.Close(func(stream *VirtualConn) {
			stream.OnClose(closeErr)
		})
	}()

	var msg Msg

	for {
		in, err = mux.conn.Recv(ctx)
		if err != nil {
			return
		}

		msg, err = mux.codec.Decode(in)
		if err != nil {
			err = NewDecodeErr(err)
			return
		}

		handle := getHandler(getName(mux))
		handle(mux, &msg)
	}
}

// loopVirtualConn
// server side, recvLoop the virtual connection
// 服务端侧，有新的虚拟链接进来，需要循环处理
// 业务侧只需要break/return即可
func (mux *Multiplexer) acceptVirtualConn(ctx context.Context, conn transport.IConn, id int64) {
	vc := virtualConn(ctx, id, conn, mux)
	mux.virtualConns.Reg(id, vc)
	utils.GoRecoverPanic(func() {
		mux.handleVirtualConn(vc)
	})
}

func (mux *Multiplexer) handleVirtualConn(conn *VirtualConn) {
	defer func() {
		if r := recover(); r != nil {
			debug.PrintStack()
		}

		if _, exist := mux.virtualConns.GetAndDel(conn.Id()); exist {
			err := conn.rb.GetErr()
			if err == nil || err == io.EOF {
				conn.sendMsgFin()
			}
		}

		// Close the stream
		// Simultaneously disconnect the input and output of virtual connections
		conn.OnClose(io.EOF)
	}()

	handle := mux.server.handle
	if err := handle(conn); err != nil {
		//TODO:
	}
}

func handleDataClientSide(mux *Multiplexer, in *Msg) {
	switch in.Type {
	case MessageRaw:
		if len(in.Data) > 0 {
			v, ok := mux.virtualConns.Get(in.Id)
			if ok {
				v.put(in.Data)
			}
		}
	case MessageFin:
		stream, ok := mux.virtualConns.GetAndDel(in.Id)
		if ok {
			stream.OnClose(io.EOF)
		}
	}
}

func handleDataServerSide(mux *Multiplexer, in *Msg) {
	switch in.Type {
	case MessageStart:
		if mux.virtualConns.Exist(in.Id) {
			return
		}

		md := metadata.MD{}
		if err := metadata.Unmarshal(in.Data, &md); err != nil {
			//remote close the virtual connection
			pack := mux.codec.Encode(&Msg{
				Type: MessageFin,
				Id:   in.Id,
			})
			_ = mux.conn.Send(pack.Data())
			packet.Return(pack)
			log.Println("[TcpServer] [func:handleStartFrame] metadata unmarshal failed: ", err.Error())
			return
		}

		ctx := metadata.NewIncomingContext(mux.ctx, md)
		mux.acceptVirtualConn(ctx, mux.conn, in.Id)

	case MessageRaw:
		streamId := in.Id
		if in.End {
			vc, ok := mux.virtualConns.Get(streamId)
			if ok {
				vc.OnClose(io.EOF)
			}
			return
		}

		if len(in.Data) > 0 {
			v, ok := mux.virtualConns.Get(in.Id)
			if ok {
				v.put(in.Data)
			}
		}
	}
}

func getName(mux *Multiplexer) string {
	if mux.isClient {
		return handleNameClient
	}
	return handleNameServer
}
