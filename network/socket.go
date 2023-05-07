package network

import "net"

const (
	SSF_NULL = iota
	SSF_RUN  = iota
	SSF_STOP = iota
)

const (
	CLIENT_CONNECT = iota
	SERVER_CONNECT = iota
)

const (
	MAX_SEND_CHAN  = 100
	HEART_TIME_OUT = 30
)

type (
	PacketFunc   func(packet rpc.Packet) bool
	HandlePacket func(buff []byte)

	Op struct {
		kcp bool
	}

	OpOption func(*Op)

	Socket struct {
		conn         net.Conn
		port         int32
		ip           string
		state        int32
		connType     int32
		recvBuffSize int32

		clientId uint32
		seqId    int64

		totalNum     int32
		acceptedNum  int32
		connectedNum int32

		sendTimes int32
		recvTimes int32
	}
)
