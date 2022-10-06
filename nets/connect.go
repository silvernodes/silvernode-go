package nets

import (
	"net"
	"sync"

	"github.com/silvernodes/silvernode-go/ctx"
	"github.com/silvernodes/silvernode-go/process"
	"github.com/silvernodes/silvernode-go/utils/errutil"
	"github.com/silvernodes/silvernode-go/utils/timeutil"
)

type ConnectInfo struct {
	nodeId    string
	url       string
	proto     string
	conn      net.Conn
	worker    process.Service
	netWorker INetWorker
	ping      bool
	ts        int64
}

func NewConnectInfo(nodeId string, url string, proto string, conn net.Conn, worker process.Service, netWorker INetWorker) *ConnectInfo {
	info := new(ConnectInfo)
	info.nodeId = nodeId
	info.url = url
	info.proto = proto
	info.conn = conn
	info.worker = worker
	info.netWorker = netWorker
	return info
}

func (i *ConnectInfo) NodeId() string {
	return i.nodeId
}

func (i *ConnectInfo) Url() string {
	return i.url
}

func (i *ConnectInfo) Proto() string {
	return i.proto
}

func (i *ConnectInfo) Conn() net.Conn {
	return i.conn
}

func (i *ConnectInfo) NetWorker() INetWorker {
	return i.netWorker
}

func (i *ConnectInfo) Ping() {
	go i.Send([]byte("#ping"))
	i.ping = true
}

func (i *ConnectInfo) Pong() {
	go i.Send([]byte("#pong"))
	i.ping = true
}

func (i *ConnectInfo) CheckPingPong(msg []byte) bool {
	i.ping = false
	i.ts = timeutil.MilliSecond()
	if msg[0] == 35 && len(msg) == 5 {
		strmsg := string(msg)
		if strmsg == "#pong" {
			return true
		} else if strmsg == "#ping" {
			i.Pong()
			return true
		}
	}
	return false
}

func (i *ConnectInfo) Send(msg []byte) error {
	return i.netWorker.Send(i.conn, msg)
}

func (i *ConnectInfo) Close(err error) error {
	i.worker.Terminate()
	return i.netWorker.Close(i.nodeId, i.conn, err)
}

type ConnectManager struct {
	kv map[string]map[string]*ConnectInfo
	vk map[net.Conn]string
	sync.RWMutex
}

func NewConnectManager() *ConnectManager {
	c := new(ConnectManager)
	c.kv = make(map[string]map[string]*ConnectInfo)
	c.vk = make(map[net.Conn]string)
	return c
}

func (c *ConnectManager) AddConnectInfo(nodeId string, url string, proto string, conn net.Conn, worker process.Service, netWorker INetWorker) (*ConnectInfo, error) {
	c.Lock()
	defer c.Unlock()
	name := ctx.GetNodeNameFromId(nodeId)
	_, ok := c.KV(name)[nodeId]
	_, ok2 := c.vk[conn]
	if ok || ok2 {
		return nil, errutil.New("已建立相同键值的链接:" + nodeId)
	}
	info := NewConnectInfo(nodeId, url, proto, conn, worker, netWorker)
	c.KV(name)[nodeId] = info
	c.vk[conn] = nodeId

	return info, nil
}

func (c *ConnectManager) RemoveConnectInfo(nodeId string, conn net.Conn) {
	c.Lock()
	defer c.Unlock()

	name := ctx.GetNodeNameFromId(nodeId)
	_, ok := c.KV(name)[nodeId]
	_, ok2 := c.vk[conn]
	if ok {
		delete(c.KV(name), nodeId)
	}
	if ok2 {
		delete(c.vk, conn)
	}
}

func (c *ConnectManager) GetConnectInfo(nodeId string) (*ConnectInfo, bool) {
	c.RLock()
	defer c.RUnlock()
	name := ctx.GetNodeNameFromId(nodeId)
	info, exist := c.KV(name)[nodeId]
	return info, exist
}

func (c *ConnectManager) GetNodeIdByConn(conn net.Conn) (string, bool) {
	c.RLock()
	defer c.RUnlock()

	nodeId, exist := c.vk[conn]
	return nodeId, exist
}

func (c *ConnectManager) PingPong() {
	dirtyInfo := make([]*ConnectInfo, 0, 100)
	for {
		ms := timeutil.MilliSecond()
		process.Sleep(2000)
		for _, infos := range c.kv {
			for _, info := range infos {
				if info.proto == TCP || info.proto == UDP {
					if info.ping {
						dirtyInfo = append(dirtyInfo, info)
					} else if ms-info.ts > 2000 {
						info.Ping()
					}
				}
			}
			process.Sleep(1)
		}
		for _, info := range dirtyInfo {
			info.Close(errutil.New("PingPong超时!"))
		}
		dirtyInfo = dirtyInfo[0:0]
	}
}

func (c *ConnectManager) CheckPingPong(nodeId string, msg []byte) bool {
	info, exist := c.GetConnectInfo(nodeId)
	if exist {
		return info.CheckPingPong(msg)
	}
	return false
}

func (c *ConnectManager) GetNodes(name string) []string {
	c.RLock()
	defer c.RUnlock()

	ids := make([]string, 0, 3)
	if kvs, exists := c.kv[name]; exists {
		for id, _ := range kvs {
			ids = append(ids, id)
		}
	}
	return ids
}

func (c *ConnectManager) KV(name string) map[string]*ConnectInfo {
	if _, exists := c.kv[name]; !exists {
		c.kv[name] = make(map[string]*ConnectInfo)
	}
	kv, _ := c.kv[name]
	return kv
}
