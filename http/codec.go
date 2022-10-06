package http

import (
	"encoding/json"
	"io/ioutil"
	_http "net/http"

	"github.com/silvernodes/silvernode-go/utils/errutil"
)

type Codec interface {
	Encode(interface{}) ([]byte, error)
	Decode(*_http.Request, interface{}) error
}

type PostJsonCodec struct {
}

func NewPostJsonCodec() *PostJsonCodec {
	return new(PostJsonCodec)
}

func (j *PostJsonCodec) Encode(obj interface{}) ([]byte, error) {
	return json.Marshal(obj)
}

func (j *PostJsonCodec) Decode(r *_http.Request, ref interface{}) error {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return errutil.Extend("错误的数据读取:"+string(data), err)
	}
	return json.Unmarshal(data, ref)
}
