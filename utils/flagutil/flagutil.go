package flagutil

import (
	"flag"
)

var _args map[string]*string

func init() {
	_args = make(map[string]*string)
}

func AddFlag(key string, val string, des string) {
	_args[key] = flag.String(key, val, des)
}

func Get(key string) (string, bool) {
	if !flag.Parsed() {
		flag.Parse()
	}

	val, exist := _args[key]
	if !exist {
		return "", false
	}
	return *val, true
}
