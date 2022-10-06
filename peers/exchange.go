package peers

import (
	"encoding/gob"
	"fmt"
	"reflect"

	"github.com/silvernodes/silvernode-go/ctx"
	"github.com/silvernodes/silvernode-go/peers/proc"
	"github.com/silvernodes/silvernode-go/utils/buffutil"
	"github.com/silvernodes/silvernode-go/utils/errutil"
)

type exchange struct {
	From  string
	To    string
	Func  string
	Seq   int64
	Ret   byte
	Err   string
	Datas []byte

	args   interface{}
	parser *buffutil.Parser
}

func (e *exchange) Marshal(node string, capacity int) ([]byte, error) {
	buffer := buffutil.NewBuffer(capacity).
		WriteString(e.From).
		WriteString(e.To).
		WriteString(e.Func).
		WriteLong(e.Seq).
		WriteByte(e.Ret).
		WriteString(e.Err)
	if buffer.Error() != nil {
		return nil, errutil.Extend("交互数据头序列化出错", buffer.Error())
	}
	if codec, b := getCodec(node); b {
		data, err := codec.Encode(e.args)
		if err != nil {
			return nil, errutil.Extend("交互数据体序列化出错", err)
		}
		buffer.WriteBytes(data)
	} else {
		enc := gob.NewEncoder(buffer.Buf())
		if err := enc.Encode(e.args); err != nil {
			return nil, errutil.Extend("交互数据体序列化出错", err)
		}
	}
	return buffer.Flush()
}

func (e *exchange) Unmarshal(data []byte) error {
	parser := buffutil.NewParser(data, 0)
	e.From = parser.ReadString()
	e.To = parser.ReadString()
	e.Func = parser.ReadString()
	e.Seq = parser.ReadLong()
	e.Ret = parser.ReadByte()
	e.Err = parser.ReadString()
	if parser.Error() != nil {
		return errutil.Extend("交互数据头反序列化出错", parser.Error())
	}
	e.parser = parser
	return nil
}

func (e *exchange) FetchArgv(nodeId string, mtype *proc.MethodType, ctx interface{}) (reflect.Value, error) {
	var argv reflect.Value
	if e.args == nil { // 远端
		argIsValue := false // if true, need to indirect before calling.
		if mtype.ArgType.Kind() == reflect.Ptr {
			argv = reflect.New(mtype.ArgType.Elem())
		} else {
			argv = reflect.New(mtype.ArgType)
			argIsValue = true
		}
		if err := e.FetchArgs(nodeId, argv.Interface()); err != nil {
			return argv, err
		}
		if argIsValue {
			argv = argv.Elem()
		}
	} else { // 本地
		argtype := reflect.TypeOf(e.args)
		srctype := mtype.ArgType
		if srctype.Kind() == reflect.Ptr {
			argtype = argtype.Elem()
			srctype = srctype.Elem()
		}
		if argtype.Name() != srctype.Name() {
			return argv, errutil.New("参数类型不匹配:" + srctype.Name() + "<-->" + argtype.Name())
		}
		argv = reflect.ValueOf(e.args)
		if mtype.ArgType.Kind() != reflect.Ptr {
			argv = argv.Elem()
		}
	}
	argt := argv.Type().Elem()
	numfield := argt.NumField()
	if numfield > 0 {
		for index := 0; index < numfield; index++ {
			field := argt.Field(index)
			auto := field.Tag.Get("auto")
			if auto == "node" {
				if field.Type.Kind() == reflect.String {
					argv.Elem().Field(index).SetString(nodeId)
				}
			} else if auto == "ctx" {
				typeKind := field.Type.Kind()
				if typeKind == reflect.Ptr || typeKind == reflect.Interface {
					if field.Type.Name() == reflect.TypeOf(ctx).Name() {
						argv.Elem().Field(index).Set(reflect.ValueOf(ctx))
					}
				}
			}
		}
	}
	return argv, nil
}

func (e *exchange) FetchReplyv(nodeId string, mtype *proc.MethodType) reflect.Value {
	if nodeId == ctx.GetNodeId() && e.Seq != 0 {
		if p, exists := getpeer(e.From); exists {
			if call, b := p.getCall(e.Seq); b {
				return reflect.ValueOf(call.Reply)
			}
		}
	}
	replyv := reflect.New(mtype.ReplyType.Elem())
	switch mtype.ReplyType.Elem().Kind() {
	case reflect.Map:
		replyv.Elem().Set(reflect.MakeMap(mtype.ReplyType.Elem()))
	case reflect.Slice:
		replyv.Elem().Set(reflect.MakeSlice(mtype.ReplyType.Elem(), 0, 0))
	}
	return replyv

}

func (e *exchange) FetchArgs(node string, token interface{}) error {
	if e.parser == nil {
		return errutil.New("交互数据头反序列化尚未完成")
	}
	if codec, b := getCodec(node); b {
		err := codec.Decode(e.parser.Buf().Bytes(), token)
		if err != nil {
			return errutil.Extend("交互数据体反序列化出错", err)
		}
	} else {
		dec := gob.NewDecoder(e.parser.Buf())
		if err := dec.Decode(token); err != nil {
			return errutil.Extend("交互数据体反序列化出错", err)
		}
	}
	e.args = token
	e.PrintInfo(node, false)
	return nil
}

func (e *exchange) PrintInfo(node string, sender bool) {
	if _setup.OnMonitor == nil {
		return
	}
	s := ctx.GetNodeId()
	r := node
	argStr := fmt.Sprint(e.args)
	if !sender {
		r, s = s, r
	}
	var info *ExChangeMessage = nil
	if e.Ret > 0 {
		info = &ExChangeMessage{
			Sender:   s,
			From:     e.From,
			Receiver: r,
			To:       e.To,
			Func:     fmt.Sprint(e.Seq),
			Args:     argStr,
			Err:      e.Err,
		}
	} else {
		info = &ExChangeMessage{
			Sender:   s,
			From:     e.From,
			Receiver: r,
			To:       e.To,
			Func:     e.Func,
			Args:     argStr,
			Err:      e.Err,
		}
	}
	_setup.OnMonitor(info)
}

type ExChangeMessage struct {
	Sender   string
	From     string
	Receiver string
	To       string
	Func     string
	Args     string
	Err      string
}

type callFunc struct {
	Method  string
	Args    interface{}
	Reply   interface{}
	Error   error
	Done    func(error)
	Chan    chan error
	TimeOut int64
}
