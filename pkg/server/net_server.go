package server

import (
	"context"
	"github.com/jin1ming/Gedis/pkg/config"
	"github.com/jin1ming/Gedis/pkg/db"
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
	scheme       string
	host         string
	protocol     string
	addr         string
	buffer       chan CmdBuffer
	cmd          map[string]Cmd
	aofBuffer    chan<- redcon.Command
	returnCmd    map[string]struct{}
	aofCmd       map[string]struct{}
	returnBuffer <-chan interface{}
	chanPool     *ChanPool
}

type Cmd struct {
	f       func(conn redcon.Conn, cp db.CmdPackage)
	argvMin int
	argvMax int
	ch      chan interface{}
}

func NewServer(ab chan<- redcon.Command) *Server {
	var empty struct{}
	s := &Server{
		scheme:    "gedis",
		host:      "127.0.0.1",
		protocol:  "resp3",
		addr:      config.GetConfig().Base.Bind + ":" + strconv.Itoa(config.GetConfig().Base.Port),
		buffer:    make(chan CmdBuffer, 0),
		cmd:       make(map[string]Cmd),
		aofBuffer: ab,
		returnCmd: map[string]struct{}{
			"get": empty, "setnx": empty, "del": empty, "rpush": empty, "llen": empty,
			"rpop": empty, "lpop": empty, "sadd": empty, "smembers": empty, "sismember": empty,
		},
		aofCmd: map[string]struct{}{
			"set": empty, "expire": empty, "setnx": empty, "del": empty, "rpush": empty,
			"rpop": empty, "lpop": empty, "sadd": empty, "smembers": empty, "sismember": empty,
		},
		chanPool: NewChanPool(),
	}
	return s
}

func (s *Server) Start(ctx context.Context) {
	log.Printf("started server at %s", s.addr)

	s.register()

	rs := redcon.NewServerNetwork("tcp", s.addr,
		func(conn redcon.Conn, cmd redcon.Command) {
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
