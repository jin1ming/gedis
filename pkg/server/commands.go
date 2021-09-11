package server

import (
	"container/list"
	"github.com/jin1ming/Gedis/pkg/data_struct"
	"github.com/jin1ming/Gedis/pkg/db"
	"github.com/tidwall/redcon"
	"strconv"
	"strings"
	"sync"
	"time"
)

func (s *Server) registerCmd(names []string, argcMin int, argcMax int,
	f func(conn redcon.Conn, args [][]byte)) {

	c := Cmd{
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
	c.f(conn, cmd.Args)
}

func (s *Server) registerCmds() {
	var mu sync.RWMutex
	db := db.GetDB()
	strMap := db.StrMap
	listMap := db.ListMap
	setMap := db.SetMap
	var ps redcon.PubSub
	s.registerCmd([]string{"ping"}, 0, 0, func(conn redcon.Conn, args [][]byte) {
		conn.WriteString("PONG")
	})
	s.registerCmd([]string{"quit"}, 0, 0, func(conn redcon.Conn, args [][]byte) {
		conn.WriteString("OK")
	})
	s.registerCmd([]string{"set"}, 3, 3, func(conn redcon.Conn, args [][]byte) {
		mu.Lock()
		strMap[string(args[1])] = args[2]
		mu.Unlock()
		conn.WriteString("OK")
	})
	s.registerCmd([]string{"get"}, 2, 2, func(conn redcon.Conn, args [][]byte) {
		mu.RLock()
		val, ok := strMap[string(args[1])]
		mu.RUnlock()
		if !ok {
			conn.WriteNull()
		} else {
			conn.WriteBulk(val)
		}
	})
	s.registerCmd([]string{"expire"}, 3, 3, func(conn redcon.Conn, args [][]byte) {
		_, ok := strMap[string(args[1])]
		if !ok {
			conn.WriteInt(0)
			return
		}
		t, _ := strconv.Atoi(string(args[2]))
		s.RunAfter(time.Duration(t)*time.Second, func() {
			delete(strMap, string(args[1]))
		})
		conn.WriteInt(1)
	})
	s.registerCmd([]string{"setex"}, 4, 4, func(conn redcon.Conn, args [][]byte) {
		strMap[string(args[1])] = args[3]
		t, _ := strconv.Atoi(string(args[2]))
		s.RunAfter(time.Duration(t)*time.Second, func() {
			delete(strMap, string(args[1]))
		})
		conn.WriteString("OK")
	})
	s.registerCmd([]string{"setnx"}, 3, 3, func(conn redcon.Conn, args [][]byte) {
		_, ok := strMap[string(args[1])]
		if ok {
			conn.WriteInt(0)
			return
		}
		strMap[string(args[1])] = args[2]
		conn.WriteInt(1)
	})
	s.registerCmd([]string{"del"}, 2, 2, func(conn redcon.Conn, args [][]byte) {
		mu.Lock()
		_, ok := strMap[string(args[1])]
		delete(strMap, string(args[1]))
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
	s.registerCmd([]string{"rpush"}, 3, 0, func(conn redcon.Conn, args [][]byte) {
		var l *list.List
		if n, ok := listMap[string(args[1])]; ok {
			l = n
		} else {
			l = list.New()
			listMap[string(args[1])] = l
		}
		for i := 2; i < len(args); i++ {
			l.PushBack(args[i])
		}
		conn.WriteInt(l.Len())
	})
	s.registerCmd([]string{"llen"}, 2, 2, func(conn redcon.Conn, args [][]byte) {
		if n, ok := listMap[string(args[1])]; ok {
			conn.WriteInt(n.Len())
		} else {
			conn.WriteInt(0)
		}
	})
	s.registerCmd([]string{"rpop", "lpop"}, 2, 2, func(conn redcon.Conn, args [][]byte) {
		var l *list.List
		var ok bool
		if l, ok = listMap[string(args[1])]; !ok {
			conn.WriteNull()
		}

		var targetNode *list.Element
		if string(args[0]) == "lpop" {
			targetNode = l.Front()
		} else {
			targetNode = l.Back()
		}

		res := targetNode.Value.([]byte)
		l.Remove(targetNode)
		if l.Len() == 0 {
			// TODO: 好像没起作用
			delete(listMap, string(args[1]))
		}
		conn.WriteBulk(res)
	})
	s.registerCmd([]string{"sadd"}, 3, 0, func(conn redcon.Conn, args [][]byte) {
		var set *data_struct.Set
		var ok bool
		if set, ok = setMap[string(args[1])]; !ok {
			set = data_struct.NewSet()
			setMap[string(args[1])] = set
		}
		repeatNum := 0
		for i := 2; i < len(args); i++ {
			if set.Has(string(args[i])) {
				repeatNum++
				continue
			}
			set.Add(string(args[i]))
		}
		conn.WriteInt(len(args) - 2 - repeatNum)
	})
	s.registerCmd([]string{"smembers"}, 2, 2, func(conn redcon.Conn, args [][]byte) {
		var set *data_struct.Set
		var ok bool
		if set, ok = setMap[string(args[1])]; !ok {
			conn.WriteNull()
		}
		conn.WriteAny(set.GetAllBytes())
	})
	s.registerCmd([]string{"sismember"}, 3, 3, func(conn redcon.Conn, args [][]byte) {
		if set, ok := setMap[string(args[1])]; ok {
			if set.Has(string(args[2])) {
				conn.WriteInt(1)
				return
			}
		}
		conn.WriteInt(0)
	})
}
