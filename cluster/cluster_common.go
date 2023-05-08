package cluster

import (
	"fmt"
	"strings"

	"github.com/fengqk/mars-base/common"
	"github.com/fengqk/mars-base/rpc"
)

type (
	ClusterInfo rpc.ClusterInfo

	IClusterInfo interface {
		Id() uint32
		String() string
		ServerType() rpc.SERVICE
		IpString() string
	}
)

func (c *ClusterInfo) IpString() string {
	return fmt.Sprintf("%s:%d", c.Ip, c.Port)
}

func (c *ClusterInfo) String() string {
	return strings.ToLower(c.Type.String())
}

func (c *ClusterInfo) Id() uint32 {
	return common.ToHash(c.IpString())
}

func (c *ClusterInfo) ServiceType() rpc.SERVICE {
	return c.Type
}
