package common

import (
	"fmt"
	"strings"

	"github.com/fengqk/mars-base/rpc"
)

type (
	IClusterInfo interface {
		Id() uint32
		String() string
		ServerType() rpc.SERVICE
		IpString() string
	}

	ClusterInfo rpc.ClusterInfo
	StubMailBox rpc.StubMailBox
)

func (c *ClusterInfo) IpString() string {
	return fmt.Sprintf("%s:%d", c.Ip, c.Port)
}

func (c *ClusterInfo) String() string {
	return strings.ToLower(c.Type.String())
}

func (c *ClusterInfo) Id() uint32 {
	return ToHash(c.IpString())
}

func (c *ClusterInfo) ServiceType() rpc.SERVICE {
	return c.Type
}

func (s *StubMailBox) StubName() string {
	return s.StubType.String()
}

func (s *StubMailBox) Key() string {
	return fmt.Sprintf("%s/%d", s.StubType.String(), s.Id)
}
