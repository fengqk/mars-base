package cluster

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"time"

	"github.com/fengqk/mars-base/actor"
	"github.com/fengqk/mars-base/base"
	"github.com/fengqk/mars-base/base/vector"
	"github.com/fengqk/mars-base/cluster/etcd"
	"github.com/fengqk/mars-base/common"
	"github.com/fengqk/mars-base/network"
	"github.com/fengqk/mars-base/rpc"
	"github.com/nats-io/nats.go"
)

const (
	MAX_CLUSTER_NUM = int(rpc.SERVICE_NUM)
	CALL_TIME_OUT   = 500 * time.Millisecond
)

var (
	MGR Cluster
)

type (
	ClusterMap       map[uint32]*base.ClusterInfo
	ClusterSocketMap map[uint32]*base.ClusterInfo

	Op struct {
		mailBoxEndpoints     []string
		stubMailBoxEndpoints []string
		stub                 common.Stub
	}

	OpOption func(*Op)

	ICluster interface {
		actor.IActor
		InitCluster(info *base.ClusterInfo, endpoints []string, natsUrl string, params ...OpOption)
		RegisterClusterCall() //注册集群通用回调
		AddCluster(info *base.ClusterInfo)
		DelCluster(info *base.ClusterInfo)
		GetCluster(rpc.RpcHead) *base.ClusterInfo
		BindPacketFunc(packetFunc network.PacketFunc)
		CallMsg(interface{}, rpc.RpcHead, string, ...interface{}) error //同步给集群特定服务器
		RandomCluster(head rpc.RpcHead) rpc.RpcHead                     //随机分配
		IsEnoughStub(stub rpc.STUB) bool
	}

	Cluster struct {
		actor.Actor
		*Service
		clusterMap     [MAX_CLUSTER_NUM]ClusterMap
		clusterLocker  [MAX_CLUSTER_NUM]*sync.RWMutex
		hashRing       [MAX_CLUSTER_NUM]*base.HashRing
		conn           *nats.Conn
		dieChan        chan bool
		master         *Master
		clusterInfoMap map[uint32]*base.ClusterInfo
		packetFuncList *vector.Vector[network.PacketFunc]
		MailBox        etcd.MailBox
		StubMailBox    etcd.StubMailBox
		Stub           common.Stub
	}

	EmptyClusterInfo struct {
		base.ClusterInfo
	}

	CallFunc struct {
		Func       interface{}
		FuncType   reflect.Type
		FuncVal    reflect.Value
		FuncParams string
	}
)

func (c *Cluster) InitCluster(info *base.ClusterInfo, endpoints []string, natsUrl string, params ...OpOption) {
	c.Actor.Init()
	for i := 0; i < MAX_CLUSTER_NUM; i++ {
		c.clusterLocker[i] = &sync.RWMutex{}
		c.clusterMap[i] = make(ClusterMap)
		c.hashRing[i] = base.NewHashRing()
	}
	c.clusterInfoMap = make(map[uint32]*base.ClusterInfo)
	c.packetFuncList = &vector.Vector[network.PacketFunc]{}

	conn, err := setupNatsConn(
		natsUrl,
		c.dieChan,
	)
	if err != nil {
		base.LOG.Fatalln("nats connect error!!!!")
	}
	c.conn = conn

	c.conn.Subscribe(getChannel(*info), func(msg *nats.Msg) {
		c.HandlePacket(rpc.Packet{Buff: msg.Data})
	})

	c.conn.Subscribe(getTopicChannel(*info), func(msg *nats.Msg) {
		c.HandlePacket(rpc.Packet{Buff: msg.Data})
	})

	c.conn.Subscribe(getCallChannel(*info), func(msg *nats.Msg) {
		c.HandlePacket(rpc.Packet{Buff: msg.Data, Reply: msg.Reply})
	})

	op := Op{}
	op.applyOpts(params)
	if len(op.mailBoxEndpoints) > 0 {
		c.MailBox.Init(info, op.mailBoxEndpoints)
	}
	if len(op.stubMailBoxEndpoints) > 0 {
		c.StubMailBox.Init(info, op.stubMailBoxEndpoints)
		c.Stub = op.stub
	}

	rpc.MGR = c
	actor.MGR.RegisterActor(c)
	c.Actor.Start()
	//注册服务器
	c.Service = NewService(info, endpoints)
	c.master = NewMaster(&EmptyClusterInfo{}, endpoints)
}

