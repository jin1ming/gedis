package server

import (
	"github.com/RussellLuo/timingwheel"
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
	scheme      string
	host        string
	protocol    string
	addr        string
	buffer      chan CmdBuffer
	cmd         map[string]Cmd
	timingWheel *timingwheel.TimingWheel
}

type Cmd struct {
	f       func(conn redcon.Conn, args [][]byte)
	argvMin int
	argvMax int
}

func New() *Server {
	s := &Server{
		scheme:   "gedis",
		host:     "127.0.0.1",
		protocol: "resp3",
		addr:     config.GetConfig().Base.Bind + ":" + strconv.Itoa(config.GetConfig().Base.Port),
		buffer:   make(chan CmdBuffer, 0),
		cmd:      make(map[string]Cmd),
	}
	return s
}

func (s *Server) Start() {
	log.Printf("started server at %s", s.addr)
	s.timingWheel = timingwheel.NewTimingWheel(1*time.Second, 60)
	s.timingWheel.Start()
	defer s.timingWheel.Stop()
	s.registerCmds()
	err := redcon.ListenAndServe(s.addr,
		s.ExecCommand,
		func(conn redcon.Conn) bool {
			// Use this function to accept or deny the connection.
			log.Printf("accept: %s", conn.RemoteAddr())
			return true
		},
		func(conn redcon.Conn, err error) {
			// This is called when the connection has been closed
			log.Printf("closed: %s, err: %v", conn.RemoteAddr(), err)
		},
	)
	if err != nil {
		log.Fatal(err)
	}
}

// RunAfter 延时任务
func (s *Server) RunAfter(d time.Duration, f func()) *timingwheel.Timer {
	return s.timingWheel.AfterFunc(d, f)
}

//// RunEvery 定时任务
//func (s *Server) RunEvery(d time.Duration, f func()) *timingwheel.Timer {
//	return s.timingWheel.ScheduleFunc(&everyScheduler{Interval: d}, f)
//}
