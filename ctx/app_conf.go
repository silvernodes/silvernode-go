package ctx

import (
	"errors"
	"strings"

	"github.com/silvernodes/silvernode-go/utils/yamlutil"
)

type AppConf struct {
	datas map[interface{}]interface{}
}

var _appConf *AppConf

func init() {
	_appConf = NewAppConf()
}

func CoreConf() *AppConf {
	return _appConf
}

func NewAppConf() *AppConf {
	a := new(AppConf)
	a.datas = make(map[interface{}]interface{})
	return a
}

func (a *AppConf) LoadAppYaml(yaml string) error {
	return yamlutil.Unmarshal(yaml, a.datas)
}

func (a *AppConf) GetConfDatas(prefix string, ref interface{}) error {
	prefixs := strings.Split(prefix, ".")
	data, b := a.getConfDatasByPrefix(prefixs, 0, a.datas)
	if !b {
		return errors.New("应用配置中查找不到对应的前缀信息:" + prefix)
	}
	yamlstr, err := yamlutil.Marshal(data)
	if err != nil {
		return err
	}
	if err := yamlutil.Unmarshal(yamlstr, ref); err != nil {
		return err
	}
	return nil
}

func (a *AppConf) getConfDatasByPrefix(prefixs []string, deep int, datas map[interface{}]interface{}) (interface{}, bool) {
	ln := len(prefixs)
	if deep >= ln {
		return nil, false
	}
	data, b := datas[prefixs[deep]]
	if !b {
		return nil, false
	}
	if deep == ln-1 {
		return data, true
	}
	datamap, ok := data.(map[interface{}]interface{})
	if !ok {
		return nil, false
	}
	return a.getConfDatasByPrefix(prefixs, deep+1, datamap)
}

func (a *AppConf) CheckConfExists(prefix string) bool {
	prefixs := strings.Split(prefix, ".")
	_, b := a.getConfDatasByPrefix(prefixs, 0, a.datas)
	return b
}
