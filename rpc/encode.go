package rpc

import (
	"bytes"
	"encoding/gob"

	"github.com/fengqk/mars-base/common"

	"github.com/golang/protobuf/proto"
)

// rpc  Marshal
func Marshal(head *RpcHead, funcName *string, params ...interface{}) Packet {
	return marshal(head, funcName, params...)
}

// rpc  marshal
func marshal(head *RpcHead, funcName *string, params ...interface{}) Packet {
	defer func() {
		if err := recover(); err != nil {
			common.TraceCode(err)
		}
	}()

	*funcName = Route(head, *funcName)
	rpcPacket := &RpcPacket{FuncName: *funcName, ArgLen: int32(len(params)), RpcHead: (*RpcHead)(head)}
	buf := bytes.NewBuffer([]byte{})
	enc := gob.NewEncoder(buf)
	for _, param := range params {
		enc.Encode(param)
	}
	rpcPacket.RpcBody = buf.Bytes()
	data, _ := proto.Marshal(rpcPacket)
	return Packet{Buff: data, RpcPacket: rpcPacket}
}

// rpc  MarshalPB
func marshalPB(bitstream *common.BitStream, packet proto.Message) {
	bitstream.WriteString(proto.MessageName(packet))
	buf, _ := proto.Marshal(packet)
	nLen := len(buf)
	bitstream.WriteInt(nLen, 32)
	bitstream.WriteBits(buf, nLen<<3)
}
