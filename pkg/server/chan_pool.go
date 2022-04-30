package server

import "sync"

func NewChanPool() *sync.Pool {
	return &sync.Pool{New: func() interface{} {
		return make(chan interface{}, 1)
	}}
}
