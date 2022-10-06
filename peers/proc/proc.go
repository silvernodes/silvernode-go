package proc

import "reflect"

func init() {
	_metas = make(map[reflect.Type]*_MetaInfo)
}
