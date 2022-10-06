package gobutil

import (
	"bytes"
	"encoding/gob"
)

func Marshal(obj interface{}) ([]byte, error) {
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(obj); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func Unmarshal(data []byte, ref interface{}) error {
	buf := new(bytes.Buffer)
	if _, err := buf.Write(data); err != nil {
		return err
	}
	dec := gob.NewDecoder(buf)
	return dec.Decode(ref)
}
