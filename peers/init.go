package peers

import (
	silvernode "github.com/silvernodes/silvernode-go"
	"github.com/silvernodes/silvernode-go/process"
	"github.com/silvernodes/silvernode-go/utils/errutil"
	"github.com/silvernodes/silvernode-go/utils/timeutil"
)

type Peer interface {
	Nick() string
	Proc() interface{}
	Processor() process.Processor
	Invoke(node string, method string, args interface{}, reply interface{}) error
	Call(node string, method string, args interface{}, reply interface{}, done func(error))
	SendEvent(node string, method string, args interface{}) error
}

type SetupParam struct {
	Timeout   int
	OnPreProc func(nodeId string, peerNick string, funcName string) (interface{}, error)
	OnMonitor func(info *ExChangeMessage)
}

var _setup *SetupParam

func init() {
	_peers = make(map[string]*peer)
	_setup = new(SetupParam)
	_setup.Timeout = 15000
	_setup.OnPreProc = func(nodeId string, peerNick string, funcName string) (interface{}, error) {
		return nil, nil
	}
	_outerCodec = NewJsonCodec()

}

func Setup(param *SetupParam) {
	if param.Timeout > 0 {
		_setup.Timeout = param.Timeout
	}
	if param.OnPreProc != nil {
		_setup.OnPreProc = param.OnPreProc
	}
	if param.OnMonitor != nil {
		_setup.OnMonitor = param.OnMonitor
	}
}

func Boot() {
	silvernode.BindPipeline(&silvernode.Pipeline{
		OnMessage: func(nodeId string, msg interface{}) error {
			datas, ok := msg.([]byte)
			if !ok {
				return errutil.New("Peer接受来自OnInBound数据格式必须为[]byte")
			}
			return onExchange(nodeId, datas)
		},
	})
	go func() {
		for {
			timeutil.Wait(7)
			loopCheck()
		}
	}()
}

func loopCheck() {
	_lock.RLock()
	defer _lock.RUnlock()

	for _, peer := range _peers {
		if peer.disposing {
			continue
		}
		peer.checkTimeOut()
	}
}
