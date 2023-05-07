package network

import (
	"net"
	"sync/atomic"

	"github.com/fengqk/mars-base/common/vector"
	"github.com/fengqk/mars-base/rpc"
)

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
	PacketFunc func(packet rpc.Packet) bool

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

		sendTimes      int32
		recvTimes      int32
		packetFuncList *vector.Vector[PacketFunc]

		isHalf       bool
		halfSize     int32
		packetParser PacketParser
		heartTime    int64
		isKcp        bool
	}

	ISocket interface {
		Init(string, int, ...OpOption) bool
		Start() bool
		Stop() bool
		Run() bool
		Restart() bool
		Connect() bool
		Disconnect(bool) bool
		OnNetFail(int)
		Clear()
		Close()
		SendMsg(rpc.RpcHead, string, ...interface{})
		Send(rpc.RpcHead, rpc.Packet) int
		CallMsg(rpc.RpcHead, string, ...interface{}) //回调消息处理

		GetId() uint32
		GetState() int32
		SetState(int32)
		SetReceiveBufferSize(int32)
		GetReceiveBufferSize() int32
		SetMaxPacketLen(int32)
		GetMaxPacketLen() int32
		BindPacketFunc(PacketFunc)
		SetConnectType(int32)
		SetConn(net.Conn)
		HandlePacket([]byte)
	}
)

func (op *Op) applyOpts(opts []OpOption) {
	for _, opt := range opts {
		opt(op)
	}
}

func WithKcp() OpOption {
	return func(op *Op) {
		op.kcp = true
	}
}

func (this *Socket) Init(ip string, port int32, params ...OpOption) bool {
	op := &Op{}
	op.applyOpts(params)
	this.ip = ip
	this.port = port
	this.packetFuncList = &vector.Vector[PacketFunc]{}
	this.SetState(SSF_NULL)
	this.recvBuffSize = 1024
	this.connType = SERVER_CONNECT
	this.isHalf = false
	this.halfSize = 0
	this.heartTime = 0
	this.packetParser = NewPacketParser(PacketConfig{Func: this.HandlePacket})
	if op.kcp {
		this.isKcp = true
	}
	return true
}

func (this *Socket) Start() bool {
	return true
}

func (this *Socket) Stop() bool {
	if atomic.CompareAndSwapInt32(&this.state, SSF_RUN, SSF_STOP) {
		if this.conn != nil {
			this.conn.Close()
		}
	}
	return false
}

func (this *Socket) Run() bool {
	return true
}

func (this *Socket) Restart() bool {
	return true
}

func (this *Socket) Connect() bool {
	return true
}

func (this *Socket) Disconnect(bool) bool {
	return true
}

func (this *Socket) OnNetFail(int) {
	this.Stop()
}

func (this *Socket) GetId() uint32 {
	return this.clientId
}

func (this *Socket) GetState() int32 {
	return atomic.LoadInt32(&this.state)
}

func (this *Socket) SetState(state int32) {
	atomic.StoreInt32(&this.state, state)
}

func (this *Socket) SendMsg(head rpc.RpcHead, funcName string, params ...interface{}) {
	return
}

func (this *Socket) Send(rpc.RpcHead, rpc.Packet) int {
	return 0
}

func (this *Socket) Clear() {
	this.SetState(SSF_NULL)
	this.conn = nil
	this.recvBuffSize = 1024
	this.isHalf = false
	this.halfSize = 0
	this.heartTime = 0
}

func (this *Socket) Close() {
	this.Clear()
}

func (this *Socket) GetMaxPacketLen() int32 {
	return this.packetParser.maxLen
}

func (this *Socket) SetMaxPacketLen(maxReceiveSize int32) {
	this.packetParser.maxLen = maxReceiveSize
}

func (this *Socket) GetReceiveBufferSize() int32 {
	return this.recvBuffSize
}

func (this *Socket) SetReceiveBufferSize(maxSendSize int32) {
	this.recvBuffSize = maxSendSize
}

func (this *Socket) SetConnectType(connType int32) {
	this.connType = connType
}

func (this *Socket) SetConn(conn net.Conn) {
	this.conn = conn
}

func (this *Socket) BindPacketFunc(callfunc PacketFunc) {
	this.packetFuncList.PushBack(callfunc)
}

func (this *Socket) CallMsg(head rpc.RpcHead, funcName string, params ...interface{}) {
	this.HandlePacket(rpc.Marshal(&head, &funcName, params...).Buff)
}

func (this *Socket) HandlePacket(buff []byte) {
	packet := rpc.Packet{Id: this.clientId, Buff: buff}
	for _, v := range this.packetFuncList.Values() {
		if v(packet) {
			break
		}
	}
}
