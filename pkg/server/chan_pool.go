package server

type ChanPool struct {
	idle chan chan interface{}
	chs  []chan interface{}
}

func NewChanPool() *ChanPool {
	idle := make(chan chan interface{}, 1024*1024)
	chs := make([]chan interface{}, 1024*1024)
	for i := range chs {
		chs[i] = make(chan interface{})
		idle <- chs[i]
	}
	return &ChanPool{
		idle: idle,
		chs:  chs,
	}
}

func (cp *ChanPool) Get() chan interface{} {
	return <-cp.idle
}

func (cp *ChanPool) Put(ch chan interface{}) {
	cp.idle <- ch
}
