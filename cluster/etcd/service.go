package etcd

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/fengqk/mars-base/cluster"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type (
	Service struct {
		*cluster.ClusterInfo
		client  *clientv3.Client
		lease   clientv3.Lease
		leaseId clientv3.LeaseID
		status  STATUS
	}
)

func (s *Service) Init(info *cluster.ClusterInfo, endpoints []string) {
	cfg := clientv3.Config{
		Endpoints: endpoints,
	}

	etcdClient, err := clientv3.New(cfg)
	if err != nil {
		log.Fatal("cannot connec to etcd:", err)
	}
	lease := clientv3.NewLease(etcdClient)
	s.client = etcdClient
	s.lease = lease
	s.ClusterInfo = info
	s.Start()
}

func (s *Service) Start() {
	go s.Run()
}

func (s *Service) Run() {
	for {
		switch s.status {
		case SET:
			s.SET()
		case TTL:
			s.TTL()
		}
	}
}

func (s *Service) SET() {
	leaseResp, _ := s.lease.Grant(context.Background(), 10)
	s.leaseId = leaseResp.ID
	key := ETCD_DIR + s.String() + "/" + s.IpString()
	data, _ := json.Marshal(s.ClusterInfo)
	s.client.Put(context.Background(), key, string(data), clientv3.WithLease(s.leaseId))
	s.status = TTL
	time.Sleep(time.Second * 3)
}

func (s *Service) TTL() {
	//保持ttl
	_, err := s.lease.KeepAliveOnce(context.Background(), s.leaseId)
	if err != nil {
		s.status = SET
	} else {
		time.Sleep(time.Second * 3)
	}
}
