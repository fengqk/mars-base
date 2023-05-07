package actor

import (
	"log"
	"mars-base/common"
	"mars-base/network"
	"mars-base/rpc"
	"reflect"
)

var (
	MGR *ActorMgr
)

type (
	IActorMgr interface {
		Init()
		RegisterActor(ac IActor, params ...OpOption)
		PacketFunc(rpc.Packet) bool
		SendMsg(rpc.RpcHead, string, ...interface{})
	}

	ICluster interface {
		BindPacketFunc(packetFunc network.PacketFunc)
	}

	ActorMgr struct {
		actorTypeMap map[reflect.Type]IActor
		actorMap     map[string]IActor
		isStart      bool
	}
)

func init() {
	MGR = &ActorMgr{}
	MGR.Init()
}

func WithType(actor_type ACTOR_TYPE) OpOption {
	return func(op *Op) {
		op.actorType = actor_type
	}
}

func withPool(pPool IActorPool) OpOption {
	return func(op *Op) {
		op.pool = pPool
	}
}

func (a *ActorMgr) Init() {
	a.actorTypeMap = make(map[reflect.Type]IActor)
	a.actorMap = make(map[string]IActor)
}

func (a *ActorMgr) Start() {
	a.isStart = true
}

func (a *ActorMgr) RegisterActor(ac IActor, params ...OpOption) {
	op := Op{}
	op.applyOpts(params)
	rType := reflect.TypeOf(ac)
	name := common.GetClassName(rType)
	_, bEx := a.actorTypeMap[rType]
	if bEx {
		log.Panicf("RegisterActor, %s must unique name", name)
		return
	}
	op.name = name
	ac.register(ac, op)
	a.actorTypeMap[rType] = ac
	a.actorMap[name] = ac
	if op.pool != nil {
		ac.bindPool(op.pool)
	}
}

func (a *ActorMgr) SendMsg(head rpc.RpcHead, funcName string, params ...interface{}) {
	head.SocketId = 0
	a.SendActor(funcName, head, rpc.Marshal(&head, &funcName, params...))
}

func (a *ActorMgr) SendActor(funcName string, head rpc.RpcHead, packet rpc.Packet) bool {
	var ac, bEx = a.actorMap[head.ActorName]
	if bEx && ac != nil {
		if ac.HasRpc(funcName) {
			switch ac.GetActorType() {
			case ACTOR_TYPE_SINGLETON:
				ac.Actor().Send(head, packet)
				return true
			case ACTOR_TYPE_VIRTUAL:
				return ac.getPool().SendActor(head, packet)
			case ACTOR_TYPE_POOL:
				return ac.getPool().SendActor(head, packet)
			}
		}
	}
	return false
}

func (a *ActorMgr) PacketFunc(packet rpc.Packet) bool {
	rpcPacket, head := rpc.Unmarshal(packet.Buff)
	packet.RpcPacket = rpcPacket
	head.SocketId = packet.Id
	head.Reply = packet.Reply
	return a.SendActor(rpcPacket.FuncName, head, packet)
}
