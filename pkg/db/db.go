package db

import (
	"container/list"
	"github.com/RussellLuo/timingwheel"
	"github.com/jin1ming/Gedis/pkg/data_struct"
	"github.com/jin1ming/Gedis/pkg/event"
	"runtime"
	"strconv"
	"time"
)

type DB struct {
	StrMap       map[string][]byte
	ListMap      map[string]*list.List
	SetMap       map[string]*data_struct.Set
	ZSkipListMap map[string]*data_struct.ZSkipList
	tw           *timingwheel.TimingWheel
	returnBuffer chan interface{}
	ExecQueue    chan CmdPackage
	funcMap      map[string]func(args ...[]byte) interface{}
}

type CmdPackage struct {
	Args [][]byte
	Ch   chan interface{}
}

var db *DB

func init() {
	db = new(DB)
	db.StrMap = make(map[string][]byte)
	db.ListMap = make(map[string]*list.List)
	db.SetMap = make(map[string]*data_struct.Set)
	db.ZSkipListMap = make(map[string]*data_struct.ZSkipList)
	db.tw = event.GetGlobalTimingWheel()

	db.returnBuffer = make(chan interface{})
	db.ExecQueue = make(chan CmdPackage, 1024*1024)
	db.funcMap = map[string]func(args ...[]byte) interface{}{
		"set": db.Set, "get": db.Get, "expire": db.Expire, "setex": db.SetEx,
		"setnx": db.SetNx, "del": db.Del, "rpush": db.RPush, "llen": db.LLen,
		"rpop": db.RLPop, "lpop": db.RLPop, "sadd": db.SAdd, "sismember": db.SisMember,
		"smembers": db.SMembers,
	}
}

func GetDB() *DB {
	return db
}

func (d *DB) Work() {
	runtime.LockOSThread()
	for {
		cp := <-db.ExecQueue
		f, ok := db.funcMap[string(cp.Args[0])]
		if !ok {
			continue
		}
		val := f(cp.Args...)
		if cp.Ch != nil {
			cp.Ch <- val
		}
	}
}

func (d *DB) Set(args ...[]byte) interface{} {
	d.StrMap[string(args[1])] = args[2]
	return nil
}

func (d *DB) Get(args ...[]byte) interface{} {
	val, ok := d.StrMap[string(args[1])]
	if !ok {
		return nil
	}
	return val
}

func (d *DB) Expire(args ...[]byte) interface{} {
	_, ok := d.StrMap[string(args[1])]
	if ok {
		t, _ := strconv.Atoi(string(args[2]))
		d.tw.AfterFunc(time.Duration(t)*time.Second, func() {
			delete(d.StrMap, string(args[1]))
		})
		return 1
	}
	return 0
}

func (d *DB) SetEx(args ...[]byte) interface{} {
	d.StrMap[string(args[1])] = args[3]
	t, _ := strconv.Atoi(string(args[2]))
	d.tw.AfterFunc(time.Duration(t)*time.Second, func() {
		delete(d.StrMap, string(args[1]))
	})
	return nil
}

func (d *DB) SetNx(args ...[]byte) interface{} {
	_, ok := d.StrMap[string(args[1])]
	if ok {
		return 0
	}
	d.StrMap[string(args[1])] = args[2]
	return 1
}

func (d *DB) Del(args ...[]byte) interface{} {
	_, ok := d.StrMap[string(args[1])]
	delete(d.StrMap, string(args[1]))

	if !ok {
		return 0
	} else {
		return 1
	}
}

func (d *DB) RPush(args ...[]byte) interface{} {
	var l *list.List
	if n, ok := d.ListMap[string(args[1])]; ok {
		l = n
	} else {
		l = list.New()
		d.ListMap[string(args[1])] = l
	}
	for i := 2; i < len(args); i++ {
		l.PushBack(args[i])
	}
	return l.Len()
}

func (d *DB) LLen(args ...[]byte) interface{} {
	if n, ok := d.ListMap[string(args[1])]; ok {
		return n.Len()
	} else {
		return 0
	}
}

func (d *DB) RLPop(args ...[]byte) interface{} {
	var l *list.List
	var ok bool
	if l, ok = d.ListMap[string(args[1])]; !ok {
		return nil
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
		delete(d.ListMap, string(args[1]))
	}
	return res
}

func (d *DB) SAdd(args ...[]byte) interface{} {
	var set *data_struct.Set
	var ok bool
	if set, ok = d.SetMap[string(args[1])]; !ok {
		set = data_struct.NewSet()
		d.SetMap[string(args[1])] = set
	}
	repeatNum := 0
	for i := 2; i < len(args); i++ {
		if set.Has(string(args[i])) {
			repeatNum++
			continue
		}
		set.Add(string(args[i]))
	}
	return len(args) - 2 - repeatNum
}

func (d *DB) SMembers(args ...[]byte) interface{} {
	var set *data_struct.Set
	var ok bool
	if set, ok = d.SetMap[string(args[1])]; !ok {
		return nil
	}
	return set.GetAllBytes()
}

func (d *DB) SisMember(args ...[]byte) interface{} {
	if set, ok := d.SetMap[string(args[1])]; ok {
		if set.Has(string(args[2])) {
			return 1
		}
	}
	return 0
}
