package network

import (
	"encoding/binary"
	"fmt"

	"github.com/fengqk/mars-base/common"
)

const (
	PACKET_LEN_BYTE  = 1
	PACKET_LEN_WORD  = 2
	PACKET_LEN_DWORD = 4
)

type (
	HandlePacket func(buff []byte)

	PacketParser struct {
		len           int32
		maxLen        int32
		littleEndian  bool
		maxPacketBuff []byte
		packetFunc    HandlePacket
	}

	PacketConfig struct {
		MaxPacketLen *int
		Func         HandlePacket
	}
)

func NewPacketParser(conf PacketConfig) PacketParser {
	p := PacketParser{}
	p.len = PACKET_LEN_DWORD
	p.maxLen = common.MAX_PACKET
	p.littleEndian = true
	if conf.Func != nil {
		p.packetFunc = conf.Func
	} else {
		p.packetFunc = func(buff []byte) {}
	}
	return p
}

func (p *PacketParser) readLen(buff []byte) (bool, int) {
	len := int32(len(buff))
	if len < p.len {
		return false, 0
	}

	bufMsgLen := buff[:p.len]
	var msgLen int32
	switch p.len {
	case PACKET_LEN_BYTE:
		msgLen = int32(bufMsgLen[0])
	case PACKET_LEN_WORD:
		if p.littleEndian {
			msgLen = int32(binary.LittleEndian.Uint16(bufMsgLen))
		} else {
			msgLen = int32(binary.BigEndian.Uint16(bufMsgLen))
		}
	case PACKET_LEN_DWORD:
		if p.littleEndian {
			msgLen = int32(binary.LittleEndian.Uint32(bufMsgLen))
		} else {
			msgLen = int32(binary.BigEndian.Uint32(bufMsgLen))
		}
	}
	if msgLen+p.len <= len {
		return true, int(msgLen + p.len)
	}
	return false, 0
}

func (p *PacketParser) Read(data []byte) bool {
	buff := append(p.maxPacketBuff, data...)
	p.maxPacketBuff = []byte{}
	cursize := 0
ParsePacket:
	packetsize := 0
	buffersize := len(buff[cursize:])
	findflag := false
	findflag, packetsize = p.readLen(buff[cursize:])
	if findflag {
		if buffersize == packetsize { //完整包
			p.packetFunc(buff[cursize+int(p.len) : cursize+packetsize])
			cursize += packetsize
		} else if buffersize > packetsize {
			p.packetFunc(buff[cursize+int(p.len) : cursize+packetsize])
			cursize += packetsize
			goto ParsePacket
		}
	} else if buffersize < int(p.maxLen) {
		p.maxPacketBuff = buff[cursize:]
	} else {
		fmt.Println("超出最大包限制，丢弃该包")
		return false
	}
	return true
}

func (p *PacketParser) Write(data []byte) []byte {
	msgLen := len(data)
	if msgLen+int(p.len) > common.MAX_PACKET {
		fmt.Println("write over common.MAX_PACKET")
	}

	msg := make([]byte, int(p.len)+msgLen)

	switch p.len {
	case PACKET_LEN_BYTE:
		msg[0] = byte(msgLen)
	case PACKET_LEN_WORD:
		if p.littleEndian {
			binary.LittleEndian.PutUint16(msg, uint16(msgLen))
		} else {
			binary.BigEndian.PutUint16(msg, uint16(msgLen))
		}
	case PACKET_LEN_DWORD:
		if p.littleEndian {
			binary.LittleEndian.PutUint32(msg, uint32(msgLen))
		} else {
			binary.BigEndian.PutUint32(msg, uint32(msgLen))
		}
	}

	copy(msg[p.len:], data)
	return msg
}
