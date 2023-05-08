package cluster

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/fengqk/mars-base/actor"
	"github.com/fengqk/mars-base/common"
	"github.com/fengqk/mars-base/rpc"
)

const (
	fsm_idle    fsm_type = iota //空闲
	fsm_publish fsm_type = iota //注册
	fsm_lease   fsm_type = iota //ttl
)

type (
	fsm_type uint32

	Stub struct {
		fsm         fsm_type
		StubMailBox common.StubMailBox
		isRegister  int32
	}
)

func (s *Stub) InitStub(stub rpc.STUB) {
	s.StubMailBox.StubType = stub
	s.StubMailBox.ClusterId = MGR.Id()
	go s.updateFsm()
}

func (s *Stub) IsRegister() bool {
	return atomic.LoadInt32(&s.isRegister) == 1
}

func (s *Stub) lease() {
	err := MGR.StubMailBox.Lease(&s.StubMailBox)
	if err != nil {
		s.fsm = fsm_idle
		atomic.StoreInt32(&s.isRegister, 0)
		actor.MGR.SendMsg(rpc.RpcHead{SendType: rpc.SEND_BOARD_CAST}, fmt.Sprintf("%s.OnStubUnRegister", s.StubMailBox.StubType.String()))
		common.LOG.Printf("stub [%s]注销成功[%d]", s.StubMailBox.StubType.String(), s.StubMailBox.Id)
	} else {
		time.Sleep(STUB_TTL_TIME / 3)
	}
}
