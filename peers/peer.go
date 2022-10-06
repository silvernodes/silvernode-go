package peers

import (
	"reflect"
	"strings"
	"sync"

	silvernode "github.com/silvernodes/silvernode-go"
	"github.com/silvernodes/silvernode-go/ctx"
	_proc "github.com/silvernodes/silvernode-go/peers/proc"

	"github.com/silvernodes/silvernode-go/process"
	"github.com/silvernodes/silvernode-go/utils/errutil"
	"github.com/silvernodes/silvernode-go/utils/snowflake"
	"github.com/silvernodes/silvernode-go/utils/timeutil"
)

type peer struct {
	nick      string
	typ       reflect.Type
	proc      reflect.Value
	methods   map[string]*_proc.MethodType
	callbacks map[int64]*callFunc
	processor process.Processor
	procctx   interface{}
	lock      sync.RWMutex
	inner     bool
	disposing bool
	meta      _proc.ProcMeta
}

func copeer(nick string, inner bool, proc interface{}, procTp reflect.Type, processor process.Processor) *peer {
	p := new(peer)
	p.nick = nick
	p.typ = procTp
	p.proc = reflect.ValueOf(proc)
	p.methods = _proc.SuitableMethods(p.typ)
	p.callbacks = make(map[int64]*callFunc)
	p.processor = processor
	p.procctx = proc
	p.inner = inner
	p.disposing = false
	p.meta = nil
	if meta, ok := _proc.CheckHasMeta(p.typ, p.proc); ok {
		p.meta = meta
	}
	return p
}

func (p *peer) filedsAutoLoad() error {
	for i := 0; i < p.typ.Elem().NumField(); i++ {
		field := p.typ.Elem().Field(i)
		if field.Type.Kind() == reflect.Ptr || field.Type.Kind() == reflect.Interface {
			if field.Name == "Peer" {
				p.proc.Elem().Field(i).Set(reflect.ValueOf(p))
			}
		}
	}
	return nil
}

func (p *peer) Nick() string {
	return p.nick
}

func (p *peer) Proc() interface{} {
	return p.procctx
}

func (p *peer) Processor() process.Processor {
	return p.processor
}

func (p *peer) onExchange(nodeId string, e *exchange) {
	if p.processor == nil || !p.processor.Running() {
		defer errutil.Catch(func(err error) {
			silvernode.Error(err)
		})
		p.dealExchange(nodeId, e)
	} else {
		p.processor.Execute(func() {
			p.dealExchange(nodeId, e)
		})
	}
}

func (p *peer) dealExchange(nodeId string, e *exchange) {
	if e.Ret == 0 {
		if p.inner && ctx.IsGuest(nodeId) {
			p.response(nodeId, e, nil, errutil.New("没有访问权限:"+p.nick))
			return
		}

		ctx, err := _setup.OnPreProc(nodeId, e.To, e.Func)
		if err != nil {
			p.response(nodeId, e, nil, err)
			return
		}

		// 利用反射调取proc方法
		mtype, b := p.methods[e.Func]
		if b {
			if p.meta != nil {
				args, reply, err := p.meta.CreateBeans(e.Func, nodeId, ctx)
				if err != nil {
					p.response(nodeId, e, nil, err)
					return
				}
				if e.args != nil { // 本地调用
					args = e.args
					if e.Seq != 0 {
						if p, exists := getpeer(e.From); exists {
							if call, b := p.getCall(e.Seq); b {
								reply = call.Reply
							}
						}
					}
				} else if err := e.FetchArgs(nodeId, args); err != nil {
					p.response(nodeId, e, nil, errutil.Extend("请求数据反序列化出错", err))
					return
				}
				if err := p.meta.ProcessFlow(e.Func, p.Proc(), args, reply); err != nil {
					p.response(nodeId, e, nil, err)
					return
				}
				p.response(nodeId, e, reply, nil)

			} else {
				argv, err := e.FetchArgv(nodeId, mtype, ctx)
				if err != nil {
					p.response(nodeId, e, nil, err)
					return
				}
				if e.Seq == 0 {
					function := mtype.Method.Func
					returnValues := function.Call([]reflect.Value{p.proc, argv})
					errInter := returnValues[0].Interface()
					var reterr error = nil
					if errInter != nil {
						reterr = errInter.(error)
						silvernode.Error(errutil.Extend("目标事件执行异常:"+e.Func, reterr))
					}
				} else {
					replyv := e.FetchReplyv(nodeId, mtype)
					function := mtype.Method.Func
					returnValues := function.Call([]reflect.Value{p.proc, argv, replyv})
					errInter := returnValues[0].Interface()
					var reterr error = nil
					if errInter != nil {
						reterr = errInter.(error)
					}
					p.response(nodeId, e, replyv.Interface(), reterr)
				}
			}
		} else {
			if e.Seq == 0 {
				silvernode.Error(errutil.New("目标事件不存在:" + e.Func))
			} else {
				p.response(nodeId, e, nil, errutil.New("方法不存在:"+e.Func))
			}
		}
	} else {
		call, b := p.takeoutCall(e.Seq)
		if b {
			if e.Err != "" {
				call.Error = errutil.New(e.Err)
			} else {
				if nodeId != ctx.GetNodeId() {
					if err := e.FetchArgs(nodeId, call.Reply); err != nil {
						call.Error = errutil.Extend("应答结果反序列化出错", err)
					}
				}
			}
			if call.Done != nil {
				call.Done(call.Error)
			} else {
				call.Chan <- call.Error
			}
		}
	}
}

