package silvernode

import (
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"sync"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/silvernodes/silvernode-go/board"
	"github.com/silvernodes/silvernode-go/cluster"
	"github.com/silvernodes/silvernode-go/ctx"
	"github.com/silvernodes/silvernode-go/log"

	"github.com/silvernodes/silvernode-go/nets"
	"github.com/silvernodes/silvernode-go/plugins"
	"github.com/silvernodes/silvernode-go/utils/errutil"
	"github.com/silvernodes/silvernode-go/utils/fileutil"
	"github.com/silvernodes/silvernode-go/utils/flagutil"
	"github.com/silvernodes/silvernode-go/utils/netutil"
	"github.com/silvernodes/silvernode-go/utils/snowflake"
)

type SetupParam struct {
	AppConf     string
	ClusterNode string
	SpecifiedId string
}

func FlagParam() *SetupParam {
	s := new(SetupParam)
	flagutil.AddFlag("appconf", "", "指定配置文件的路径")
	flagutil.AddFlag("clusternode", "", "指定集群中心节点配置")
	flagutil.AddFlag("specifiedid", "", "指定特定的节点id")
	s.AppConf, _ = flagutil.Get("appconf")
	s.ClusterNode, _ = flagutil.Get("clusternode")
	s.SpecifiedId, _ = flagutil.Get("specifiedid")
	return s
}

type Pipeline struct {
	Init       func()
	Start      func()
	OnConnect  func(nodeId string)
	OnInBound  func(nodeId string, msg []byte) (interface{}, error)
	OnMessage  func(nodeId string, msg interface{}) error
	OnOutBound func(nodeId string, msg interface{}) ([]byte, error)
	OnClose    func(nodeId string, err error)
	OnError    func(err error)
}

type _SilverNode struct {
	info       *ctx.NodeInfo
	reg        cluster.IRegistry
	log        *log.Logger
	netWorkers map[string]nets.INetWorker
	sync.RWMutex
}

var _node *_SilverNode
var _setup *SetupParam
var _pipe *Pipeline
var _inited bool

func init() {
	_node = new(_SilverNode)
	_node.info = ctx.NewNodeInfo()
	_node.netWorkers = make(map[string]nets.INetWorker)

	_setup = new(SetupParam)
	_setup.AppConf = fileutil.CurrentDir() + "app.yml"

	_pipe = new(Pipeline)
	_pipe.OnConnect = func(nodeId string) {

	}
	_pipe.OnInBound = func(nodeId string, data []byte) (interface{}, error) {
		return data, nil
	}
	_pipe.OnMessage = func(nodeId string, msg interface{}) error {
		return Send(nodeId, msg) // echo server
	}
	_pipe.OnOutBound = func(nodeId string, msg interface{}) ([]byte, error) {
		return msg.([]byte), nil
	}
	_pipe.OnClose = func(nodeId string, err error) {

	}
	_pipe.OnError = func(err error) {
		_node.log.Log(log.ERROR, err)
	}
	errutil.CustomErrFunc(_pipe.OnError)
}

func Setup(param *SetupParam) {
	if param.AppConf != "" {
		_setup.AppConf = param.AppConf
	}
	if param.ClusterNode != "" {
		_setup.ClusterNode = param.ClusterNode
	}
	if param.SpecifiedId != "" {
		_setup.SpecifiedId = param.SpecifiedId
	}
}

func BindPipeline(pipe *Pipeline) {
	if pipe != nil {
		if pipe.Init != nil {
			_pipe.Init = pipe.Init
		}
		if pipe.Start != nil {
			_pipe.Start = pipe.Start
		}
		if pipe.OnConnect != nil {
			_pipe.OnConnect = pipe.OnConnect
		}
		if pipe.OnInBound != nil {
			_pipe.OnInBound = pipe.OnInBound
		}
		if pipe.OnMessage != nil {
			_pipe.OnMessage = pipe.OnMessage
		}
		if pipe.OnOutBound != nil {
			_pipe.OnOutBound = pipe.OnOutBound
		}
		if pipe.OnClose != nil {
			_pipe.OnClose = pipe.OnClose
		}
		if pipe.OnError != nil {
			_pipe.OnError = pipe.OnError
			errutil.CustomErrFunc(_pipe.OnError)
		}
	}
}

