package server

import (
	"github.com/jin1ming/Gedis/pkg/db"
	"github.com/tidwall/redcon"
	"strings"
)

func (s *Server) registerCmd(names []string, argcMin int, argcMax int,
	f0 func(args [][]byte), f func(conn redcon.Conn, result interface{})) {

	c := Cmd{
		f0:      f0,
		f:       f,
		argvMin: argcMin,
		argvMax: argcMax,
	}

	for _, name := range names {
		name = strings.ToLower(name)
		s.cmd[name] = c
	}
}

func (s *Server) ExecCommand(conn redcon.Conn, cmd redcon.Command) {
	c, ok := s.cmd[strings.ToLower(string(cmd.Args[0]))]
	if !ok {
		conn.WriteError("ERR unknown command '" + string(cmd.Args[0]) + "'")
		return
	}
	if c.argvMin > len(cmd.Args) || (c.argvMax != 0 && c.argvMax < len(cmd.Args)) {
		conn.WriteError("ERR wrong number of arguments for '" + string(cmd.Args[0]) + "' command")
		return
	}
	if s.aofBuffer != nil {
		s.aofBuffer <- cmd
	}
	c.f0(cmd.Args)
	var result interface{}
	if _, ok = s.returnCmd[string(cmd.Args[0])]; ok {
		result = <-s.returnBuffer
	}
	c.f(conn, result)
}

func (s *Server) register() {
	DB := db.GetDB()
	var ps redcon.PubSub

	s.registerCmd([]string{"ping"}, 0, 0, func(args [][]byte) {
	}, func(conn redcon.Conn, result interface{}) {
		conn.WriteString("PONG")
	})
	s.registerCmd([]string{"quit", "exit"}, 0, 0, func(args [][]byte) {
	}, func(conn redcon.Conn, result interface{}) {
		conn.WriteString("OK")
	})
	s.registerCmd([]string{"set"}, 3, 3, func(args [][]byte) {
		DB.Set(args...)
	}, func(conn redcon.Conn, result interface{}) {
		conn.WriteString("OK")
	})
	s.registerCmd([]string{"get"}, 2, 2, func(args [][]byte) {
		conn
	}, func(conn redcon.Conn, result interface{}) {
		val, ok := DB.Get(args...)
		if !ok {
			conn.WriteNull()
		} else {
			conn.WriteBulk(val)
		}
	})
	s.registerCmd([]string{"expire"}, 3, 3, func(conn redcon.Conn, args [][]byte) {
		num := DB.Expire(args...)
		conn.WriteInt(num)
	})
	s.registerCmd([]string{"setex"}, 4, 4, func(conn redcon.Conn, args [][]byte) {
		DB.SetEx(args...)
		conn.WriteString("OK")
	})
	s.registerCmd([]string{"setnx"}, 3, 3, func(conn redcon.Conn, args [][]byte) {
		num := DB.SetNx(args...)
		conn.WriteInt(num)
	})
	s.registerCmd([]string{"del"}, 2, 2, func(conn redcon.Conn, args [][]byte) {
		num := DB.Del(args...)
		conn.WriteInt(num)
	})
	s.registerCmd([]string{"rpush"}, 3, 0, func(conn redcon.Conn, args [][]byte) {
		l := DB.RPush(args...)
		conn.WriteInt(l)
	})
	s.registerCmd([]string{"llen"}, 2, 2, func(conn redcon.Conn, args [][]byte) {
		l := DB.LLen(args...)
		conn.WriteInt(l)
	})
	s.registerCmd([]string{"rpop", "lpop"}, 2, 2, func(conn redcon.Conn, args [][]byte) {
		res := DB.RLPop(args...)
		if res == nil {
			conn.WriteNull()
		} else {
			conn.WriteBulk(res)
		}
	})
	s.registerCmd([]string{"sadd"}, 3, 0, func(conn redcon.Conn, args [][]byte) {
		num := DB.SAdd(args...)
		conn.WriteInt(num)
	})
	s.registerCmd([]string{"smembers"}, 2, 2, func(conn redcon.Conn, args [][]byte) {
		res := DB.SMembers(args...)
		if res == nil {
			conn.WriteNull()
		} else {
			conn.WriteAny(res)
		}
	})
	s.registerCmd([]string{"sismember"}, 3, 3, func(conn redcon.Conn, args [][]byte) {
		num := DB.SisMember(args...)
		conn.WriteInt(num)
	})
	// TODO: 订阅模式待实现
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
