package cluster

import (
	"github.com/silvernodes/silvernode-go/ctx"
	"github.com/silvernodes/silvernode-go/utils/errutil"
	"github.com/silvernodes/silvernode-go/utils/netutil"
	"github.com/silvernodes/silvernode-go/utils/yamlutil"

	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
)

type NacosInfo struct {
	NameSpace string
	IpAddress string
	Port      uint64
	Username  string
	Password  string
}

type NacosIns struct {
	info       *NacosInfo
	client     naming_client.INamingClient
	confClient config_client.IConfigClient
}

func NewNacosIns() *NacosIns {
	n := new(NacosIns)
	n.info = new(NacosInfo)
	return n
}

func (n *NacosIns) Install() error {
	if err := ctx.CoreConf().GetConfDatas("cluster.nacos", n.info); err != nil {
		return errutil.Extend("Nacos注册中心配置信息加载失败", err)
	}

	sc := []constant.ServerConfig{
		{
			IpAddr:      n.info.IpAddress,
			Port:        n.info.Port,
			ContextPath: "/nacos",
		},
	}
	cc := constant.ClientConfig{
		NamespaceId:         n.info.NameSpace,
		TimeoutMs:           5000,
		NotLoadCacheAtStart: true,
		LogLevel:            "debug",
	}
	if n.info.Username != "" {
		cc.Username = n.info.Username
		cc.Password = n.info.Password
	}
	client, err := clients.NewNamingClient(
		vo.NacosClientParam{
			ClientConfig:  &cc,
			ServerConfigs: sc,
		},
	)
	if err != nil {
		return errutil.Extend("Nacos注册中心装载错误", err)
	}
	confClient, err := clients.NewConfigClient(
		vo.NacosClientParam{
			ClientConfig:  &cc,
			ServerConfigs: sc,
		},
	)
	if err != nil {
		return errutil.Extend("Nacos配置中心装载错误", err)
	}
	n.client = client
	n.confClient = confClient
	return nil
}

func (n *NacosIns) RegNodeInfo(nodeInfo *ctx.NodeInfo) error {
	_, ip, _, err := netutil.ParseUrlInfo(nodeInfo.EndPoints[0])
	if err != nil {
		return errutil.Extend("解析节点注册信息发生错误", err)
	}
	str, err := nodeInfo.Marshal()
	if err != nil {
		return errutil.Extend("节点注册并序列化时发生错误", err)
	}
	p := vo.RegisterInstanceParam{
		Ip:          ip,
		Port:        nodeInfo.MainPort,
		ServiceName: nodeInfo.Name,
		Weight:      50,
		GroupName:   n.info.NameSpace,
		Enable:      true,
		Healthy:     true,
		Ephemeral:   false,
		Metadata: map[string]string{
			"nodeid": nodeInfo.NodeId,
			"info":   str,
		},
	}
	success, err := n.client.RegisterInstance(p)
	if err != nil {
		return errutil.Extend("节点注册时发生错误", err)
	} else if !success {
		return errutil.New("节点注册失败")
	}
	return nil
}
func (n *NacosIns) GetNodeById(nodeId string) (*ctx.NodeInfo, error) {
	p := vo.SelectInstancesParam{
		ServiceName: ctx.GetNodeNameFromId(nodeId),
		GroupName:   n.info.NameSpace,
		HealthyOnly: true,
	}
	svcs, err := n.client.SelectInstances(p)
	if err != nil {
		return nil, errutil.Extend("获取节点信息发生错误:"+nodeId, err)
	}
	for _, svc := range svcs {
		nodeid, exists := svc.Metadata["nodeid"]
		if !exists {
			return nil, errutil.Extend("获取节点nodeid信息发生错误:"+nodeId, err)
		}
		if nodeId == nodeid && svc.Healthy {
			str, exists := svc.Metadata["info"]
			if !exists {
				return nil, errutil.Extend("获取节点meta信息发生错误:"+nodeId, err)
			}
			nodeInfo := ctx.NewNodeInfo()
			if err := nodeInfo.Unmarshal(str); err != nil {
				return nil, errutil.Extend("解析节点meta信息发生错误:"+nodeId, err)
			}
			return nodeInfo, nil
		}
	}
	return nil, errutil.New("未能找到对应的节点信息:" + nodeId)
}
func (n *NacosIns) SelectNodesByName(name string) ([]*ctx.NodeInfo, error) {
	p := vo.SelectInstancesParam{
		ServiceName: name,
		GroupName:   n.info.NameSpace,
		HealthyOnly: true,
	}
	svcs, err := n.client.SelectInstances(p)
	if err != nil {
		return nil, errutil.Extend("筛选节点信息发生错误:"+name, err)
	}
	nodeInfos := make([]*ctx.NodeInfo, 0, 0)
	for _, svc := range svcs {
		str, exists := svc.Metadata["info"]
		if !exists {
			return nil, errutil.Extend("获取节点meta信息发生错误:"+svc.ServiceName, err)
		}
		nodeInfo := ctx.NewNodeInfo()
		if err := nodeInfo.Unmarshal(str); err != nil {
			return nil, errutil.Extend("解析节点meta信息发生错误:"+svc.ServiceName, err)
		}
		nodeInfos = append(nodeInfos, nodeInfo)
	}
	return nodeInfos, nil
}
func (n *NacosIns) CheckNodeSig(nodeId string, sig string) (bool, error) {
	nodeInfo, err := n.GetNodeById(nodeId)
	if err != nil {
		return false, errutil.Extend("节点验签失败:"+nodeId+"<--->"+sig, err)
	}
	return nodeInfo.Sig == sig, nil
}
func (n *NacosIns) SetConfig(key string, val interface{}) error {
	str, err := yamlutil.Marshal(val)
	if err != nil {
		return errutil.Extend("对象写入配置中心时序列化出错", err)
	}
	_, err2 := n.confClient.PublishConfig(vo.ConfigParam{
		DataId:  key,
		Group:   "DEFAULT_GROUP",
		Content: str,
	})
	if err2 != nil {
		return errutil.Extend("对象写入配置中心时出错", err2)
	}
	return nil
}
func (n *NacosIns) GetConfig(key string, ref interface{}) error {
	str, err := n.confClient.GetConfig(vo.ConfigParam{
		DataId: key,
		Group:  "DEFAULT_GROUP",
	})
	if err != nil {
		return errutil.Extend("对象从配置中心读取时出错", err)
	}
	if err := yamlutil.Unmarshal(str, ref); err != nil {
		return errutil.Extend("对象从配置中心读取反序列化时出错", err)
	}
	return nil
}
func (n *NacosIns) DelConfig(key string) error {
	_, err := n.confClient.DeleteConfig(vo.ConfigParam{
		DataId: key,
	})
	if err != nil {
		return errutil.Extend("对象从配置中心删除时出错", err)
	}
	return nil
}
