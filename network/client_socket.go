package network

import (
	"fmt"
	"io"
	"log"
	"mars-base/common"
	"mars-base/rpc"
	"net"

	"github.com/xtaci/kcp-go"
)

type (
	IClientSocket interface {
		ISocket
	}

	ClientSocket struct {
		Socket
		maxClients int32
		minClients int32
	}
)

func (c *ClientSocket) Init(ip string, port int32, params ...OpOption) bool {
	if c.port == port || c.ip == ip {
		return false
	}

	return c.Socket.Init(ip, port, params...)
}

func (c *ClientSocket) Start() bool {
	if c.ip == "" {
		c.ip = "127.0.0.1"
	}

	if c.Connect() {
		go c.Run()
	}

	return true
}

func (c *ClientSocket) SendMsg(head rpc.RpcHead, funcName string, params ...interface{}) {
	c.Send(head, rpc.Marshal(&head, &funcName, params...))
}

func (c *ClientSocket) Send(head rpc.RpcHead, packet rpc.Packet) int {
	defer func() {
		if err := recover(); err != nil {
			common.TraceCode(err)
		}
	}()

	if c.conn == nil {
		return 0
	}

	n, err := c.conn.Write(c.packetParser.Write(packet.Buff))
	log.Printf("错误：%s\n", err.Error())
	if n > 0 {
		return n
	}
	return 0
}

func (c *ClientSocket) Restart() bool {
	return true
}

func (c *ClientSocket) Connect() bool {
	strRemote := fmt.Sprintf("%s:%d", c.ip, c.port)
	connectStr := "Tcp"
	if c.isKcp {
		conn, err := kcp.Dial(strRemote)
		if err != nil {
			return false
		}
		c.SetConn(conn)
		connectStr = "Kcp"
	} else {
		tcpAddr, err := net.ResolveTCPAddr("tcp4", strRemote)
		if err != nil {
			log.Printf("%v", err)
		}
		conn, err := net.DialTCP("tcp4", nil, tcpAddr)
		if err != nil {
			return false
		}
		c.SetConn(conn)
	}

	fmt.Printf("%s 连接成功\n", connectStr)
	c.CallMsg(rpc.RpcHead{}, "COMMON_RegisterRequest")
	return true
}

func (c *ClientSocket) OnDisconnect() {
}

func (c *ClientSocket) OnNetFail(int) {
	c.Stop()
	c.CallMsg(rpc.RpcHead{}, "DISCONNECT", c.clientId)
}

func (c *ClientSocket) Run() bool {
	c.SetState(SSF_RUN)
	var buff = make([]byte, c.recvBuffSize)
	loop := func() bool {
		defer func() {
			if err := recover(); err != nil {
				common.TraceCode(err)
			}
		}()

		if c.conn == nil {
			return false
		}

		n, err := c.conn.Read(buff)
		if err == io.EOF {
			fmt.Printf("远程链接：%s已经关闭！\n", c.conn.RemoteAddr().String())
			c.OnNetFail(0)
			return false
		}
		if err != nil {
			log.Printf("错误：%s\n", err.Error())
			c.OnNetFail(0)
			return false
		}
		if n > 0 {
			c.packetParser.Read(buff[:n])
		}
		return true
	}

	for {
		if !loop() {
			break
		}
	}

	c.Close()
	return true
}
