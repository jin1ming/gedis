package server

import (
	"github.com/jin1ming/Gedis/pkg/config"
)

type Server struct {
	scheme		string
	host		string
	protocol	string
	addr		string
}

func New() *Server {
	s := &Server{
		scheme:   "gedis",
		host:     "127.0.0.1",
		protocol: "resp3",
		addr:     config.GetConfig().Base.Bind,
	}
	return s
}

func (s *Server) Start()  {
	go func() {

	}()
}