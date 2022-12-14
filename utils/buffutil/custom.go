package buffutil

import (
	"reflect"
)

const (
	Byte   byte = 'b'
	Short       = 't'
	Int         = 'i'
	Long        = 'l'
	String      = 's'
	Float       = 'f'
	Ints        = 'I'
	Array       = 'A'
	Map         = 'M'
	Hash        = 'H'
	Bool        = 'B'
	Null        = 'N'
	Unknow      = 0
)

type CustomType struct {
	itype           reflect.Type
	bSign           byte
	serializeFunc   func(b *Buffer, obj interface{})
	deserializeFunc func(p *Parser) interface{}
}

func NewCustomType(itype reflect.Type, bSign byte, serializeFunc func(b *Buffer, obj interface{}), deserializeFunc func(p *Parser) interface{}) *CustomType {
	c := new(CustomType)
	c.itype = itype
	c.bSign = bSign
	c.serializeFunc = serializeFunc
	c.deserializeFunc = deserializeFunc
	return c
}

var _customBufferExtends map[reflect.Type]*CustomType
var _customParserExtends map[byte]*CustomType

func init() {
	_customBufferExtends = make(map[reflect.Type]*CustomType)
	_customParserExtends = make(map[byte]*CustomType)
}

func ExtendCustomType(itype reflect.Type, bSign byte, serializeFunc func(b *Buffer, obj interface{}), deserializeFunc func(p *Parser) interface{}) {
	c := NewCustomType(itype, bSign, serializeFunc, deserializeFunc)
	_customBufferExtends[itype] = c
	_customParserExtends[bSign] = c
}