func (p *peer) buildCall(method string, args interface{}, reply interface{}, done func(error), c chan error) (int64, *callFunc) {
	p.lock.Lock()
	defer p.lock.Unlock()

	seq := snowflake.GenerateRaw()
	call := &callFunc{
		Method:  method,
		Args:    args,
		Reply:   reply,
		Error:   nil,
		Done:    done,
		Chan:    c,
		TimeOut: timeutil.Time() + int64(_setup.Timeout),
	}
	p.callbacks[seq] = call
	return seq, call
}

func (p *peer) getCall(seq int64) (*callFunc, bool) {
	p.lock.Lock()
	defer p.lock.Unlock()
	call, b := p.callbacks[seq]
	return call, b
}

func (p *peer) takeoutCall(seq int64) (*callFunc, bool) {
	p.lock.Lock()
	defer p.lock.Unlock()
	call, b := p.callbacks[seq]
	if b {
		delete(p.callbacks, seq)
	}
	return call, b
}

func (p *peer) checkTimeOut() {
	p.lock.Lock()
	defer p.lock.Unlock()

	now := timeutil.Time()
	dirtyList := make([]int64, 0, 0)
	for seq, call := range p.callbacks {
		if call.TimeOut > now {
			err := errutil.New("请求超时!")
			if call.Done != nil {
				call.Done(err)
			} else {
				call.Chan <- err
			}
			dirtyList = append(dirtyList, seq)
		}
	}
	for _, seq := range dirtyList {
		delete(p.callbacks, seq)
	}
	dirtyList = nil
}

func (p *peer) request(node string, method string, args interface{}, reply interface{}, done func(error), c chan error) (*callFunc, error) {
	var call *callFunc = nil
	seq := int64(0)
	if reply != nil {
		seq, call = p.buildCall(method, args, reply, done, c)
	}
	methodInfo := strings.Split(method, ".")
	if len(methodInfo) != 2 {
		return nil, errutil.New("方法名必须符合PeerNick.FuncName的规范:" + method)
	}
	e := &exchange{
		From: p.nick,
		To:   methodInfo[0],
		Func: methodInfo[1],
		args: args,
		Seq:  seq,
		Ret:  0,
		Err:  "",
	}
	e.PrintInfo(node, true)
	if node == ctx.GetNodeId() {
		localExchange(node, e)
	} else {
		data, err := e.Marshal(node, 1024)
		if err != nil {
			return nil, errutil.Extend("交互数据序列化出错:"+ctx.GetNodeId()+" -> "+node, err)
		}
		// 跨节点发送
		if err := silvernode.Send(node, data); err != nil {
			return nil, errutil.Extend("跨节点交互出错:"+ctx.GetNodeId()+" -> "+node, err)
		}
	}
	return call, nil
}

func (p *peer) Do(method string, args interface{}, reply interface{}) error {
	node := ctx.GetNodeId()
	return p.Invoke(node, method, args, reply)
}

func (p *peer) Invoke(node string, method string, args interface{}, reply interface{}) error {
	c := make(chan error, 5)
	call, err := p.request(node, method, args, reply, nil, c)
	if err != nil {
		return err
	}
	err2 := <-call.Chan
	close(call.Chan)
	call = nil
	return err2
}

func (p *peer) Call(node string, method string, args interface{}, reply interface{}, done func(error)) {
	if _, err := p.request(node, method, args, reply, done, nil); err != nil {
		done(err)
	}
}

func (p *peer) SendEvent(node string, method string, args interface{}) error {
	_, err := p.request(node, method, args, nil, nil, nil)
	return err
}

func (p *peer) response(node string, e *exchange, reply interface{}, err error) error {
	if e.Seq != 0 {
		Err := ""
		if err != nil {
			Err = err.Error()
		}
		r := &exchange{
			From: e.To,
			To:   e.From,
			Func: "",
			args: reply,
			Seq:  e.Seq,
			Ret:  1,
			Err:  Err,
		}
		r.PrintInfo(node, true)
		if node == ctx.GetNodeId() {
			localExchange(node, r)
		} else {
			msg, err := r.Marshal(node, 4096)
			if err != nil {
				return errutil.Extend(ctx.GetNodeId()+" -> "+node, err)
			}
			// 跨节点发送
			if err := silvernode.Send(node, msg); err != nil {
				return errutil.Extend("跨节点交互出错:"+ctx.GetNodeId()+" -> "+node, err)
			}
		}
	}
	return nil
}
