package server

import (
	"context"
	"github.com/jin1ming/Gedis/pkg/config"
	"github.com/tidwall/redcon"
	"log"
	"strconv"
	"time"
)

type CmdBuffer struct {
	args [][]byte
	conn *redcon.Conn
}

type Server struct {
	scheme    string
	host      string
	protocol  string
	addr      string
	buffer    chan CmdBuffer
	cmd       map[string]Cmd
	aofBuffer chan<- redcon.Command
}

type Cmd struct {
	f       func(conn redcon.Conn, args [][]byte)
	argvMin int
	argvMax int
}

func New(ab chan<- redcon.Command) *Server {
	s := &Server{
		scheme:    "gedis",
		host:      "127.0.0.1",
		protocol:  "resp3",
		addr:      config.GetConfig().Base.Bind + ":" + strconv.Itoa(config.GetConfig().Base.Port),
		buffer:    make(chan CmdBuffer, 0),
		cmd:       make(map[string]Cmd),
		aofBuffer: ab,
	}
	return s
}

func (s *Server) Start(ctx context.Context) {
	log.Printf("started server at %s", s.addr)

	s.registerCmds()

	rs := redcon.NewServerNetwork("tcp", s.addr,
		func(conn redcon.Conn, cmd redcon.Command) {
			if s.aofBuffer != nil {
				s.aofBuffer <- cmd
			}
			s.ExecCommand(conn, cmd)
		}, func(conn redcon.Conn) bool {
			log.Printf("accept: %s", conn.RemoteAddr())
			return true
		}, func(conn redcon.Conn, err error) {
			log.Printf("closed: %s, err: %v", conn.RemoteAddr(), err)
		})

	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Println("NetServer is closing...")
				_ = rs.Close()
				return
			default:
				time.Sleep(1 * time.Second)
			}
		}
	}()

	err := rs.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
