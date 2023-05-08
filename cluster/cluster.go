package cluster

import (
	"time"

	"github.com/fengqk/mars-base/rpc"
)

const (
	MAX_CLUSTER_NUM = int(rpc.SERVICE_NUM)
	CALL_TIME_OUT   = 500 * time.Millisecond
)

type ()
