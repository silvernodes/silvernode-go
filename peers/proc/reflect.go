package proc

import (
	"go/token"
	"reflect"
)

var typeOfError = reflect.TypeOf((*error)(nil)).Elem()

type MethodType struct {
	Method    reflect.Method
	ArgType   reflect.Type
	ReplyType reflect.Type
}

func WalkSuitableMethods(proc interface{}, on func(reflect.Method, reflect.Type, reflect.Type)) {
	typ := reflect.TypeOf(proc)
	for m := 0; m < typ.NumMethod(); m++ {
		method := typ.Method(m)
		mtype := method.Type
		if method.PkgPath != "" {
			continue // 方法必须为Public.
		}
		numIn := mtype.NumIn()
		if numIn != 3 {
			continue // Method needs three ins: receiver, *args, *reply.
		}
		argType := mtype.In(1)
		if !isExportedOrBuiltinType(argType) {
			continue // First arg need not be a pointer.
		}
		var replyType reflect.Type
		replyType = mtype.In(2)
		if replyType.Kind() != reflect.Ptr {
			continue
		}
		if !isExportedOrBuiltinType(replyType) {
			continue // Reply type must be exported.
		}
		if mtype.NumOut() != 1 {
			continue // Method needs one out.
		}
		returnType := mtype.Out(0)
		if returnType != typeOfError && numIn == 3 {
			continue // The return type of the method must be error.
		}
		on(method, argType, replyType)
	}
}

func SuitableMethods(typ reflect.Type) map[string]*MethodType {
	methods := make(map[string]*MethodType)
	for m := 0; m < typ.NumMethod(); m++ {
		method := typ.Method(m)
		mtype := method.Type
		mname := method.Name
		if method.PkgPath != "" {
			continue // 方法必须为Public.
		}
		if mname == "GetMeta" {
			continue
		}
		numIn := mtype.NumIn()
		if numIn != 3 && numIn != 2 {
			continue // Method needs three ins: receiver, *args, *reply.
		}
		argType := mtype.In(1)
		if !isExportedOrBuiltinType(argType) {
			continue // First arg need not be a pointer.
		}
		var replyType reflect.Type
		if numIn > 2 {
			replyType = mtype.In(2)
			if replyType.Kind() != reflect.Ptr {
				continue // Second arg must be a pointer.
			}
			if !isExportedOrBuiltinType(replyType) {
				continue // Reply type must be exported.
			}
		}
		if mtype.NumOut() != 1 {
			continue // Method needs one out.
		}
		returnType := mtype.Out(0)
		if returnType != typeOfError && numIn == 3 {
			continue // The return type of the method must be error.
		}
		if numIn > 2 {
			methods[mname] = &MethodType{Method: method, ArgType: argType, ReplyType: replyType}
		} else {
			methods[mname] = &MethodType{Method: method, ArgType: argType}
		}
	}
	return methods
}

func isExportedOrBuiltinType(t reflect.Type) bool {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return token.IsExported(t.Name()) || t.PkgPath() == ""
}
