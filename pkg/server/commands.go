package server

import (
	"github.com/jin1ming/Gedis/pkg/db"
	"github.com/tidwall/redcon"
	"strings"
)

func (s *Server) registerCmd(names []string, argcMin int, argcMax int,
	f func(conn redcon.Conn, cp db.CmdPackage)) {

	c := Cmd{
		f:       f,
		argvMin: argcMin,
		argvMax: argcMax,
	}

	for _, name := range names {
		//name = strings.ToLower(name)
		s.cmd[name] = c
	}
}

func (s *Server) ExecCommand(conn redcon.Conn, cmd redcon.Command) {
	method := strings.ToLower(string(cmd.Args[0]))
	c, ok := s.cmd[method]
	if !ok {
		conn.WriteError("ERR unknown command '" + string(cmd.Args[0]) + "'")
		return
	}
	if c.argvMin > len(cmd.Args) || (c.argvMax != 0 && c.argvMax < len(cmd.Args)) {
		conn.WriteError("ERR wrong number of arguments for '" + string(cmd.Args[0]) + "' command")
		return
	}
	if s.aofBuffer != nil {
		if _, ok = s.aofCmd[method]; ok {
			d := db.GetDB()
			index := d.Hash(cmd.Args[1])
			s.aofBuffer[index] <- cmd
			//s.aof.WriteCmd(index, cmd.Args)
		}
	}
	var ch chan interface{}
	if _, ok = s.returnCmd[method]; ok {
		ch = s.chanPool.Get().(chan interface{})
	}
	cmd.Args[0] = []byte(method)
	cp := db.CmdPackage{Args: cmd.Args, Ch: ch}
	c.f(conn, cp)
}

func (s *Server) register() {
	var ps redcon.PubSub

	s.registerCmd([]string{"ping"}, 0, 0, func(conn redcon.Conn, _ db.CmdPackage) {
		conn.WriteString("PONG")
	})
	s.registerCmd([]string{"quit", "exit"}, 0, 0, func(conn redcon.Conn, _ db.CmdPackage) {
		conn.WriteString("OK")
	})
	s.registerCmd([]string{"set"}, 3, 3, callBackOk)
	s.registerCmd([]string{"get"}, 2, 2, callBackBytes)
	s.registerCmd([]string{"expire"}, 3, 3, callBackInt)
	s.registerCmd([]string{"setex"}, 4, 4, callBackOk)
	s.registerCmd([]string{"setnx"}, 3, 3, callBackInt)
	s.registerCmd([]string{"del"}, 2, 2, callBackInt)
	s.registerCmd([]string{"rpush"}, 3, 0, callBackInt)
	s.registerCmd([]string{"llen"}, 2, 2, callBackInt)
	s.registerCmd([]string{"rpop", "lpop"}, 2, 2, callBackBytes)
	s.registerCmd([]string{"sadd"}, 3, 0, callBackInt)
	s.registerCmd([]string{"smembers"}, 2, 2, callBackBytes)
	s.registerCmd([]string{"sismember"}, 3, 3, callBackInt)
	// TODO: 订阅模式待实现
	s.registerCmd([]string{"publish"}, 3, 3, func(conn redcon.Conn, cp db.CmdPackage) {
		conn.WriteInt(ps.Publish(string(cp.Args[1]), string(cp.Args[2])))
	})
	s.registerCmd([]string{"subscribe", "psubscribe"}, 2, 0, func(conn redcon.Conn, cp db.CmdPackage) {
		command := strings.ToLower(string(cp.Args[0]))
		for i := 1; i < len(cp.Args); i++ {
			if command == "psubscribe" {
				ps.Psubscribe(conn, string(cp.Args[i]))
			} else {
				ps.Subscribe(conn, string(cp.Args[i]))
			}
		}
	})
}

func callBackInt(conn redcon.Conn, cp db.CmdPackage) {
	d := db.GetDB()
	index := d.Hash(cp.Args[1])
	d.ExecQueue[index] <- cp
	val := <-cp.Ch
	conn.WriteInt(val.(int))
}

func callBackBytes(conn redcon.Conn, cp db.CmdPackage) {
	d := db.GetDB()
	index := d.Hash(cp.Args[1])
	d.ExecQueue[index] <- cp
	val := <-cp.Ch
	if val == nil {
		conn.WriteNull()
	} else {
		conn.WriteBulk(val.([]byte))
	}
}

func callBackOk(conn redcon.Conn, cp db.CmdPackage) {
	d := db.GetDB()
	index := d.Hash(cp.Args[1])
	d.ExecQueue[index] <- cp
	conn.WriteString("OK")
}
