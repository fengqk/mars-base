package cluster

import (
	"github.com/fengqk/mars-base/cluster/etcd"
	"github.com/fengqk/mars-base/common"
)

type (
	Master    etcd.Master
	Service   etcd.Service
	Snowflake etcd.Snowflake
)

func NewMaster(info common.IClusterInfo, endpoints []string) *Master {
	master := &etcd.Master{}
	master.Init(info, endpoints)
	return (*Master)(master)
}

func NewService(info *common.ClusterInfo, endpoints []string) *Service {
	service := &etcd.Service{}
	service.Init(info, endpoints)
	return (*Service)(service)
}

func NewSnowflake(endpoints []string) *Snowflake {
	uuid := &etcd.Snowflake{}
	uuid.Init(endpoints)
	return (*Snowflake)(uuid)
}
