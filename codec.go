package mux

import (
	"encoding/binary"
	"errors"
	"github.com/orbit-w/meteor/modules/net/packet"
)

/*
   @Author: orbit-w
   @File: streamer
   @2024 7月 周日 16:22
*/

const (
	typeFlagLength     = 1
	endFlagLength      = 1
	streamIdFlagLength = 8
)

type Codec struct{}

type Msg struct {
	Type int8
	End  bool
	Id   int64
	Data []byte
}

func (f *Codec) Encode(msg *Msg) packet.IPacket {
	w := packet.WriterP(1 + 1 + 8 + len(msg.Data))
	w.WriteInt8(msg.Type)
	w.WriteBool(msg.End)
	w.WriteInt64(msg.Id)
	if data := msg.Data; data != nil || len(data) > 0 {
		msg.Data = nil
		w.Write(data)
	}
	return w
}

func (f *Codec) Decode(data []byte) (Msg, error) {
	reader := packet.ReaderP(data)
	defer packet.Return(reader)
	msg := Msg{}
	ft, err := reader.ReadInt8()
	if err != nil {
		return msg, err
	}

	end, err := reader.ReadBool()
	if err != nil {
		return msg, err
	}

	sId, err := reader.ReadUint64()
	if err != nil {
		return msg, err
	}

	msg.Id = int64(sId)
	msg.Type = ft
	msg.End = end
	if len(reader.Remain()) > 0 {
		msg.Data = reader.CopyRemain()
	}
	return msg, nil
}

func (f *Codec) DecodeV2(data []byte) (Msg, error) {
	msg := Msg{}
	var off int
	if len(data) < typeFlagLength+endFlagLength+streamIdFlagLength {
		return msg, errors.New("decode failed")
	}

	ft := int8(data[0])
	end := data[1] == byte(1)
	off += 2
	sid := binary.BigEndian.Uint64(data[off : off+streamIdFlagLength])
	off += streamIdFlagLength
	msg.Id = int64(sid)
	msg.Type = ft
	msg.End = end

	data = data[off:]
	if len(data) > 0 {
		msg.Data = data
	}
	return msg, nil
}
