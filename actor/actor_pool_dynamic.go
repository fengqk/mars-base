package actor

import (
	"reflect"
	"sync"

	"github.com/fengqk/mars-base/base"
	"github.com/fengqk/mars-base/rpc"
)

// ********************************************************
// actorpooldynamic管理, 可以动态添加
// ********************************************************
type (
	IActorPoolDynamic interface {
		GetActor(Id int64) IActor
		AddActor(ac IActor)
		DelActor(Id int64)
		GetActorNum() int
		GetMgr() IActor
	}

	ActorPoolDynamic struct {
		MGR       IActor
		actorMap  map[int64]IActor
		actorLock *sync.RWMutex
	}
)

func (a *ActorPoolDynamic) InitActor(pPool IActorPool, rType reflect.Type) {
	a.actorMap = make(map[int64]IActor)
	a.actorLock = &sync.RWMutex{}
	a.MGR = reflect.New(rType).Interface().(IActor)
	MGR.RegisterActor(a.MGR, WithType(ACTOR_TYPE_VIRTUAL), withPool(pPool))
}

func (a *ActorPoolDynamic) AddActor(ac IActor) {
	rType := reflect.TypeOf(ac)
	op := Op{actorType: ACTOR_TYPE_VIRTUAL, name: base.GetClassName(rType)}
	ac.register(ac, op)
	a.actorLock.Lock()
	a.actorMap[ac.GetId()] = ac
	a.actorLock.Unlock()
}

func (a *ActorPoolDynamic) DelActor(Id int64) {
	a.actorLock.Lock()
	delete(a.actorMap, Id)
	a.actorLock.Unlock()
}

func (a *ActorPoolDynamic) GetActor(Id int64) IActor {
	a.actorLock.RLock()
	ac, bEx := a.actorMap[Id]
	a.actorLock.RUnlock()
	if bEx {
		return ac
	}
	return nil
}

func (a *ActorPoolDynamic) GetActorNum() int {
	nLen := 0
	a.actorLock.RLock()
	nLen = len(a.actorMap)
	a.actorLock.RUnlock()
	return nLen
}

func (a *ActorPoolDynamic) GetMgr() IActor {
	return a.MGR
}

func (a *ActorPoolDynamic) SendAcotr(head rpc.RpcHead, packet rpc.Packet) bool {
	if a.MGR.HasRpc(packet.RpcPacket.FuncName) {
		if head.Id != 0 {
			ac := a.GetActor(head.Id)
			if ac != nil {
				ac.getActor().Send(head, packet)
			}
		}
		return true
	}
	return false
}
