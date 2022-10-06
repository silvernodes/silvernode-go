package plugins

import (
	"reflect"
	"strings"
	"sync"

	"github.com/silvernodes/silvernode-go/cluster"
	"github.com/silvernodes/silvernode-go/ctx"
	"github.com/silvernodes/silvernode-go/utils/errutil"
)

var _starters map[string]PluginStarter
var _plugins map[reflect.Type]map[string]interface{}
var _lock sync.RWMutex

func init() {
	_starters = make(map[string]PluginStarter)
	_plugins = make(map[reflect.Type]map[string]interface{})
}

func Register(starter PluginStarter) error {
	return RegisterIns("", starter)
}

func RegisterIns(insName string, starter PluginStarter) error {
	_lock.Lock()
	defer _lock.Unlock()

	prefix, err := getPluginName(starter, insName)
	if err != nil {
		return err
	}
	_, exist := _starters[prefix+"#"+insName]
	if exist {
		return errutil.New("本地已存在同名插件:" + prefix)
	}
	_starters[prefix+"#"+insName] = starter
	return nil
}

func InstallPlugins(reg cluster.IRegistry) error {
	for prefix, starter := range _starters {
		if err := installPlugin(prefix, starter, reg); err != nil {
			return err
		}
	}
	return nil
}

func installPlugin(prefix string, starter PluginStarter, reg cluster.IRegistry) error {
	infos := strings.Split(prefix, "#")
	prefix = infos[0]
	insName := infos[1]
	if err := getPluginConf(prefix, starter.ConfBean(), reg); err != nil {
		return err
	}
	plugin, err := starter.OnInstall()
	if err != nil {
		return errutil.Extend("安装插件出错:"+prefix, err)
	}
	tp := reflect.TypeOf(plugin)
	getWiredPlugins(tp)[insName] = plugin
	return nil
}

func GetPlugin[T any]() (T, error) {
	return GetPluginIns[T]("")
}

func GetPluginIns[T any](insName string) (T, error) {
	var empty T
	tp := reflect.TypeOf(empty)
	tmp, exist := _plugins[tp]
	if !exist {
		return empty, errutil.New("插件未成功安装或尚未初始化完毕:" + tp.Name())
	}
	plugin, exist2 := tmp[insName]
	if !exist2 {
		return empty, errutil.New("插件实例未成功安装或尚未初始化完毕:" + tp.Name() + "#" + insName)
	}
	return plugin.(T), nil
}

func WirePlugins(obj interface{}) error {
	_lock.RLock()
	defer _lock.RUnlock()

	typ := reflect.TypeOf(obj)
	if typ.Kind() != reflect.Ptr {
		return errutil.New("未能满足插件装载要求:obj必须为指向结构体的指针!")
	}
	argt := typ.Elem()
	if argt.Kind() != reflect.Struct {
		return errutil.New("未能满足插件装载要求:obj必须为指向结构体的指针!")
	}
	argv := reflect.ValueOf(obj).Elem()
	numfield := argt.NumField()
	for i := 0; i < numfield; i++ {
		field := argt.Field(i)
		tag := field.Tag.Get("plugin")
		if tag != "" {
			insName := ""
			if tag != "auto" {
				insName = tag
			}
			tp := field.Type
			tmp, exist := _plugins[tp]
			if !exist {
				return errutil.New("插件未成功安装或尚未初始化完毕:" + tp.Name())
			}
			plugin, exist2 := tmp[insName]
			if !exist2 {
				return errutil.New("插件实例未成功安装或尚未初始化完毕:" + tp.Name() + "#" + insName)
			}
			argv.Field(i).Set(reflect.ValueOf(plugin))
		}
	}
	return nil
}

func getPluginConf(prefix string, ref interface{}, reg cluster.IRegistry) error {
	if ctx.CoreConf().CheckConfExists(prefix) {
		if err := ctx.CoreConf().GetConfDatas(prefix, ref); err != nil {
			return errutil.Extend("插件配置解析错误:"+prefix, err)
		}
	} else if reg != nil {
		if err := reg.GetConfig(prefix, ref); err != nil {
			return errutil.Extend("从注册中心获取插件配置信息出错:"+prefix, err)
		}
	} else {
		return errutil.New("无法获取对应的插件配置:" + prefix)
	}
	return nil
}

func getPluginName(starter PluginStarter, name string) (string, error) {
	pkg := reflect.TypeOf(starter).Elem().PkgPath()
	tmp := strings.ReplaceAll(pkg, "/", ".")
	infos := strings.Split(tmp, ".")
	tmpName := infos[len(infos)-1]
	if !strings.HasPrefix(tmpName, "silvernode-plugin-") || !strings.HasSuffix(tmpName, "-silvernode-plugin") {
		return "", errutil.New("插件名称须遵守形如:silvernode-plugin-xxx或者xxx-silvernode-plugin的命名规范")
	}
	pkgName := strings.TrimSuffix(tmp, "."+tmpName)
	pluginName := ""
	if strings.HasPrefix(tmpName, "silvernode-plugin-") {
		pluginName = strings.TrimPrefix(tmpName, "silvernode-plugin-")
	} else {
		pluginName = strings.TrimSuffix(tmpName, "-silvernode-plugin")
	}
	if name != "" {
		pluginName = name
	}
	return "plugins." + pkgName + "." + pluginName, nil
}

func getWiredPlugins(tp reflect.Type) map[string]interface{} {
	if old, exist := _plugins[tp]; exist {
		return old
	}
	_new := make(map[string]interface{})
	_plugins[tp] = _new
	return _new
}
