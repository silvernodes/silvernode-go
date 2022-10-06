package peers

import (
	"github.com/silvernodes/silvernode-go/ctx"
	"github.com/silvernodes/silvernode-go/utils/gobutil"
	"github.com/silvernodes/silvernode-go/utils/jsonutil"
)

var _outerCodec Codec = nil
var _innerCodec Codec = nil

type Codec interface {
	Encode(interface{}) ([]byte, error)
	Decode([]byte, interface{}) error
}

func SetOuterCodec(codec Codec) {
	_outerCodec = codec
}

func SetInnerCodec(codec Codec) {
	_innerCodec = codec
}

func getCodec(node string) (Codec, bool) {
	if ctx.IsGuest(node) {
		return _outerCodec, _outerCodec != nil
	}
	return _innerCodec, _innerCodec != nil
}

type JsonCodec struct {
}

func NewJsonCodec() *JsonCodec {
	return new(JsonCodec)
}

func (j *JsonCodec) Encode(obj interface{}) ([]byte, error) {
	return jsonutil.MarshalRaw(obj)
}

func (j *JsonCodec) Decode(data []byte, ref interface{}) error {
	return jsonutil.UnmarshalRaw(data, ref)
}

type GobCodec struct {
}

func NewGobCodec() *GobCodec {
	return new(GobCodec)
}

func (g *GobCodec) Encode(obj interface{}) ([]byte, error) {
	return gobutil.Marshal(obj)
}

func (g *GobCodec) Decode(data []byte, ref interface{}) error {
	return gobutil.Unmarshal(data, ref)
}