func (c *Cluster) AddCluster(info *base.ClusterInfo) {
	c.clusterLocker[info.Type].Lock()
	c.clusterMap[info.Type][info.Id()] = info
	c.clusterLocker[info.Type].Unlock()
	c.hashRing[info.Type].Add(info.IpString())
	base.LOG.Printf("服务器[%s:%s:%d]建立连接", info.String(), info.Ip, info.Port)
}

func (c *Cluster) DelCluster(info *base.ClusterInfo) {
	c.clusterLocker[info.Type].RLock()
	_, bEx := c.clusterMap[info.Type][info.Id()]
	c.clusterLocker[info.Type].RUnlock()
	if bEx {
		c.clusterLocker[info.Type].Lock()
		delete(c.clusterMap[info.Type], info.Id())
		c.clusterLocker[info.Type].Unlock()
	}

	c.hashRing[info.Type].Remove(info.IpString())
	base.LOG.Printf("服务器[%s:%s:%d]断开连接", info.String(), info.Ip, info.Port)
}

func (c *Cluster) GetCluster(head rpc.RpcHead) *base.ClusterInfo {
	c.clusterLocker[head.DestServerType].RLock()
	defer c.clusterLocker[head.DestServerType].RUnlock()
	client, bEx := c.clusterMap[head.DestServerType][head.ClusterId]
	if bEx {
		return client
	}
	return nil
}

func (c *Cluster) BindPacketFunc(callfunc network.PacketFunc) {
	c.packetFuncList.PushBack(callfunc)
}

func (c *Cluster) HandlePacket(packet rpc.Packet) {
	for _, v := range c.packetFuncList.Values() {
		if v(packet) {
			break
		}
	}
}

func (c *Cluster) SendMsg(head rpc.RpcHead, funcName string, params ...interface{}) {
	head.SrcClusterId = c.Id()
	c.Send(head, rpc.Marshal(&head, &funcName, params...))
}

func (c *Cluster) Send(head rpc.RpcHead, packet rpc.Packet) {
	switch head.SendType {
	//case rpc.SEND_BALANCE:
	//	_, head.ClusterId = c.hashRing[head.DestServerType].Get64(head.Id)
	//	c.conn.Publish(getRpcChannel(head), packet.Buff)
	case rpc.SEND_POINT:
		if head.ClusterId == 0 && head.DestServerType == rpc.SERVICE_GAME {
			pMailBox := c.MailBox.Get(head.Id)
			if pMailBox != nil {
				head.ClusterId = pMailBox.ClusterId
			}
		} else if head.ClusterId == 0 {
			stubCount, bEx := c.Stub.StubCount[head.ActorName]
			if bEx {
				index := head.Id % stubCount
				stubType := rpc.STUB(rpc.STUB_value[head.ActorName])
				pStub := c.StubMailBox.Get(stubType, index)
				if pStub != nil {
					head.ClusterId = pStub.ClusterId
				}
			}
		}
		c.conn.Publish(getRpcChannel(head), packet.Buff)
	default:
		c.conn.Publish(getRpcTopicChannel(head), packet.Buff)
	}
}

// params[0]:rpc.RpcHead
// params[1]:error
func (c *Cluster) Call(parmas ...interface{}) {
	head := *parmas[0].(*rpc.RpcHead)
	reply := head.Reply
	head.Reply = ""
	head.ClusterId = head.SrcClusterId
	if parmas[1] == nil {
		parmas[1] = ""
	} else {
		parmas[1] = parmas[1].(error).Error()
	}
	funcName := ""
	packet := rpc.Marshal(&head, &funcName, parmas[1:]...)
	c.conn.Publish(reply, packet.Buff)
}

