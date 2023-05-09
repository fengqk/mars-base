package etcd

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/fengqk/mars-base/base"
	clientv3 "go.etcd.io/etcd/client/v3"
)

const (
	UUID_DIR = "uuid/"
	TTL_TIME = 1800
)

const (
	SET STATUS = iota
	TTL STATUS = iota
)

type (
	STATUS uint32

	Snowflake struct {
		id      int64
		client  *clientv3.Client
		lease   clientv3.Lease
		leaseId clientv3.LeaseID
		status  STATUS
	}
)

func (s *Snowflake) Init(endpoints []string) {
	cfg := clientv3.Config{
		Endpoints: endpoints,
	}

	etcdClient, err := clientv3.New(cfg)
	if err != nil {
		log.Fatal("cannot connec to etcd:", err)
	}
	lease := clientv3.NewLease(etcdClient)
	s.id = int64(base.RAND.RandI(1, int(base.WorkeridMax)))
	s.client = etcdClient
	s.lease = lease
	for !s.SET() {
	}
	s.Start()
}

func (s *Snowflake) Start() {
	go s.Run()
}

func (s *Snowflake) Run() {
	for {
		switch s.status {
		case SET:
			s.SET()
		case TTL:
			s.TTL()
		}
	}
}

func (s *Snowflake) Key() string {
	return UUID_DIR + fmt.Sprintf("%d", s.id)
}

func (s *Snowflake) SET() bool {
	key := s.Key()
	tx := s.client.Txn(context.Background())
	//key no exist
	leaseResp, err := s.lease.Grant(context.Background(), TTL_TIME)
	if err != nil {
		return false
	}
	s.leaseId = leaseResp.ID
	tx.If(clientv3.Compare(clientv3.CreateRevision(key), "=", 0)).
		Then(clientv3.OpPut(key, "", clientv3.WithLease(s.leaseId))).
		Else()
	txnRes, err := tx.Commit()
	if err != nil || !txnRes.Succeeded { //抢锁失败
		s.id = int64(base.RAND.RandI(1, int(base.WorkeridMax)))
		return false
	}

	base.UUID.Init(s.id) //设置uuid
	s.status = TTL
	return true
}

func (s *Snowflake) TTL() {
	//保持ttl
	_, err := s.lease.KeepAliveOnce(context.Background(), s.leaseId)
	if err != nil {
		s.status = SET
	} else {
		time.Sleep(TTL_TIME / 3)
	}
}
