package actor

import (
	"reflect"

	"github.com/fengqk/mars-base/base"
	"github.com/fengqk/mars-base/rpc"
)

// ********************************************************
// actorpool管理,不能动态分配
// ********************************************************
type (
	IActorPool interface {
		SendActor(head rpc.RpcHead, packet rpc.Packet) bool
	}

	ActorPool struct {
		MGR       IActor
		actorList []IActor
		actorSize int32
	}
)

func (a *ActorPool) InitPool(pool IActorPool, rType reflect.Type, num int32) {
	a.actorList = make([]IActor, num)
	a.actorSize = num
	for i := 0; i < int(num); i++ {
		ac := reflect.New(rType).Interface().(IActor)
		rType := reflect.TypeOf(ac)
		op := Op{actorType: ACTOR_TYPE_POOL, name: base.GetClassName(rType)}
		ac.register(ac, op)
		ac.Init()
		a.actorList[i] = ac
	}
	a.MGR = reflect.New(rType).Interface().(IActor)
	MGR.RegisterActor(a.MGR, WithType(ACTOR_TYPE_POOL), withPool(pool))
}

func (a *ActorPool) GetPoolSize() int32 {
	return a.actorSize
}

func (a *ActorPool) SendActor(head rpc.RpcHead, packet rpc.Packet) bool {
	if a.MGR.HasRpc(packet.RpcPacket.FuncName) {
		switch head.SendType {
		case rpc.SEND_POINT:
			index := head.Id % int64(a.actorSize)
			a.actorList[index].getActor().Send(head, packet)
		default:
			for i := 0; i < int(a.actorSize); i++ {
				a.actorList[i].getActor().Send(head, packet)
			}
		}
		return true
	}
	return false
}
