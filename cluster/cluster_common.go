package cluster

import (
	"fmt"
	"strings"

	"github.com/fengqk/mars-base/cluster/etcd"
	"github.com/fengqk/mars-base/common"
	"github.com/fengqk/mars-base/rpc"
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

func getChannel(clusterInfo common.ClusterInfo) string {
	return fmt.Sprintf("%s/%s/%d", etcd.ETCD_DIR, clusterInfo.String(), clusterInfo.Id())
}

func getTopicChannel(clusterInfo common.ClusterInfo) string {
	return fmt.Sprintf("%s/%s", etcd.ETCD_DIR, clusterInfo.String())
}

func getCallChannel(clusterInfo common.ClusterInfo) string {
	return fmt.Sprintf("%s/%s/call/%d", etcd.ETCD_DIR, clusterInfo.String(), clusterInfo.Id())
}

func getRpcChannel(head rpc.RpcHead) string {
	return fmt.Sprintf("%s/%s/%d", etcd.ETCD_DIR, strings.ToLower(head.DestServerType.String()), head.ClusterId)
}

func getRpcTopicChannel(head rpc.RpcHead) string {
	return fmt.Sprintf("%s/%s", etcd.ETCD_DIR, strings.ToLower(head.DestServerType.String()))
}

func getRpcCallChannel(head rpc.RpcHead) string {
	return fmt.Sprintf("%s/%s/call/%d", etcd.ETCD_DIR, strings.ToLower(head.DestServerType.String()), head.ClusterId)
}
