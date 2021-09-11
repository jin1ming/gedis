package server

import (
	"net"
	"sync"
	"time"
)

type Client struct {
	conn     net.Conn
	mu       sync.Mutex
	waiting  sync.WaitGroup
	bufWrite chan []byte
}

func NewClient(conn net.Conn) *Client {
	return &Client{
		conn: conn,
	}
}

func (c *Client) Close() {
	ch := make(chan interface{})
	go func() {
		defer close(ch)
		c.waiting.Wait()
	}()
	select {
	case <-ch:
		break
	case <-time.After(8 * time.Second):
		break
	}
	_ = c.conn.Close()
}

func (c *Client) Write(b []byte) error {
	if len(b) == 0 {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	_, err := c.conn.Write(b)
	return err
}
