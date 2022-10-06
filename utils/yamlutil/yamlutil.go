package yamlutil

import (
	_yaml "gopkg.in/yaml.v2"
)

func Marshal(obj interface{}) (string, error) {
	b, err := _yaml.Marshal(obj)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func Unmarshal(str string, ref interface{}) error {
	err := _yaml.Unmarshal([]byte(str), ref)
	return err
}

func MarshalRaw(obj interface{}) ([]byte, error) {
	return _yaml.Marshal(obj)
}

func UnmarshalRaw(data []byte, ref interface{}) error {
	return _yaml.Unmarshal(data, ref)
}
