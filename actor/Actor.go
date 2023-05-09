package actor

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"runtime"
	"strings"
	"sync/atomic"
	"time"

	"github.com/fengqk/mars-base/base"
	"github.com/fengqk/mars-base/base/mpsc"
	"github.com/fengqk/mars-base/common/timer"
	"github.com/fengqk/mars-base/rpc"
)

var (
	g_IdSeed int64
)

const (
	ASF_NULL = iota
	ASF_RUN  = iota
	ASF_STOP = iota
)

const (
	DESTROY_EVENT = iota
)

const (
	ACTOR_TYPE_SINGLETON ACTOR_TYPE = iota //单例
	ACTOR_TYPE_VIRTUAL   ACTOR_TYPE = iota //玩家 必须初始一个全局的actor 作为类型判断
	ACTOR_TYPE_POOL      ACTOR_TYPE = iota //固定数量actor池
	ACTOR_TYPE_STUB      ACTOR_TYPE = iota //stub
)

type (
	ACTOR_TYPE uint32

	Op struct {
		name      string
		actorType ACTOR_TYPE
		pool      IActorPool
	}

	OpOption func(*Op)

	IActor interface {
		Init()
		Start()
		Stop()
		SendMsg(head rpc.RpcHead, funcName string, params ...interface{})
		Send(head rpc.RpcHead, packet rpc.Packet)
		RegisterTimer(duration time.Duration, fun func(), opts ...timer.OpOption)
		GetId() int64
		GetState() int32
		GetRpcHead(ctx context.Context) rpc.RpcHead
		GetName() string
		GetActorType() ACTOR_TYPE
		HasRpc(string) bool
		Actor() *Actor
		register(IActor, Op)
		setState(state int32)
		bindPool(IActorPool)
		getPool() IActorPool
	}

	IActorPool interface {
		SendActor(head rpc.RpcHead, packet rpc.Packet) bool
	}

	ActorBase struct {
		actorName string
		actorType ACTOR_TYPE
		rType     reflect.Type
		rValue    reflect.Value
		Self      IActor
	}

	Actor struct {
		ActorBase
		actorChan chan int
		id        int64
		state     int32
		trace     TraceInfo
		mailBox   *mpsc.Queue[*CallIO]
		mailIn    [8]int64
		mailChan  chan bool
		timerId   *int64
		pool      IActorPool
		timerMap  map[uintptr]func()
	}

	CallIO struct {
		rpc.RpcHead
		*rpc.Packet
		Buff []byte
	}

	TraceInfo struct {
		funcName  string
		fileName  string
		filePath  string
		className string
	}
)

func (a *ActorBase) IsActorType(actorType ACTOR_TYPE) bool {
	return a.actorType == actorType
}

func AssignActorId() int64 {
	return atomic.AddInt64(&g_IdSeed, 1)
}

func (a *Actor) Init() {
	a.mailChan = make(chan bool, 1)
	a.mailBox = mpsc.New[*CallIO]()
	a.actorChan = make(chan int, 1)
	a.timerMap = make(map[uintptr]func())
	a.trace.Init()
	if a.id == 0 {
		a.id = AssignActorId()
	}
	a.timerId = new(int64)
}

func (a *Actor) Start() {
	if atomic.CompareAndSwapInt32(&a.state, ASF_NULL, ASF_RUN) {
		go a.run()
	}
}

func (a *Actor) Stop() {
	timer.RegisterTimer(a.timerId, timer.TICK_INTERVAL, func() {
		timer.StopTimer(a.timerId)
		if atomic.CompareAndSwapInt32(&a.state, ASF_RUN, ASF_STOP) {
			a.actorChan <- DESTROY_EVENT
		}
	})
}

func (a *Actor) SendMsg(head rpc.RpcHead, funcName string, params ...interface{}) {
	head.SocketId = 0
	a.Send(head, rpc.Marshal(&head, &funcName, params...))
}

func (a *Actor) Send(head rpc.RpcHead, packet rpc.Packet) {
	defer func() {
		if err := recover(); err != nil {
			base.TraceCode(err)
		}
	}()

	var io CallIO
	io.RpcHead = head
	io.RpcPacket = packet.RpcPacket
	io.Buff = packet.Buff
	a.mailBox.Push(&io)
	if atomic.LoadInt64(&a.mailIn[0]) == 0 && atomic.CompareAndSwapInt64(&a.mailIn[0], 0, 1) {
		a.mailChan <- true
	}
}

func (a *Actor) Trace(funcName string) {
	a.trace.funcName = funcName
}

