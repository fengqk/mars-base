package cluster

import (
	"sync"
	"time"

	"github.com/fengqk/mars-base/actor"
	"github.com/fengqk/mars-base/common"
	"github.com/fengqk/mars-base/rpc"
)

const (
	MAX_CLUSTER_NUM = int(rpc.SERVICE_NUM)
	CALL_TIME_OUT   = 500 * time.Millisecond
)

type (
	ClusterMap       map[uint32]*common.ClusterInfo
	ClusterSocketMap map[uint32]*common.ClusterInfo

	Op struct {
		mailBoxEndpoints     []string
		stubMailBoxEndpoints []string
		stub                 common.Stub
	}

	OpOption func(*Op)

	ICluster interface {
		actor.IActor
		InitCluster(info *common.ClusterInfo, endpoints []string, natsUrl string, params ...OpOption)
	}

	Cluster struct {
		actor.Actor
		*Service
		clusterMap    [MAX_CLUSTER_NUM]ClusterMap
		clusterLocker [MAX_CLUSTER_NUM]*sync.RWMutex
		hashRing      [MAX_CLUSTER_NUM]*common.hashRing
	}
)