func Serve() error {
	runtime.GOMAXPROCS(runtime.NumCPU())

	text, err := fileutil.LoadFile(_setup.AppConf)
	if err != nil {
		return errutil.Extend("读取配置文件发生错误:"+_setup.AppConf, err)
	}
	if err := ctx.CoreConf().LoadAppYaml(text); err != nil {
		return errutil.Extend("解析配置文件发生错误", err)
	}

	if ctx.CoreConf().CheckConfExists("cluster") {
		clusterType := ""
		if err := ctx.CoreConf().GetConfDatas("cluster.type", &clusterType); err != nil {
			return errutil.Extend("集群注册中心类型缺失", err)
		}
		reg, err := cluster.CreateRegistry(clusterType)
		if err != nil {
			return err
		}
		_node.reg = reg
	}

	if err := plugins.InstallPlugins(_node.reg); err != nil {
		return errutil.Extend("初始化插件系统发生错误", err)
	}

	if ctx.CoreConf().CheckConfExists("node") { // 本地配置
		if err := ctx.CoreConf().GetConfDatas("node", _node.info); err != nil {
			return errutil.Extend("加载节点配置信息出错", err)
		}
	} else if _node.reg != nil && _setup.ClusterNode != "" { // 云端配置
		if err := _node.reg.GetConfig(_setup.ClusterNode, _node.info); err != nil {
			return errutil.Extend("从注册中心获取节点配置信息出错", err)
		}
	} else {
		return errutil.New("无法获取正确的节点配置信息！")
	}

	if len(_node.info.EndPoints) <= 0 {
		return errutil.New("每个节点至少应包含一个主EndPoint！")
	}
	if _node.info.MainPort == 0 {
		port, err := netutil.GetAvailablePort()
		if err != nil {
			return errutil.Extend("本地无法获得可用的随机端口", err)
		}
		_node.info.MainPort = port
	}
	ctx.NodeSignature(_node.info)
	if _setup.SpecifiedId != "" {
		_node.info.NodeId = _setup.SpecifiedId
	}
	_, ip, _, err := netutil.ParseUrlInfo(_node.info.EndPoints[0])
	if err != nil {
		return errutil.Extend("解析节点注册信息发生错误", err)
	}
	mainUrl := fmt.Sprintf("%s:%d", ip, _node.info.MainPort)
	mainMux := http.NewServeMux()
	mainMux.HandleFunc("/", board.DashBoard)
	if _node.info.Metrics {
		mainMux.Handle("/metrics", promhttp.Handler())
	}
	go func() {
		if err := http.ListenAndServe(mainUrl, mainMux); err != nil {
			_pipe.OnError(errutil.Extend("开启检测监听时发生错误", err))
		}
	}()
	_node.log = log.NewLogger(_node.info.NodeId, _node.info.LogLevel, nil)

	nets.BindEventListener(&nets.NetEventListener{
		OnConnect: func(nodeId string) {
			defer errutil.Catch(_pipe.OnError)
			_node.log.Log(log.INFO, "新的链接已建立:"+nodeId)
			_pipe.OnConnect(nodeId)
		},
		OnMessage: func(nodeId string, msg []byte) {
			defer errutil.Catch(_pipe.OnError)
			if !_inited {
				_pipe.OnError(errutil.New("节点尚未初始化完毕!"))
				return
			}
			data, err := _pipe.OnInBound(nodeId, msg)
			if err != nil {
				_pipe.OnError(err)
			} else {
				if err := _pipe.OnMessage(nodeId, data); err != nil {
					_pipe.OnError(err)
				}
			}
		},
		OnClose: func(nodeId string, err error) {
			defer errutil.Catch(_pipe.OnError)
			_node.log.Log(log.ERROR, "链接已关闭:"+nodeId+"|"+err.Error())
			_pipe.OnClose(nodeId, err)
		},
		OnError:     _pipe.OnError,
		OnCheckNode: onCheckNode,
	})

	if _pipe.Init != nil {
		_pipe.Init()
	}
	for _, ep := range _node.info.EndPoints {
		if err := Listen(ep); err != nil {
			return err
		}
	}
	if _node.reg != nil {
		if err := cluster.Serve(_node.reg, &cluster.ClusterParam{
			SelfInfo:   _node.info,
			OnScanning: onScanning,
		}); err != nil {
			return err
		}
	}
	_inited = true
	printInfo()
	if _pipe.Start != nil {
		_pipe.Start()
	}

	select {}
	return nil
}

func Listen(url string) error {
	if _, _, _, err := netutil.ParseUrlInfo(url); err != nil {
		return err
	}
	go func() {
		_node.Lock()
		netWorker, err := getNetWorker(url)
		_node.Unlock()
		if err != nil {
			_pipe.OnError(err)
		}
		if err := netWorker.Listen(url); err != nil {
			_pipe.OnError(err)
		}
	}()
	return nil
}

func Connect(nodeId string, url string) (string, error) {
	info, exists := nets.ConnectManagerIns().GetConnectInfo(nodeId)
	if exists {
		if info.Url() == url {
			return nodeId, nil
		} else {
			proto, _, port, err := netutil.ParseUrlInfo(url)
			if err != nil {
				return "", err
			}
			nodeId = nodeId + "@" + proto + fmt.Sprint(port)
		}
	}
	netWorker, err := getNetWorker(url)
	if err != nil {
		return "", err
	}
	originInfo := nets.CombineOriginInfo(_node.info.NodeId, _node.info.EndPoints[0], _node.info.Sig)
	return nodeId, netWorker.Connect(nodeId, url, originInfo)
}

