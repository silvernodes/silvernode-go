package peers

import (
	"fmt"
	"reflect"
	"sync"

	_proc "github.com/silvernodes/silvernode-go/peers/proc"
	"github.com/silvernodes/silvernode-go/process"
	"github.com/silvernodes/silvernode-go/utils/errutil"
)

var _peers map[string]*peer
var _lock sync.RWMutex

func Register(proc interface{}, processor process.Processor) (Peer, error) {
	return RegisterWithNick("", false, proc, processor)
}

func RegisterInner(proc interface{}, processor process.Processor) (Peer, error) {
	return RegisterWithNick("", true, proc, processor)
}

func RegisterWithNick(nick string, inner bool, proc interface{}, processor process.Processor) (Peer, error) {
	_lock.Lock()
	defer _lock.Unlock()
	procTp := reflect.TypeOf(proc)
	if procTp.Kind() != reflect.Ptr {
		return nil, errutil.New("proc必须为指针类型!")
	}
	if nick == "" {
		nick = procTp.Elem().Name()
	}
	if _, b := _peers[nick]; b {
		return nil, errutil.New("已存在同昵称Peer:" + nick)
	}
	p := copeer(nick, inner, proc, procTp, processor)
	if err := p.filedsAutoLoad(); err != nil {
		return nil, err
	}
	_proc.RecordMetaRaw(p.typ, p.methods)
	_peers[nick] = p
	txt := "Peer[" + nick + "]注册完毕."
	if p.meta != nil {
		txt += "(!)"
	}
	fmt.Println(txt)
	return p, nil
}

func Dispose(nick string, withProcessor bool) {
	_lock.Lock()
	defer _lock.Unlock()

	if peer, b := _peers[nick]; b {
		peer.disposing = true
		if withProcessor {
			peer.processor.Terminate()
		}
	}
}

func GetPeer(nick string) (Peer, bool) {
	return getpeer(nick)
}

func getpeer(nick string) (*peer, bool) {
	_lock.Lock()
	defer _lock.Unlock()

	peer, b := _peers[nick]
	if b && !peer.disposing {
		return peer, true
	}
	return nil, false
}

func onExchange(nodeId string, data []byte) error {
	e := &exchange{}
	if err := e.Unmarshal(data); err != nil {
		return err
	}
	localExchange(nodeId, e)
	return nil
}

func localExchange(nodeId string, e *exchange) {
	if peer, b := getpeer(e.To); b {
		peer.onExchange(nodeId, e)
	}
}
