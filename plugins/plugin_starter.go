package plugins

type PluginStarter interface {
	ConfBean() interface{}
	OnInstall() (interface{}, error)
}
