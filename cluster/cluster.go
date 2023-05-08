package cluster

import (
	"time"

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
)
