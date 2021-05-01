package server

import (
	"net"
	"sync"
)

type Client struct {
	conn	*net.Conn
	mu		*sync.RWMutex

}