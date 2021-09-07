package server

import (
	"github.com/jin1ming/Gedis/pkg/config"
	"github.com/tidwall/redcon"
	"log"
	"strconv"
	"strings"
	"sync"
)

type CmdBuffer struct {
	args	[][]byte
	conn	*redcon.Conn
}

type Server struct {
	scheme		string
	host		string
	protocol	string
	addr		string
	buffer		chan CmdBuffer
	cmd 		map[string]Cmd
}

type Cmd struct {
	f		func(conn redcon.Conn, args [][]byte)
	argvMin	int
	argvMax	int
}

func New() *Server {
	s := &Server{
		scheme:   "gedis",
		host:     "127.0.0.1",
		protocol: "resp3",
		addr:     config.GetConfig().Base.Bind + ":" + strconv.Itoa(config.GetConfig().Base.Port),
		buffer:   make(chan CmdBuffer, 0),
		cmd:	  make(map[string]Cmd),
	}
	return s
}

//func (s *Server) Serve() {
//	dbMap := db.GetDB().DataMap
//	for {
//
//	}
//}

func (s *Server) registerCmd(names []string, argcMin int, argcMax int,
	f func(conn redcon.Conn, args [][]byte)){

	c := Cmd{
		f:    f,
		argvMin: argcMin,
		argvMax: argcMax,
	}

	for _, name := range names{
		name = strings.ToLower(name)
		s.cmd[name] = c
	}
}

func (s *Server) registerCmds() {
	var mu sync.RWMutex
	var items = make(map[string][]byte)
	var ps redcon.PubSub
	s.registerCmd([]string{"ping"}, 0, 0, func(conn redcon.Conn, args [][]byte) {
		conn.WriteString("PONG")
	})
	s.registerCmd([]string{"quit"}, 0, 0, func(conn redcon.Conn, args [][]byte) {
		conn.WriteString("OK")
	})
	s.registerCmd([]string{"set"}, 3, 3, func(conn redcon.Conn, args [][]byte) {
		mu.Lock()
		items[string(args[1])] = args[2]
		mu.Unlock()
		conn.WriteString("OK")
	})
	s.registerCmd([]string{"get"}, 2, 2, func(conn redcon.Conn, args [][]byte) {
		mu.RLock()
		val, ok := items[string(args[1])]
		mu.RUnlock()
		if !ok {
			conn.WriteNull()
		} else {
			conn.WriteBulk(val)
		}
	})
	s.registerCmd([]string{"del"}, 2, 2, func(conn redcon.Conn, args [][]byte) {
		mu.Lock()
		_, ok := items[string(args[1])]
		delete(items, string(args[1]))
		mu.Unlock()
		if !ok {
			conn.WriteInt(0)
		} else {
			conn.WriteInt(1)
		}
	})
	s.registerCmd([]string{"publish"}, 3, 3, func(conn redcon.Conn, args [][]byte) {
		conn.WriteInt(ps.Publish(string(args[1]), string(args[2])))
	})
	s.registerCmd([]string{"subscribe", "psubscribe"}, 2, 0, func(conn redcon.Conn, args [][]byte) {
		command := strings.ToLower(string(args[0]))
		for i := 1; i < len(args); i++ {
			if command == "psubscribe" {
				ps.Psubscribe(conn, string(args[i]))
			} else {
				ps.Subscribe(conn, string(args[i]))
			}
		}
	})
}

func (s *Server) ExecCommand(conn redcon.Conn, cmd redcon.Command) {
	c, ok := s.cmd[strings.ToLower(string(cmd.Args[0]))]
	if !ok {
		conn.WriteError("ERR unknown command '" + string(cmd.Args[0]) + "'")
	}
	if c.argvMin > len(cmd.Args) || (c.argvMax != 0 && c.argvMax < len(cmd.Args)) {
		conn.WriteError("ERR wrong number of arguments for '" + string(cmd.Args[0]) + "' command")
		return
	}
	c.f(conn, cmd.Args)
}

func (s *Server) Start()  {
	log.Printf("started server at %s", s.addr)
	s.registerCmds()

	err := redcon.ListenAndServe(s.addr,
		s.ExecCommand,
		func(conn redcon.Conn) bool {
			// Use this function to accept or deny the connection.
			// log.Printf("accept: %s", conn.RemoteAddr())
			return true
		},
		func(conn redcon.Conn, err error) {
			// This is called when the connection has been closed
			// log.Printf("closed: %s, err: %v", conn.RemoteAddr(), err)
		},
	)
	if err != nil {
		log.Fatal(err)
	}
}
