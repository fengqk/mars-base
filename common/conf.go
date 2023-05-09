package common

import "github.com/fengqk/mars-base/rpc"

type (
	Server struct {
		Ip   string `yaml:"ip"`
		Port int    `yaml:"port"`
	}

	Mysql struct {
		Ip           string `yaml:"ip"`
		Port         int    `yaml:"port"`
		Name         string `yaml:"name"`
		User         string `yaml:"user"`
		Password     string `yaml:"password"`
		MaxIdleConns int    `yaml:"maxIdleConns"`
		MaxOpenConns int    `yaml:"maxOpenConns"`
	}

	Redis struct {
		Ip       string `yaml:"ip"`
		Port     int    `yaml:"port"`
		Password string `yaml:"password"`
	}

	Etcd struct {
		Endpoints []string `yaml:"endpoints"`
	}

	SnowFlake struct {
		Endpoints []string `yaml:"endpoints"`
	}

	Nats struct {
		Endpoints string `yaml:"endpoints"`
	}

	Raft struct {
		Endpoints []string `yaml:"endpoints"`
	}

	StubRouter struct {
		STUB rpc.STUB `yaml:"stub"`
	}

	Stub struct {
		StubCount map[string]int64 `yaml:"stub_count"`
	}
)
