package proc

import (
	"fmt"
	"reflect"
	"strconv"
)

// 自定义Peer调用形式
// 用以规避默认反射形式带来的效率耗损
type ProcMeta interface {
	CreateBeans(method string, from string, ctx interface{}) (interface{}, interface{}, error)
	ProcessFlow(method string, proc interface{}, args interface{}, reply interface{}) error
}

type _MetaInfo struct {
	typ     reflect.Type
	methods map[string]*MethodType
}

var _metas map[reflect.Type]*_MetaInfo = nil

func CheckHasMeta(typ reflect.Type, proc reflect.Value) (ProcMeta, bool) {
	for m := 0; m < typ.NumMethod(); m++ {
		method := typ.Method(m)
		mname := method.Name
		if mname == "GetMeta" {
			returnValues := method.Func.Call([]reflect.Value{proc})
			meta, ok := returnValues[0].Interface().(ProcMeta)
			if !ok {
				return nil, false
			}
			return meta, true
		}
	}
	return nil, false
}

func RecordMeta(proc interface{}) {
	typ := reflect.TypeOf(proc)
	methods := SuitableMethods(typ)
	RecordMetaRaw(typ, methods)
}

func RecordMetaRaw(typ reflect.Type, methods map[string]*MethodType) {
	_metas[typ] = &_MetaInfo{
		typ:     typ,
		methods: methods,
	}
}

func PublishMetas() error {
	for _, p := range _metas {
		if err := buildPeerMeta(p); err != nil {
			return err
		}
	}
	fmt.Println("metas发布完成![" + strconv.Itoa(len(_metas)) + "]")
	return nil
}

func CleanMetas() error {
	for _, p := range _metas {
		removePeerMeta(p)
	}
	fmt.Println("metas清理完毕!")
	return nil
}
