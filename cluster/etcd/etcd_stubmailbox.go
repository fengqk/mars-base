package etcd

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/fengqk/mars-base/common"
	"github.com/fengqk/mars-base/rpc"
	clientv3 "go.etcd.io/etcd/client/v3"
)

const (
	STUB_DIR      = "stub/"
	STUB_TTL_TIME = 30
)

type (
	StubMailBoxMap map[int64]*common.StubMailBox

	StubMailBox struct {
		*common.ClusterInfo
		client            *clientv3.Client
		lease             clientv3.Lease
		stubMailBoxMap    [rpc.STUB_END]StubMailBoxMap
		stubMailBoxLocker [rpc.STUB_END]*sync.RWMutex
	}
)

func (s *StubMailBox) Create(info *common.StubMailBox) bool {
	leaseResp, err := s.lease.Grant(context.Background(), STUB_TTL_TIME)
	if err == nil {
		leaseId := leaseResp.ID
		info.LeaseId = int64(leaseId)
		key := fmt.Sprintf("%s%s", STUB_DIR, info.Key())
		data, _ := json.Marshal(info)
		//设置key
		tx := s.client.Txn(context.Background())
		tx.If(clientv3.Compare(clientv3.CreateRevision(key), "=", 0)).
			Then(clientv3.OpPut(key, string(data), clientv3.WithLease(leaseId))).
			Else()
		txnRes, err := tx.Commit()
		return err == nil && txnRes.Succeeded
	}
	return false
}

func (s *StubMailBox) Init(info *common.ClusterInfo, endpoints []string) {
	cfg := clientv3.Config{
		Endpoints: endpoints,
	}

	etcdClient, err := clientv3.New(cfg)
	if err != nil {
		log.Fatal("cannot connec to etcd:", err)
	}
	s.ClusterInfo = info
	lease := clientv3.NewLease(etcdClient)
	s.client = etcdClient
	s.lease = lease
	for i := 0; i < int(rpc.STUB_END); i++ {
		s.stubMailBoxLocker[i] = &sync.RWMutex{}
		s.stubMailBoxMap[i] = make(StubMailBoxMap)
	}
	s.Start()
}

func (s *StubMailBox) Start() {
	go s.Run()
}

func (s *StubMailBox) Run() {
	wch := s.client.Watch(context.Background(), STUB_DIR, clientv3.WithPrefix(), clientv3.WithPrevKV())
	s.getAll()
	for v := range wch {
		for _, v1 := range v.Events {
			if v1.Type.String() == "PUT" {
				info := nodeToStubMailBox(v1.Kv.Value)
				s.add(info)
			} else {
				info := nodeToStubMailBox(v1.PrevKv.Value)
				s.del(info)
			}
		}
	}
}

func (s *StubMailBox) Lease(info *common.StubMailBox) error {
	_, err := s.lease.KeepAliveOnce(context.Background(), clientv3.LeaseID(info.LeaseId))
	return err
}

func (s *StubMailBox) Get(stubType rpc.STUB, Id int64) *common.StubMailBox {
	s.stubMailBoxLocker[stubType].RLock()
	stub, bEx := s.stubMailBoxMap[stubType][Id]
	s.stubMailBoxLocker[stubType].RUnlock()
	if bEx {
		return stub
	}
	return nil
}

func (s *StubMailBox) Count(stubType rpc.STUB) int64 {
	s.stubMailBoxLocker[stubType].RLock()
	nLen := len(s.stubMailBoxMap[stubType])
	s.stubMailBoxLocker[stubType].RUnlock()
	return int64(nLen)
}

func (s *StubMailBox) add(info *common.StubMailBox) {
	s.stubMailBoxLocker[info.StubType].Lock()
	stub, bOk := s.stubMailBoxMap[info.StubType][info.Id]
	if !bOk {
		s.stubMailBoxMap[info.StubType][info.Id] = info
	} else {
		*stub = *info
	}
	s.stubMailBoxLocker[info.StubType].Unlock()
}

func (s *StubMailBox) del(info *common.StubMailBox) {
	s.stubMailBoxLocker[info.StubType].Lock()
	delete(s.stubMailBoxMap[info.StubType], info.Id)
	s.stubMailBoxLocker[info.StubType].Unlock()
}

func (s *StubMailBox) getAll() {
	resp, err := s.client.Get(context.Background(), STUB_DIR, clientv3.WithPrefix())
	if err == nil && (resp != nil && resp.Kvs != nil) {
		for _, v := range resp.Kvs {
			info := nodeToStubMailBox(v.Value)
			s.add(info)
		}
	}
}

func nodeToStubMailBox(val []byte) *common.StubMailBox {
	info := &common.StubMailBox{}
	err := json.Unmarshal([]byte(val), info)
	if err != nil {
		log.Print(err)
	}
	return info
}
