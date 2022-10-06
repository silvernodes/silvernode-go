package jsonutil

import (
	// json "github.com/json-iterator/go"
	"encoding/json"
)

func Marshal(obj interface{}) (string, error) {
	b, err := json.Marshal(obj)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func Unmarshal(str string, ref interface{}) error {
	err := json.Unmarshal([]byte(str), ref)
	return err
}

func MarshalRaw(obj interface{}) ([]byte, error) {
	return json.Marshal(obj)
}

func UnmarshalRaw(data []byte, ref interface{}) error {
	return json.Unmarshal(data, ref)
}
