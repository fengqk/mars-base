package etcd

import (
	"context"
	"encoding/json"
	"log"

	"github.com/fengqk/mars-base/actor"
	"github.com/fengqk/mars-base/base"
	"github.com/fengqk/mars-base/rpc"
	clientv3 "go.etcd.io/etcd/client/v3"
)

const (
	ETCD_DIR = "server/"
)

type (
	Master struct {
		base.IClusterInfo
		client *clientv3.Client
	}
)

func (m *Master) Init(info base.IClusterInfo, endpoints []string) {
	cfg := clientv3.Config{
		Endpoints: endpoints,
	}

	etcdClient, err := clientv3.New(cfg)
	if err != nil {
		log.Fatal("cannot connect to etcd:", err)
	}

	m.client = etcdClient
	m.IClusterInfo = info
	m.Start()
	m.InitServices()
}

func (m *Master) Start() {
	go m.Run()
}

func (m *Master) Run() {
	wch := m.client.Watch(context.Background(), ETCD_DIR+m.String(), clientv3.WithPrefix(), clientv3.WithPrevKV())
	for v := range wch {
		for _, v1 := range v.Events {
			if v1.Type.String() == "PUT" {
				info := nodeToService(v1.Kv.Value)
				m.addService(info)
			} else {
				info := nodeToService(v1.PrevKv.Value)
				m.delService(info)
			}
		}
	}
}

func (m *Master) InitServices() {
	resp, err := m.client.Get(context.Background(), ETCD_DIR, clientv3.WithPrefix())
	if err == nil && (resp != nil && resp.Kvs != nil) {
		for _, v := range resp.Kvs {
			info := nodeToService(v.Value)
			m.addService(info)
		}
	}
}

func (m *Master) addService(info *base.ClusterInfo) {
	actor.MGR.SendMsg(rpc.RpcHead{}, "Cluster.Cluster_Add", info)
}

func (m *Master) delService(info *base.ClusterInfo) {
	actor.MGR.SendMsg(rpc.RpcHead{}, "Cluster.Cluster_Del", info)
}

func nodeToService(val []byte) *base.ClusterInfo {
	info := &base.ClusterInfo{}
	err := json.Unmarshal([]byte(val), info)
	if err != nil {
		log.Print(err)
	}
	return info
}
