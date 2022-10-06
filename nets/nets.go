package nets

import (
	"net"
	"net/url"
	"strings"

	"github.com/silvernodes/silvernode-go/utils/errutil"
)

const LOCAL string = "local://"

const (
	TCP string = "tcp"
	UDP string = "udp"
	WS  string = "ws"
)

type NetEventListener struct {
	OnConnect   func(nodeId string)
	OnMessage   func(nodeId string, msg []byte)
	OnClose     func(nodeId string, err error)
	OnError     func(err error)
	OnCheckNode func(nodeId string) (string, error)
}

type INetWorker interface {
	Listen(url string) error
	Connect(nodeId string, url string, origin string) error
	Send(conn net.Conn, msg []byte) error
	Close(nodeId string, conn net.Conn, err error) error
}

func CreateNetWorker(proto string) (INetWorker, error) {
	switch proto {
	case WS:
		return NewWSNetWorker(), nil
	case UDP:
		return NewKcpNetWorker(), nil
	case TCP:
		return NewTcpNetWorker(), nil
	default:
		return nil, errutil.New("不支持的协议类型:" + proto)
	}
}

var _connectManager *ConnectManager
var _eventListener *NetEventListener

func init() {
	_connectManager = NewConnectManager()
	// go _connectManager.PingPong()
}

func BindEventListener(eventListener *NetEventListener) {
	_eventListener = eventListener
}

func ConnectManagerIns() *ConnectManager {
	return _connectManager
}

func CombineOriginInfo(nodeId string, url string, sig string) string {
	return url + "?node=" + nodeId + "&sig=" + sig
}

func ParseOriginInfo(origin string) (string, string, string, error) {
	if origin == "" {
		return "", "", "", errutil.New("来源信息缺失!")
	}
	URL, err := url.Parse(origin)
	if err != nil {
		return "", "", "", errutil.Extend("非法的来源信息:"+origin, err)
	}
	query := ParseQuery(origin)
	node, ok1 := query["node"]
	sig, ok2 := query["sig"]
	if !ok1 || !ok2 {
		return "", "", "", errutil.New("非法的来源信息:" + origin)
	}
	return node, URL.Host, sig, nil
}

func ParseQuery(url string) map[string]string {
	args := make(map[string]string)
	tmp := strings.Split(url, "?")
	if len(tmp) <= 1 {
		return args
	}
	tmp = strings.Split(tmp[1], "&")
	for _, t := range tmp {
		tt := strings.Split(t, "=")
		if len(tt) == 2 {
			args[tt[0]] = tt[1]
		}
	}
	return args
}

func IsEOF(err error) bool {
	return err.Error() == "EOF"
}
