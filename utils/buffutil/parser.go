package buffutil

import (
	"bytes"
	"encoding/binary"
	"errors"
	"strconv"

	"github.com/silvernodes/silvernode-go/utils/errutil"
)

type Parser struct {
	buffer      []byte
	bytesBuffer *bytes.Buffer
	err         error
}

func NewParser(buffer []byte, offset int) *Parser {
	parser := new(Parser)
	parser.buffer = buffer
	parser.bytesBuffer = bytes.NewBuffer(parser.buffer)
	parser.bytesBuffer.Grow(offset)
	return parser
}

func (b *Parser) ReadByte() byte {
	if b.err != nil {
		return 0
	}
	var ret byte
	err := binary.Read(b.bytesBuffer, binary.LittleEndian, &ret)
	if err != nil {
		b.err = errutil.Extend("ReadByte", err)
		return 0
	}
	return ret
}

func (b *Parser) ReadBytes() []byte {
	return b.bytesBuffer.Bytes()
}

func (b *Parser) ReadBool() bool {
	if b.err != nil {
		return false
	}
	var ret bool
	err := binary.Read(b.bytesBuffer, binary.LittleEndian, &ret)
	if err != nil {
		b.err = errutil.Extend("ReadBool", err)
		return false
	}
	return ret
}

func (b *Parser) ReadShort() int16 {
	if b.err != nil {
		return 0
	}
	var ret int16
	err := binary.Read(b.bytesBuffer, binary.LittleEndian, &ret)
	if err != nil {
		b.err = errutil.Extend("ReadShort", err)
		return 0
	}
	return ret
}

func (b *Parser) ReadInt() int32 {
	if b.err != nil {
		return 0
	}
	var ret int32
	err := binary.Read(b.bytesBuffer, binary.LittleEndian, &ret)
	if err != nil {
		b.err = errutil.Extend("ReadInt", err)
		return 0
	}
	return ret
}

func (b *Parser) ReadLong() int64 {
	if b.err != nil {
		return 0
	}
	var ret int64
	err := binary.Read(b.bytesBuffer, binary.LittleEndian, &ret)
	if err != nil {
		b.err = errutil.Extend("ReadLong", err)
		return 0
	}
	return ret
}

func (b *Parser) ReadString() string {
	length := b.ReadInt() // get the string len
	if b.err != nil {
		return ""
	}
	if length == 0 {
		return ""
	}
	if length > 10240 || length < 0 {
		b.err = errutil.Extend("ReadString", errutil.New("字符串长度非法:"+strconv.Itoa(int(length))))
		return ""
	}
	var tempBuffer []byte = make([]byte, length)
	if err := binary.Read(b.bytesBuffer, binary.LittleEndian, &tempBuffer); err != nil {
		b.err = errutil.Extend("ReadString", err)
		return ""
	}
	return string(tempBuffer)
}

func (b *Parser) ReadFloat() float32 {
	if b.err != nil {
		return 0
	}
	var ret float32
	err := binary.Read(b.bytesBuffer, binary.LittleEndian, &ret)
	if err != nil {
		b.err = errutil.Extend("ReadFloat", err)
		return 0
	}
	return ret
}

func (b *Parser) ReadInts() []int32 {
	length := b.ReadInt() // get the []int32 len
	array := make([]int32, 0, length)
	var i int32
	for i = 0; i < length; i++ {
		item := b.ReadInt()
		array = append(array, item)
	}
	if b.err != nil {
		b.err = errutil.Extend("ReadInts", b.err)
		return nil
	}
	return array
}

func (b *Parser) ReadArray() []interface{} {
	length := b.ReadInt() // get the []int32 len
	array := make([]interface{}, 0, length)
	var i int32
	for i = 0; i < length; i++ {
		item := b.ReadObject()
		array = append(array, item)
	}
	if b.err != nil {
		b.err = errutil.Extend("ReadArray", b.err)
		return nil
	}
	return array
}

func (b *Parser) ReadMap() map[string]interface{} {
	length := b.ReadInt() // get the hash len
	Map := make(map[string]interface{})
	var i int32
	for i = 0; i < length; i++ {
		k := b.ReadString()
		v := b.ReadObject()
		Map[k] = v
	}
	if b.err != nil {
		b.err = errutil.Extend("ReadMap", b.err)
		return nil
	}
	return Map
}

func (b *Parser) ReadHash() map[interface{}]interface{} {
	length := b.ReadInt() // get the hash len
	hash := make(map[interface{}]interface{})
	var i int32
	for i = 0; i < length; i++ {
		k := b.ReadObject()
		v := b.ReadObject()
		hash[k] = v
	}
	if b.err != nil {
		b.err = errutil.Extend("ReadHash", b.err)
		return nil
	}
	return hash
}

func (b *Parser) ReadObject() interface{} {
	c := b.ReadByte()
	if b.err != nil {
		return nil
	}
	switch c {
	case Byte:
		return b.ReadByte()
	case Bool:
		return b.ReadBool()
	case Short:
		return b.ReadShort()
	case Int:
		return b.ReadInt()
	case Long:
		return b.ReadLong()
	case String:
		return b.ReadString()
	case Float:
		return b.ReadFloat()
	case Ints:
		return b.ReadInts()
	case Array:
		return b.ReadArray()
	case Map:
		return b.ReadMap()
	case Hash:
		return b.ReadHash()
	case Null:
		none := b.ReadByte()
		if b.err != nil {
			return nil
		} else if none != byte(0) {
			b.err = errors.New("ReadObject:未知的类型")
			return nil
		} else {
			return nil
		}
	default:
		ctype, ok := _customParserExtends[c]
		if ok {
			return ctype.deserializeFunc(b)
		} else {
			b.err = errors.New("ReadObject:未知的类型")
			return nil
		}
	}
	if b.err != nil {
		b.err = errutil.Extend("ReadObject", b.err)
		return nil
	}
	return nil
}

func (b *Parser) Buf() *bytes.Buffer {
	return b.bytesBuffer
}

func (b *Parser) OverFlow() bool {
	return b.bytesBuffer.Len() <= 0 || b.err != nil
}

func (b *Parser) Error() error {
	return b.err
}
