package cluster

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/silvernodes/silvernode-go/ctx"
	"github.com/silvernodes/silvernode-go/utils/errutil"
	"github.com/silvernodes/silvernode-go/utils/yamlutil"

	"go.etcd.io/etcd/clientv3"
)

type EtcdInfo struct {
	NameSpace string
	IpAddress string
	Port      uint64
	Username  string
	Password  string
}

type EtcdIns struct {
	info      *EtcdInfo
	client    *clientv3.Client
	rwTimeout time.Duration
}

func NewEtcdIns() *EtcdIns {
	e := new(EtcdIns)
	e.info = new(EtcdInfo)
	e.rwTimeout = 3 * time.Second
	return e
}

func (e *EtcdIns) Install() error {
	if err := ctx.CoreConf().GetConfDatas("cluster.etcd", e.info); err != nil {
		return errutil.Extend("Etcd注册中心配置信息加载失败", err)
	}
	endpoints := []string{e.info.IpAddress + ":" + fmt.Sprint(e.info.Port)}
	config := clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	}
	if e.info.Username != "" {
		config.Username = e.info.Username
		config.Password = e.info.Password
	}
	client, err := clientv3.New(config)
	if err != nil {
		return errutil.Extend("Etcd注册中心装载错误", err)
	}
	e.client = client
	return nil
}

func (e *EtcdIns) svcPath() string {
	return e.info.NameSpace + "/Services/"
}

func (e *EtcdIns) kvsPath() string {
	return e.info.NameSpace + "/KeyValues/"
}

func (e *EtcdIns) RegNodeInfo(nodeInfo *ctx.NodeInfo) error {
	str, err := nodeInfo.Marshal()
	if err != nil {
		return errutil.Extend("节点注册并序列化时发生错误", err)
	}
	//申请一个5秒的租约
	lease := clientv3.NewLease(e.client)
	leaseGrantResp, err := lease.Grant(context.TODO(), 5)
	if err != nil {
		return errutil.Extend("申请Etcd设备租约失败", err)
	}
	leaseId := leaseGrantResp.ID
	//启动自动续租
	if keepChan, err := lease.KeepAlive(context.TODO(), leaseId); err != nil {
		return errutil.Extend("启用Etcd自动续租失败", err)
	} else {
		//处理续租应答的协程
		go func() {
			for {
				<-keepChan
			}
		}()
	}
	ctx, cancel := context.WithTimeout(context.Background(), e.rwTimeout)
	_, err2 := e.client.Put(ctx, e.svcPath()+nodeInfo.Name+"/"+nodeInfo.NodeId, str, clientv3.WithLease(leaseId))
	cancel()
	if err2 != nil {
		return errutil.Extend("节点注册时发生错误", err2)
	}
	return nil
}
func (e *EtcdIns) GetNodeById(nodeId string) (*ctx.NodeInfo, error) {
	nodeName := ctx.GetNodeNameFromId(nodeId)
	path := e.svcPath() + nodeName + "/" + nodeId
	c, cancel := context.WithTimeout(context.Background(), e.rwTimeout)
	resp, err := e.client.Get(c, path)
	cancel()
	if err != nil {
		return nil, errutil.Extend("获取节点信息发生错误:"+nodeId, err)
	}
	for _, kv := range resp.Kvs {
		if string(kv.Key) == path {
			str := string(kv.Value)
			nodeInfo := ctx.NewNodeInfo()
			if err := nodeInfo.Unmarshal(str); err != nil {
				return nil, errutil.Extend("解析节点meta信息发生错误:"+nodeId, err)
			}
			return nodeInfo, nil
		}
	}
	return nil, errutil.New("查找不到对应的节点信息:" + nodeId)
}
func (e *EtcdIns) SelectNodesByName(name string) ([]*ctx.NodeInfo, error) {
	path := e.svcPath() + name + "/"
	c, cancel := context.WithTimeout(context.Background(), e.rwTimeout)
	resp, err := e.client.Get(c, path, clientv3.WithPrefix())
	cancel()
	if err != nil {
		return nil, errutil.Extend("筛选节点信息发生错误:"+name, err)
	}
	nodeInfos := make([]*ctx.NodeInfo, 0, 0)
	for _, kv := range resp.Kvs {
		key := string(kv.Key)
		nodeId := strings.TrimPrefix(key, path)
		str := string(kv.Value)
		nodeInfo := ctx.NewNodeInfo()
		if err := nodeInfo.Unmarshal(str); err != nil {
			return nil, errutil.Extend("解析节点meta信息发生错误:"+nodeId, err)
		}
		nodeInfos = append(nodeInfos, nodeInfo)
	}
	return nodeInfos, nil
}
func (e *EtcdIns) CheckNodeSig(nodeId string, sig string) (bool, error) {
	nodeInfo, err := e.GetNodeById(nodeId)
	if err != nil {
		return false, errutil.Extend("节点验签失败:"+nodeId+"<--->"+sig, err)
	}
	return nodeInfo.Sig == sig, nil
}
func (e *EtcdIns) SetConfig(key string, val interface{}) error {
	str, err := yamlutil.Marshal(val)
	if err != nil {
		return errutil.Extend("对象写入配置中心时序列化出错", err)
	}
	path := e.kvsPath() + key
	ctx, cancel := context.WithTimeout(context.Background(), e.rwTimeout)
	_, err2 := e.client.Put(ctx, path, str)
	cancel()
	if err2 != nil {
		return errutil.Extend("对象写入配置中心时出错", err2)
	}
	return nil
}
func (e *EtcdIns) GetConfig(key string, ref interface{}) error {
	path := e.kvsPath() + key
	ctx, cancel := context.WithTimeout(context.Background(), e.rwTimeout)
	resp, err := e.client.Get(ctx, path)
	cancel()
	if err != nil {
		return errutil.Extend("对象从配置中心读取时出错", err)
	}
	for _, kv := range resp.Kvs {
		if string(kv.Key) == path {
			str := string(kv.Value)
			if err := yamlutil.Unmarshal(str, ref); err != nil {
				return errutil.Extend("对象从配置中心读取反序列化时出错", err)
			}
			return nil
		}
	}
	return errutil.New("查找不到对应的Key:" + key)
}
func (e *EtcdIns) DelConfig(key string) error {
	path := e.kvsPath() + key
	ctx, cancel := context.WithTimeout(context.Background(), e.rwTimeout)
	_, err := e.client.Delete(ctx, path)
	cancel()
	if err != nil {
		return errutil.Extend("对象从配置中心删除时出错", err)
	}
	return nil
}
