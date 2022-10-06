package process

type service struct {
	pid     int64
	sch     *scheduler
	signal  chan byte
	killing bool
	recycle func()
}

func coservice(pid int64) *service {
	s := new(service)
	s.pid = pid
	s.sch = coscheduler()
	s.killing = false
	return s
}

func (s *service) Pid() int64 {
	return s.pid
}

func (s *service) Start(loop func(), recycle func()) {
	s.signal = make(chan byte, 1)
	s.recycle = recycle
	go s.sch.schedule(loop, 0, 0, 0, nil)
}

func (s *service) StartTick(loop func(), interval int, recycle func()) {
	s.signal = make(chan byte, 1)
	s.recycle = recycle
	go s.sch.schedule(loop, interval, 0, 0, nil)
}

func (s *service) Sync() bool {
	if s.signal == nil {
		return false
	}
	<-s.signal
	return true
}

func (s *service) Terminate() {
	if !s.killing {
		s.killing = true
		s.sch.terminate()
		s.signal <- 0
		close(s.signal)
		if s.recycle != nil {
			s.recycle()
		}
		remove(s.pid)
	}
}

func (s *service) Running() bool {
	return !s.killing
}