func (c *Cluster) CallMsg(cb interface{}, head rpc.RpcHead, funcName string, params ...interface{}) error {
	head.SrcClusterId = c.Id()
	packet := rpc.Marshal(&head, &funcName, params...)

	switch head.SendType {
	//case rpc.SEND_BALANCE:
	//	_, head.ClusterId = c.hashRing[head.DestServerType].Get64(head.Id)
	//	c.conn.Publish(getRpcChannel(head), packet.Buff)
	case rpc.SEND_POINT:
		if head.ClusterId == 0 && head.DestServerType == rpc.SERVICE_GAME {
			pMailBox := c.MailBox.Get(head.Id)
			if pMailBox != nil {
				head.ClusterId = pMailBox.ClusterId
			}
		} else if head.ClusterId == 0 {
			stubCount, bEx := c.Stub.StubCount[head.ActorName]
			if bEx {
				index := head.Id % stubCount
				stubType := rpc.STUB(rpc.STUB_value[head.ActorName])
				pStub := c.StubMailBox.Get(stubType, index)
				if pStub != nil {
					head.ClusterId = pStub.ClusterId
				}
			}
		}
	default:
		base.LOG.Printf("CALL MSG [%s] CAN NOT BOARDCAST", funcName)
		//_, head.ClusterId = c.hashRing[head.DestServerType].Get64(head.Id)
	}

	reply, err := c.conn.Request(getRpcCallChannel(head), packet.Buff, CALL_TIME_OUT)
	if err == nil {
		rpcPacket, _ := rpc.Unmarshal(reply.Data)
		cf := &CallFunc{Func: cb, FuncVal: reflect.ValueOf(cb), FuncType: reflect.TypeOf(cb), FuncParams: reflect.TypeOf(cb).String()}
		f := cf.FuncVal
		k := cf.FuncType
		err, params := rpc.UnmarshalBodyCall(rpcPacket, k)
		if err != nil {
			return err
		}
		iLen := len(params)
		if iLen >= 1 {
			in := make([]reflect.Value, iLen)
			for i, param := range params {
				in[i] = reflect.ValueOf(param)
			}

			f.Call(in)
		} else {
			base.LOG.Printf("CallMsg [%s] params at least one context", funcName)
			return errors.New("callmsg params at least one context")
		}
	}
	return err
}

func (c *Cluster) RandomCluster(head rpc.RpcHead) rpc.RpcHead {
	if head.Id == 0 {
		head.Id = int64(uint32(base.RAND.RandI(1, 0xFFFFFFFF)))
	}
	_, head.ClusterId = c.hashRing[head.DestServerType].Get64(head.Id)
	pCluster := c.GetCluster(head)
	if pCluster != nil {
		head.SocketId = pCluster.SocketId
	}
	return head
}

func (c *Cluster) IsEnoughStub(stub rpc.STUB) bool {
	return c.StubMailBox.Count(stub) == c.Stub.StubCount[stub.String()]
}

// 集群新加member
func (c *Cluster) Cluster_Add(ctx context.Context, info *base.ClusterInfo) {
	_, bEx := c.clusterInfoMap[info.Id()]
	if !bEx {
		c.AddCluster(info)
		c.clusterInfoMap[info.Id()] = info
	}
}

// 集群删除member
func (c *Cluster) Cluster_Del(ctx context.Context, info *base.ClusterInfo) {
	delete(c.clusterInfoMap, info.Id())
	c.DelCluster(info)
}

func (op *Op) applyOpts(opts []OpOption) {
	for _, opt := range opts {
		opt(op)
	}
}

func WithMailBoxEtcd(Endpoints []string) OpOption {
	return func(op *Op) {
		op.mailBoxEndpoints = Endpoints
	}
}

func WithStubMailBoxEtcd(Endpoints []string, stub *common.Stub) OpOption {
	return func(op *Op) {
		op.stubMailBoxEndpoints = Endpoints
		op.stub = *stub
	}
}

func (c *EmptyClusterInfo) String() string {
	return ""
}
