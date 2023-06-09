package common

import (
	"fmt"
	"strings"

	"github.com/fengqk/mars-base/base"
	"github.com/fengqk/mars-base/rpc"
)

type (
	IClusterInfo interface {
		Id() uint32
		String() string
		ServiceType() rpc.SERVICE
		IpString() string
	}

	ClusterInfo rpc.ClusterInfo

	StubMailBox struct {
		rpc.StubMailBox
	}
)

func (c *ClusterInfo) Id() uint32 {
	return base.ToHash(c.IpString())
}

func (c *ClusterInfo) String() string {
	return strings.ToLower(c.Type.String())
}

func (c *ClusterInfo) ServiceType() rpc.SERVICE {
	return c.Type
}

func (c *ClusterInfo) IpString() string {
	return fmt.Sprintf("%s:%d", c.Ip, c.Port)
}

func (s *StubMailBox) StubName() string {
	return s.StubType.String()
}

func (s *StubMailBox) Key() string {
	return fmt.Sprintf("%s/%d", s.StubType.String(), s.Id)
}
