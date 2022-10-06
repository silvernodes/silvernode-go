package cluster

import (
	"fmt"

	"github.com/hashicorp/consul/api"
	"github.com/silvernodes/silvernode-go/ctx"
	"github.com/silvernodes/silvernode-go/utils/errutil"
	"github.com/silvernodes/silvernode-go/utils/netutil"
	"github.com/silvernodes/silvernode-go/utils/yamlutil"
)

type ConsulInfo struct {
	NameSpace string
	IpAddress string
	Port      uint64
	Username  string
	Password  string
}

type ConsulIns struct {
	info   *ConsulInfo
	client *api.Client
}

func NewConsulIns() *ConsulIns {
	c := new(ConsulIns)
	c.info = new(ConsulInfo)
	return c
}

func (c *ConsulIns) Install() error {
	if err := ctx.CoreConf().GetConfDatas("cluster.consul", c.info); err != nil {
		return errutil.Extend("Consul注册中心配置信息加载失败", err)
	}
	config := api.DefaultConfig()
	config.Address = c.info.IpAddress + ":" + fmt.Sprint(c.info.Port)
	if c.info.Username != "" {
		config.HttpAuth.Username = c.info.Username
		config.HttpAuth.Password = c.info.Password
	}
	client, err := api.NewClient(config)
	if err != nil {
		return errutil.Extend("Consul注册中心装载错误", err)
	}
	c.client = client
	return nil
}

func (c *ConsulIns) RegNodeInfo(nodeInfo *ctx.NodeInfo) error {
	_, ip, _, err := netutil.ParseUrlInfo(nodeInfo.EndPoints[0])
	if err != nil {
		return errutil.Extend("解析节点注册信息发生错误", err)
	}
	str, err := nodeInfo.Marshal()
	if err != nil {
		return errutil.Extend("节点注册并序列化时发生错误", err)
	}

	registration := new(api.AgentServiceRegistration)
	registration.ID = nodeInfo.NodeId // 服务节点的名称
	registration.Name = nodeInfo.Name // 服务名称
	// registration.Namespace = c.info.NameSpace
	registration.Port = int(nodeInfo.MainPort)     // 服务端口
	registration.Tags = []string{c.info.NameSpace} // tag，可以为空
	registration.Address = ip                      // 服务 IP
	registration.Meta = map[string]string{
		"info": str,
	}
	checkUrl := fmt.Sprintf("%s:%d", registration.Address, nodeInfo.MainPort)
	registration.Check = &api.AgentServiceCheck{ // 健康检查
		HTTP:                           "http://" + checkUrl + "/",
		Timeout:                        "3s",
		Interval:                       "5s",  // 健康检查间隔
		DeregisterCriticalServiceAfter: "30s", //check失败后30秒删除本服务，注销时间，相当于过期时间
	}
	if err = c.client.Agent().ServiceRegister(registration); err != nil {
		return errutil.Extend("节点注册时发生错误", err)
	}
	return nil
}
func (c *ConsulIns) GetNodeById(nodeId string) (*ctx.NodeInfo, error) {
	q := &api.QueryOptions{}
	svc, _, err := c.client.Agent().Service(nodeId, q)
	if err != nil {
		return nil, errutil.Extend("获取节点信息发生错误:"+nodeId, err)
	}
	str, exists := svc.Meta["info"]
	if !exists {
		return nil, errutil.Extend("获取节点meta信息发生错误:"+nodeId, err)
	}
	nodeInfo := ctx.NewNodeInfo()
	if err := nodeInfo.Unmarshal(str); err != nil {
		return nil, errutil.Extend("解析节点meta信息发生错误:"+nodeId, err)
	}
	return nodeInfo, nil
}
func (c *ConsulIns) SelectNodesByName(name string) ([]*ctx.NodeInfo, error) {
	q := &api.QueryOptions{}
	svcs, _, err := c.client.Catalog().Service(name, c.info.NameSpace, q)
	if err != nil {
		return nil, errutil.Extend("筛选节点信息发生错误:"+name, err)
	}
	nodeInfos := make([]*ctx.NodeInfo, 0, 0)
	for _, svc := range svcs {
		str, exists := svc.ServiceMeta["info"]
		if !exists {
			return nil, errutil.Extend("获取节点meta信息发生错误:"+svc.ID, err)
		}
		nodeInfo := ctx.NewNodeInfo()
		if err := nodeInfo.Unmarshal(str); err != nil {
			return nil, errutil.Extend("解析节点meta信息发生错误:"+svc.ID, err)
		}
		nodeInfos = append(nodeInfos, nodeInfo)
	}
	return nodeInfos, nil
}
func (c *ConsulIns) CheckNodeSig(nodeId string, sig string) (bool, error) {
	nodeInfo, err := c.GetNodeById(nodeId)
	if err != nil {
		return false, errutil.Extend("节点验签失败:"+nodeId+"<--->"+sig, err)
	}
	return nodeInfo.Sig == sig, nil
}
func (c *ConsulIns) SetConfig(key string, val interface{}) error {
	data, err := yamlutil.MarshalRaw(val)
	if err != nil {
		return errutil.Extend("对象写入配置中心时序列化出错", err)
	}
	q := &api.WriteOptions{}
	kv := &api.KVPair{
		Key:   key,
		Value: data,
	}
	if _, err := c.client.KV().Put(kv, q); err != nil {
		return errutil.Extend("对象写入配置中心时出错", err)
	}
	return nil
}
func (c *ConsulIns) GetConfig(key string, ref interface{}) error {
	q := &api.QueryOptions{}
	kv, _, err := c.client.KV().Get(key, q)
	if err != nil {
		return errutil.Extend("对象从配置中心读取时出错", err)
	}
	if err := yamlutil.UnmarshalRaw(kv.Value, ref); err != nil {
		return errutil.Extend("对象从配置中心读取反序列化时出错", err)
	}
	return nil
}
func (c *ConsulIns) DelConfig(key string) error {
	q := &api.WriteOptions{
		Namespace: c.info.NameSpace,
	}
	if _, err := c.client.KV().Delete(key, q); err != nil {
		return errutil.Extend("对象从配置中心删除时出错", err)
	}
	return nil
}
