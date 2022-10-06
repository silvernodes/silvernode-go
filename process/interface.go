package process

type Coroutine interface {
	Sync() interface{}
}

type Scheduler interface {
	Cancel()
}

type Process interface {
	Pid() int64
	Terminate()
}

type Processor interface {
	Pid() int64
	Execute(task func()) bool
	Terminate()
	Running() bool
	TaskLen() int
	CoroutineNum() int
}

type Service interface {
	Pid() int64
	Start(loop func(), recycle func())
	StartTick(loop func(), interval int, recycle func())
	Sync() bool
	Running() bool
	Terminate()
}
