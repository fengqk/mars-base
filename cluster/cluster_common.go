package cluster

import (
	"fmt"
	"strings"

	"github.com/fengqk/mars-base/base"
	"github.com/fengqk/mars-base/cluster/etcd"
	"github.com/fengqk/mars-base/rpc"
	"github.com/nats-io/nats.go"
)

type (
	Master    etcd.Master
	Service   etcd.Service
	Snowflake etcd.Snowflake
)

func NewMaster(info base.IClusterInfo, endpoints []string) *Master {
	master := &etcd.Master{}
	master.Init(info, endpoints)
	return (*Master)(master)
}

func NewService(info *base.ClusterInfo, endpoints []string) *Service {
	service := &etcd.Service{}
	service.Init(info, endpoints)
	return (*Service)(service)
}

func NewSnowflake(endpoints []string) *Snowflake {
	uuid := &etcd.Snowflake{}
	uuid.Init(endpoints)
	return (*Snowflake)(uuid)
}

func getChannel(clusterInfo base.ClusterInfo) string {
	return fmt.Sprintf("%s/%s/%d", etcd.ETCD_DIR, clusterInfo.String(), clusterInfo.Id())
}

func getTopicChannel(clusterInfo base.ClusterInfo) string {
	return fmt.Sprintf("%s/%s", etcd.ETCD_DIR, clusterInfo.String())
}

func getCallChannel(clusterInfo base.ClusterInfo) string {
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

func setupNatsConn(connectString string, appDieChan chan bool, options ...nats.Option) (*nats.Conn, error) {
	natsOptions := append(
		options,
		nats.DisconnectHandler(func(_ *nats.Conn) {
			base.LOG.Println("disconnected from nats!")
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			base.LOG.Printf("reconnected to nats server %s with address %s in cluster %s!", nc.ConnectedServerId(), nc.ConnectedAddr(), nc.ConnectedUrl())
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			err := nc.LastError()
			if err == nil {
				base.LOG.Println("nats connection closed with no error.")
				return
			}

			base.LOG.Println("nats connection closed. reason: %q", nc.LastError())
			if appDieChan != nil {
				appDieChan <- true
			}
		}),
	)

	nc, err := nats.Connect(connectString, natsOptions...)
	if err != nil {
		return nil, err
	}
	return nc, nil
}