func getNetWorker(url string) (nets.INetWorker, error) {
	_, exists := _node.netWorkers[url]
	if !exists {
		proto := strings.Split(url, "://")[0]
		netWorker, err := nets.CreateNetWorker(proto)
		if err != nil {
			return nil, err
		}
		_node.netWorkers[url] = netWorker
	}
	return _node.netWorkers[url], nil
}

func Send(nodeId string, msg interface{}) error {
	data, err := _pipe.OnOutBound(nodeId, msg)
	if err != nil {
		if errutil.IsEOF(err) {
			return nil
		}
		return errutil.Extend("数据发送失败", err)
	}
	connInfo, exists := nets.ConnectManagerIns().GetConnectInfo(nodeId)
	if !exists {
		return errutil.New("尚未建立到对应节点的链接:" + nodeId)
	}
	return connInfo.Send(data)
}

func Close(nodeId string) error {
	connInfo, exists := nets.ConnectManagerIns().GetConnectInfo(nodeId)
	if !exists {
		return errutil.New("尚未建立到对应节点的链接:" + nodeId)
	}
	return connInfo.Close(errutil.EOF())
}

func GetNodeList(name string) []string {
	return nets.ConnectManagerIns().GetNodes(name)
}

func SetUsrData(k string, v interface{}) {
	_node.info.UsrDatas[k] = v
}

func GetUsrDatas(nodeId string) (map[string]interface{}, bool) {
	info, err := _node.reg.GetNodeById(nodeId)
	if err != nil {
		_pipe.OnError(err)
		return nil, false
	}
	return info.UsrDatas, true
}

func GetAppConf(prefix string, ref interface{}) error {
	return ctx.CoreConf().GetConfDatas(prefix, ref)
}

func onScanning(otherInfos []*ctx.NodeInfo, err error) {
	_node.Lock()
	defer _node.Unlock()
	if err == nil && otherInfos != nil {
		for _, otherInfo := range otherInfos {
			_, exists := nets.ConnectManagerIns().GetConnectInfo(otherInfo.NodeId)
			if !exists {
				if isBackEnd(otherInfo.NodeId) {
					_node.log.Log(log.INFO, "发现新节点:"+otherInfo.NodeId)
					if _, err := Connect(otherInfo.NodeId, otherInfo.EndPoints[0]); err != nil {
						_pipe.OnError(err)
					}
				}
			}
		}
	}
}

func isBackEnd(id string) bool {
	name := ctx.GetNodeNameFromId(id)
	for _, back := range _node.info.BackEnds {
		if back == name && _node.info.NodeId != id {
			return true
		}
	}
	return false
}

func onCheckNode(origin string) (string, error) {
	ret := false
	id, _, sig, err := nets.ParseOriginInfo(origin)
	if err == nil {
		ret2, err2 := _node.reg.CheckNodeSig(id, sig)
		if err2 != nil {
			return "", err2
		}
		ret = ret2
	} else {
		if !_node.info.IsPub {
			return "", err
		}
	}
	if !ret { // 如果是注册中心无法验证的节点，则视为外来节点
		if !_node.info.IsPub {
			return "", errutil.New("节点证书数据不匹配:" + id + "<--->" + sig) // 如果节点对外不开放，则直接放弃链接
		} else {
			guestId := ""
			if id == "" {
				guestId = ctx.GuestPrefix() + "#" + snowflake.Generate()
			} else if !strings.HasPrefix(id, ctx.GuestPrefix()) {
				guestId = ctx.GuestPrefix() + "#" + id
			}
			return guestId, nil
		}
	} else {
		// 过滤重复发起链接申请的内部节点
		_, exist := nets.ConnectManagerIns().GetConnectInfo(id)
		if exist {
			return "", errutil.New("本地已存在相同的链接:" + id)
		}
		return id, nil
	}
}

func Logger() *log.Logger {
	return _node.log
}

func Error(err error) {
	_pipe.OnError(err)
}

func NodeId() string {
	return _node.info.NodeId
}

func NodeInfo() *ctx.NodeInfo {
	return _node.info.Clone()
}

func printInfo() {
	fmt.Println()
	fmt.Println(BANNER)
	fmt.Println()
	info, _ := _node.info.Clone().Marshal()
	fmt.Println(info)
	fmt.Println()
}

const (
	VERSION string = "0.1.0"
	TITLE   string = "勇者"
	BANNER         = ` ------------- github.com/silvernodes/silvernode-go -------------
    _____ _ __                _   __          __           ______    
   / ___/(_) /   _____  _____/ | / /___  ____/ /__        / ____/___ 
   \__ \/ / / | / / _ \/ ___/  |/ / __ \/ __  / _ \______/ / __/ __ \
  ___/ / / /| |/ /  __/ /  / /|  / /_/ / /_/ /  __/_____/ /_/ / /_/ /
 /____/_/_/ |___/\___/_/  /_/ |_/\____/\__,_/\___/      \____/\____/ 
															 
 --- :: 致那黑夜中的呜咽与怒吼！ :: - v` + VERSION + "-" + TITLE + " ---"
)
