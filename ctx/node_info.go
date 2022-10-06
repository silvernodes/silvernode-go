package ctx

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/silvernodes/silvernode-go/utils"
	"github.com/silvernodes/silvernode-go/utils/netutil"
	"github.com/silvernodes/silvernode-go/utils/snowflake"
	"github.com/silvernodes/silvernode-go/utils/yamlutil"
)

type NodeInfo struct {
	NodeId    string
	Name      string
	EndPoints []string
	IsPub     bool
	BackEnds  []string
	LogLevel  int
	MainPort  uint64
	Metrics   bool
	Sig       string
	UsrDatas  map[string]interface{}
}

func NewNodeInfo() *NodeInfo {
	n := new(NodeInfo)
	n.EndPoints = make([]string, 0, 0)
	n.BackEnds = make([]string, 0, 0)
	n.UsrDatas = make(map[string]interface{})
	n.Metrics = true
	return n
}

func (n *NodeInfo) Marshal() (string, error) {
	return yamlutil.Marshal(n)
}

func (n *NodeInfo) Unmarshal(str string) error {
	return yamlutil.Unmarshal(str, n)
}

func (n *NodeInfo) Clone() *NodeInfo {
	clone := new(NodeInfo)
	clone.NodeId = n.NodeId
	clone.Name = n.Name
	clone.EndPoints = make([]string, 0, len(n.EndPoints))
	for _, ep := range n.EndPoints {
		clone.EndPoints = append(clone.EndPoints, ep)
	}
	clone.IsPub = n.IsPub
	clone.BackEnds = make([]string, 0, len(n.BackEnds))
	for _, be := range n.BackEnds {
		clone.BackEnds = append(clone.BackEnds, be)
	}
	clone.LogLevel = n.LogLevel
	clone.MainPort = n.MainPort
	clone.Metrics = n.Metrics
	clone.Sig = "..."
	clone.UsrDatas = make(map[string]interface{})
	for k, v := range n.UsrDatas {
		clone.UsrDatas[k] = v
	}
	return clone
}

func GetNodeNameFromId(nodeId string) string {
	infos := strings.Split(nodeId, "#")
	return infos[0]
}

func NodeSignature(n *NodeInfo) {
	if _nodeId != "" {
		return
	}
	if n.NodeId == "" {
		n.NodeId = n.Name + "#" + snowflake.Generate()
	} else if n.NodeId == "k8s.metadata.name" {
		n.NodeId = os.Getenv("K8S_META_NAME")
	}
	if !strings.HasPrefix(n.NodeId, n.Name+"#") {
		n.NodeId = n.Name + "#" + n.NodeId
	}
	for i, ep := range n.EndPoints {
		if strings.Contains(ep, "k8s.status.podIP") { // k8s网络地址适配
			K8S_POD_IP := os.Getenv("K8S_POD_IP")
			ep = strings.ReplaceAll(ep, "k8s.status.podIP", K8S_POD_IP)
		}
		n.EndPoints[i] = suitableEP(ep)
	}
	n.Sig = utils.MD5("?nodeid=" + n.NodeId + "&name=" + n.Name + "&sf=" + snowflake.Generate()) // call when register node info to DC, then other node use the sig to connect to this node <----> check url
	_nodeId = n.NodeId
}

func suitableEP(ep string) string {
	tmpInfos := strings.Split(ep, "://")
	if len(tmpInfos) < 2 {
		return ep
	}
	proto := tmpInfos[0]
	tmps := strings.Split(tmpInfos[1], "/")
	subfix := ""
	if len(tmps) >= 2 {
		subfix = tmps[1]
	}
	urlInfos := strings.Split(tmps[0], ":")
	url := urlInfos[0]
	var port uint64 = 0
	if len(urlInfos) >= 2 {
		if val, err := strconv.Atoi(urlInfos[1]); err == nil {
			port = uint64(val)
		}
	}
	// url模糊匹配
	if strings.Contains(url, ".x") {
		if localIPs, err := netutil.GetLocalIPv4s(); err == nil {
			surs := strings.Split(url, ".")
			for _, localIP := range localIPs {
				bRet := true
				dsts := strings.Split(localIP, ".")
				for i, sur := range surs {
					if sur != "x" && sur != dsts[i] {
						bRet = false
						break
					}
				}
				if bRet {
					url = localIP
					break
				}
			}
		}
	}
	// 端口随机
	if port == 0 {
		if val, err := netutil.GetAvailablePort(); err == nil {
			port = val
		}
	}
	ep = proto + "://" + url + ":" + fmt.Sprint(port)
	if subfix != "" {
		ep += "/" + subfix
	}
	return ep
}

var _nodeId string

func GetNodeId() string {
	return _nodeId
}

func GuestPrefix() string {
	return "!guest"
}

func IsGuest(nodeId string) bool {
	return strings.HasPrefix(nodeId, GuestPrefix())
}