func (a *Actor) RegisterTimer(duration time.Duration, fun func(), opts ...timer.OpOption) {
	timer.StoreTimerId(a.timerId, a.id)
	//&fun这里有问题,会产生一对闭包函数,再释放的释放有问题
	ptr := uintptr(reflect.ValueOf(fun).Pointer())
	a.timerMap[ptr] = fun
	timer.RegisterTimer(a.timerId, duration, func() {
		a.SendMsg(rpc.RpcHead{ActorName: a.actorName}, "UpdateTimer", ptr)
	}, opts...)
}

func (a *Actor) GetId() int64 {
	return a.id
}

func (a *Actor) SetId(id int64) {
	a.id = id
}

func (a *Actor) GetName() string {
	return a.actorName
}

func (a *Actor) GetRpcHead(ctx context.Context) rpc.RpcHead {
	rpcHead := ctx.Value("rpcHead").(rpc.RpcHead)
	return rpcHead
}

func (a *Actor) GetState() int32 {
	return atomic.LoadInt32(&a.state)
}

func (a *Actor) GetActorType() ACTOR_TYPE {
	return a.actorType
}

func (a *Actor) HasRpc(funcName string) bool {
	_, bEx := a.rType.MethodByName(funcName)
	return bEx
}

func (a *Actor) UpdateTimer(ctx context.Context, ptr uintptr) {
	fun, isEx := a.timerMap[ptr]
	if isEx {
		a.Trace("timer")
		(fun)()
		a.Trace("")
	}
}

func (a *Actor) Actor() *Actor {
	return a
}

func (a *Actor) setState(state int32) {
	atomic.StoreInt32(&a.state, state)
}

func (a *Actor) bindPool(pool IActorPool) {
	a.pool = pool
}

func (a *Actor) getPool() IActorPool {
	return a.pool
}

func (a *Actor) register(ac IActor, op Op) {
	rType := reflect.TypeOf(ac)
	a.ActorBase = ActorBase{rType: rType, rValue: reflect.ValueOf(ac), Self: ac, actorName: op.name, actorType: op.actorType}
}

func (a *Actor) clear() {
	a.id = 0
	a.setState(ASF_NULL)
	timer.StopTimer(a.timerId)
}

func (a *Actor) run() {
	for {
		if !a.loop() {
			break
		}
	}
	a.clear()
}

func (a *Actor) loop() bool {
	defer func() {
		if err := recover(); err != nil {
			base.TraceCode(a.trace.ToString(), err)
		}
	}()

	select {
	case <-a.mailChan:
		a.consume()
	case msg := <-a.actorChan:
		if msg == DESTROY_EVENT {
			return false
		}
	}
	return true
}

func (a *Actor) consume() {
	atomic.StoreInt64(&a.mailIn[0], 0)
	for data := a.mailBox.Pop(); data != nil; data = a.mailBox.Pop() {
		a.call(data)
	}
}

func (a *Actor) call(io *CallIO) {
	rpcPakcet := io.RpcPacket
	rpcHead := io.RpcHead
	funcName := rpcPakcet.FuncName

	m, bEx := a.rType.MethodByName(funcName)
	if !bEx {
		log.Printf("func %s has no method", funcName)
		return
	}

	rpcPakcet.RpcHead.SocketId = io.SocketId
	params := rpc.UnmarshalBody(rpcPakcet, m.Type)
	if len(params) >= 1 {
		in := make([]reflect.Value, len(params))
		in[0] = a.rValue
		for i, param := range params {
			if i == 0 {
				continue
			}
			in[i] = reflect.ValueOf(param)
		}
		a.Trace(funcName)
		ret := m.Func.Call(in)
		a.Trace("")
		if ret != nil && rpcHead.Reply != "" {
			ret = append([]reflect.Value{reflect.ValueOf(&rpcHead)}, ret...)
			rpc.MGR.Call(ret)
		}
	} else {
		log.Printf("func %s params too short", funcName)
	}
}

func (a *TraceInfo) Init() {
	_, file, _, bOk := runtime.Caller(2)
	if bOk {
		index := strings.LastIndex(file, "/")
		if index != -1 {
			a.fileName = file[index+1:]
			a.filePath = file[:index]
			index2 := strings.LastIndex(a.fileName, ".")
			if index2 != -1 {
				a.className = strings.ToLower(a.fileName[:index2])
			}
		}
	}
}

func (a *TraceInfo) ToString() string {
	return fmt.Sprintf("trace go file %s call %s\n", a.fileName, a.funcName)
}

func (op *Op) applyOpts(opts []OpOption) {
	for _, opt := range opts {
		opt(op)
	}
}

func (op *Op) IsActorType(actorType ACTOR_TYPE) bool {
	return op.actorType == actorType
}
