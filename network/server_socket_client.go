package network

import (
	"fmt"
	"hash/crc32"
	"io"
	"log"
	"sync/atomic"
	"time"

	"github.com/fengqk/mars-base/common"
	"github.com/fengqk/mars-base/common/timer"
	"github.com/fengqk/mars-base/rpc"
)

const (
	IDLE_TIMEOUT    = iota
	CONNECT_TIMEOUT = iota
	CONNECT_TYPE    = iota
)

var (
	DISCONNECTINT = crc32.ChecksumIEEE([]byte("DISCONNECT"))
	HEART_PACKET  = crc32.ChecksumIEEE([]byte("heardpacket"))
)

type (
	IServerSocketClient interface {
		ISocket
	}

	ServerSocketClient struct {
		Socket
		server   *ServerSocket
		sendChan chan []byte
		timerId  *int64
	}
)

func (s *ServerSocketClient) Init(ip string, port int32, params ...OpOption) bool {
	s.Socket.Init(ip, port, params...)
	s.timerId = new(int64)
	return true
}

func (s *ServerSocketClient) Start() bool {
	if s.server == nil {
		return false
	}

	if s.connType == CLIENT_CONNECT {
		s.sendChan = make(chan []byte, MAX_SEND_CHAN)
		timer.StoreTimerId(s.timerId, int64(s.clientId)+1<<32)
		timer.RegisterTimer(s.timerId, (HEART_TIME_OUT/3)*time.Second, func() {
			s.Update()
		})
	}

	if s.packetFuncList.Count() == 0 {
		s.packetFuncList = s.server.packetFuncList
	}

	go s.Run()

	if s.connType == CLIENT_CONNECT {
		go s.SendLoop()
	}

	return true
}

func (s *ServerSocketClient) Send(head rpc.RpcHead, packet rpc.Packet) int {
	defer func() {
		if err := recover(); err != nil {
			common.TraceCode(err)
		}
	}()

	if s.connType == CLIENT_CONNECT { //对外链接send不阻塞
		select {
		case s.sendChan <- packet.Buff:
		default: //网络太卡,tcp send缓存满了并且发送队列也满了
			s.OnNetFail(1)
		}
	} else {
		return s.DoSend(packet.Buff)
	}
	return 0
}

func (s *ServerSocketClient) DoSend(buff []byte) int {
	if s.conn == nil {
		return 0
	}

	n, err := s.conn.Write(s.packetParser.Write(buff))
	log.Printf("错误：%s\n", err.Error())
	if n > 0 {
		return n
	}

	return 0
}

func (s *ServerSocketClient) OnNetFail(error int) {
	s.Stop()
	if s.connType == CLIENT_CONNECT {
		stream := common.NewBitStream(make([]byte, 32), 32)
		stream.WriteInt(int(DISCONNECTINT), 32)
		stream.WriteInt(int(s.clientId), 32)
		s.HandlePacket(stream.GetBuffer())
	} else {
		s.CallMsg(rpc.RpcHead{}, "DISCONNECT", s.clientId)
	}
	if s.server != nil {
		s.server.DelClient(s)
	}
}

func (s *ServerSocketClient) Stop() bool {
	timer.RegisterTimer(s.timerId, timer.TICK_INTERVAL, func() {
		timer.StopTimer(s.timerId)
		if atomic.CompareAndSwapInt32(&s.state, SSF_RUN, SSF_STOP) {
			if s.conn != nil {
				s.conn.Close()
			}
		}
	})
	return false
}

func (s *ServerSocketClient) Close() {
	if s.connType == CLIENT_CONNECT {
		s.sendChan <- nil
	}
	s.Socket.Close()
	if s.server != nil {
		s.server.DelClient(s)
	}
}

func (s *ServerSocketClient) Run() bool {
	var buff = make([]byte, s.recvBuffSize)
	s.SetState(SSF_RUN)
	loop := func() bool {
		defer func() {
			if err := recover(); err != nil {
				common.TraceCode(err)
			}
		}()

		if s.conn == nil {
			return false
		}

		n, err := s.conn.Read(buff)
		if err == io.EOF {
			fmt.Printf("远程链接：%s已经关闭！\n", s.conn.RemoteAddr().String())
			s.OnNetFail(0)
			return false
		}
		if err != nil {
			log.Printf("错误：%s\n", err.Error())
			s.OnNetFail(0)
			return false
		}
		if n > 0 {
			if !s.packetParser.Read(buff[:n]) && s.connType == CLIENT_CONNECT {
				s.OnNetFail(1)
				return false
			}
		}
		s.heartTime = int64(time.Now().Unix()) + HEART_TIME_OUT
		return true
	}

	for {
		if !loop() {
			break
		}
	}

	s.Close()
	fmt.Printf("%s关闭连接", s.ip)
	return true
}

func (s *ServerSocketClient) Update() {
	now := int64(time.Now().Unix())
	// timeout
	if s.heartTime < now {
		s.OnNetFail(2)
		return
	}
}

func (s *ServerSocketClient) SendLoop() bool {
	for {
		defer func() {
			if err := recover(); err != nil {
				common.TraceCode(err)
			}
		}()

		select {
		case buff := <-s.sendChan:
			if buff == nil { //信道关闭
				return false
			} else {
				s.DoSend(buff)
			}
		}
	}
	return true
}
