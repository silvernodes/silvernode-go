package cluster

import (
	"github.com/silvernodes/silvernode-go/ctx"
	"github.com/silvernodes/silvernode-go/process"
	"github.com/silvernodes/silvernode-go/utils/errutil"
)

const (
	Consul string = "consul"
	Etcd          = "etcd"
	Nacos         = "nacos"
)

type IRegistry interface {
	RegNodeInfo(nodeInfo *ctx.NodeInfo) error
	GetNodeById(nodeId string) (*ctx.NodeInfo, error)
	SelectNodesByName(name string) ([]*ctx.NodeInfo, error)
	CheckNodeSig(nodeId string, sig string) (bool, error)
	SetConfig(key string, val interface{}) error
	GetConfig(key string, ref interface{}) error
	DelConfig(key string) error
}

func CreateRegistry(Type string) (IRegistry, error) {
	if Type == Consul {
		consul := NewConsulIns()
		if err := consul.Install(); err != nil {
			return nil, err
		}
		return consul, nil
	}
	if Type == Etcd {
		etcd := NewEtcdIns()
		if err := etcd.Install(); err != nil {
			return nil, err
		}
		return etcd, nil
	}
	if Type == Nacos {
		nacos := NewNacosIns()
		if err := nacos.Install(); err != nil {
			return nil, err
		}
		return nacos, nil
	}
	return nil, errutil.New("错误的注册中心类型:" + Type)
}

type ClusterParam struct {
	SelfInfo   *ctx.NodeInfo
	OnScanning func(otherInfos []*ctx.NodeInfo, err error)
}

var _service process.Service
var _registry IRegistry
var _param *ClusterParam

func Serve(registry IRegistry, param *ClusterParam) error {
	_registry = registry
	_param = param
	if err := registry.RegNodeInfo(param.SelfInfo); err != nil {
		return err
	}
	if len(param.SelfInfo.BackEnds) > 0 {
		_service = process.SpawnS()
		_service.Start(nodeScanning, nil)
	}
	return nil
}

func nodeScanning() {
	for _, backend := range _param.SelfInfo.BackEnds {
		otherInfos, err := _registry.SelectNodesByName(backend)
		_param.OnScanning(otherInfos, err)
		process.Sleep(2000)
	}
	process.Sleep(1000)
}
