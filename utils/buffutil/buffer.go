package buffutil

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"reflect"

	"github.com/silvernodes/silvernode-go/utils/errutil"
)

type Buffer struct {
	buffer      []byte
	bytesBuffer *bytes.Buffer
	err         error
}

func NewBuffer(capacity int) *Buffer {
	buffer := new(Buffer)
	buffer.buffer = make([]byte, capacity)
	buffer.bytesBuffer = bytes.NewBuffer(buffer.buffer)
	buffer.bytesBuffer.Reset()
	return buffer
}

func (b *Buffer) WriteByte(value byte) *Buffer {
	if b.err != nil {
		return b
	}
	err := binary.Write(b.bytesBuffer, binary.LittleEndian, value)
	if err != nil {
		b.err = errutil.Extend(fmt.Sprintf("WriteByte(%v)", value), err)
	}
	return b
}

func (b *Buffer) WriteBool(value bool) *Buffer {
	if b.err != nil {
		return b
	}
	err := binary.Write(b.bytesBuffer, binary.LittleEndian, value)
	if err != nil {
		b.err = errutil.Extend(fmt.Sprintf("WriteBool(%v)", value), err)
	}
	return b
}

func (b *Buffer) WriteBytes(value []byte) *Buffer {
	if b.err != nil {
		return b
	}
	err := binary.Write(b.bytesBuffer, binary.LittleEndian, value)
	if err != nil {
		b.err = errutil.Extend(fmt.Sprintf("WriteBytes(%v)", value), err)
	}
	return b
}

func (b *Buffer) WriteShort(value int16) *Buffer {
	if b.err != nil {
		return b
	}
	err := binary.Write(b.bytesBuffer, binary.LittleEndian, value)
	if err != nil {
		b.err = errutil.Extend(fmt.Sprintf("WriteShort(%v)", value), err)
	}
	return b
}

func (b *Buffer) WriteInt(value int32) *Buffer {
	if b.err != nil {
		return b
	}
	err := binary.Write(b.bytesBuffer, binary.LittleEndian, value)
	if err != nil {
		b.err = errutil.Extend(fmt.Sprintf("WriteInt(%v)", value), err)
	}
	return b
}

func (b *Buffer) WriteLong(value int64) *Buffer {
	if b.err != nil {
		return b
	}
	err := binary.Write(b.bytesBuffer, binary.LittleEndian, value)
	if err != nil {
		b.err = errutil.Extend(fmt.Sprintf("WriteLong(%v)", value), err)
	}
	return b
}

func (b *Buffer) WriteString(value string) *Buffer {
	if b.err != nil {
		return b
	}
	if value == "" {
		b.WriteInt(0)
		return b
	}
	buffer := ([]byte)(value)
	b.WriteInt(int32(len(buffer))) // write the len of the string
	if b.err != nil {
		return b
	}
	err := binary.Write(b.bytesBuffer, binary.LittleEndian, buffer)
	if err != nil {
		b.err = errutil.Extend(fmt.Sprintf("WriteString(%v)", value), err)
	}
	return b
}

func (b *Buffer) WriteFloat(value float32) *Buffer {
	if b.err != nil {
		return b
	}
	err := binary.Write(b.bytesBuffer, binary.LittleEndian, value)
	if err != nil {
		b.err = errutil.Extend(fmt.Sprintf("WriteFloat(%v)", value), err)
	}
	return b
}

func (b *Buffer) WriteInts(value []int32) *Buffer {
	if b.err != nil {
		return b
	}
	length := len(value)
	b.WriteInt(int32(length)) // write the len of the []int
	for _, v := range value {
		b.WriteInt(v)
	}
	if b.err != nil {
		b.err = errutil.Extend(fmt.Sprintf("WriteInts(%v)", value), b.err)
		return b
	}
	return b
}

func (b *Buffer) WriteArray(value []interface{}) *Buffer {
	if b.err != nil {
		return b
	}
	length := len(value)
	b.WriteInt(int32(length)) // write the len of the []int
	for _, v := range value {
		b.WriteObject(v)
	}
	if b.err != nil {
		b.err = errutil.Extend(fmt.Sprintf("WriteArray(%v)", value), b.err)
		return b
	}
	return b
}

func (b *Buffer) WriteMap(value map[string]interface{}) *Buffer {
	if b.err != nil {
		return b
	}
	length := len(value)
	b.WriteInt(int32(length)) // write the len of the hash
	for k, v := range value {
		b.WriteString(k)
		b.WriteObject(v)
	}
	if b.err != nil {
		b.err = errutil.Extend(fmt.Sprintf("WriteMap(%v)", value), b.err)
		return b
	}
	return b
}

func (b *Buffer) WriteHash(value map[interface{}]interface{}) *Buffer {
	if b.err != nil {
		return b
	}
	length := len(value)
	b.WriteInt(int32(length)) // write the len of the hash
	for k, v := range value {
		b.WriteObject(k)
		b.WriteObject(v)
	}
	if b.err != nil {
		b.err = errutil.Extend(fmt.Sprintf("WriteHash(%v)", value), b.err)
		return b
	}
	return b
}

func (b *Buffer) WriteObject(value interface{}) *Buffer {
	if b.err != nil {
		return b
	}
	if value == nil {
		b.WriteByte(Null)
		b.WriteByte(byte(0))
		return b
	}
	switch value.(type) {
	case byte:
		b.WriteByte(Byte)
		b.WriteByte(value.(byte))
	case bool:
		b.WriteByte(Bool)
		b.WriteBool(value.(bool))
	case int16:
		b.WriteByte(Short)
		b.WriteShort(value.(int16))
	case int:
		b.WriteByte(Int)
		b.WriteInt(int32(value.(int)))
	case int32:
		b.WriteByte(Int)
		b.WriteInt(value.(int32))
	case int64:
		b.WriteByte(Long)
		b.WriteLong(value.(int64))
	case string:
		b.WriteByte(byte('s'))
		b.WriteString(value.(string))
	case float32:
		b.WriteByte(Float)
		b.WriteFloat(value.(float32))
	case []int32:
		b.WriteByte(Ints)
		b.WriteInts(value.([]int32))
	case []interface{}:
		b.WriteByte(Array)
		b.WriteArray(value.([]interface{}))
	case map[string]interface{}:
		b.WriteByte(Map)
		b.WriteMap(value.(map[string]interface{}))
	case map[interface{}]interface{}:
		b.WriteByte(Hash)
		b.WriteHash(value.(map[interface{}]interface{}))
	default:
		itype := reflect.TypeOf(value)
		ctype, ok := _customBufferExtends[itype]
		if ok {
			b.WriteByte(ctype.bSign)
			ctype.serializeFunc(b, value)
		} else {
			b.err = errutil.Extend("不支持的类型", b.err)
		}
	}
	if b.err != nil {
		b.err = errutil.Extend(fmt.Sprintf("WriteHash(%v)", value), b.err)
		return b
	}
	return b
}

func (b *Buffer) Flush() ([]byte, error) {
	if b.err != nil {
		return nil, b.err
	}
	return b.bytesBuffer.Bytes(), nil
}

func (b *Buffer) Error() error {
	return b.err
}

func (b *Buffer) Buf() *bytes.Buffer {
	return b.bytesBuffer
}

func (b *Buffer) Dispose() {
	b.buffer = nil
}
