package db

import (
	"container/list"
	"github.com/RussellLuo/timingwheel"
	"github.com/jin1ming/Gedis/pkg/data_struct"
	"github.com/jin1ming/Gedis/pkg/event"
	"hash"
	"hash/crc32"
	"hash/fnv"
	"runtime"
	"strconv"
	"time"
)

type DB struct {
	StrMap       []map[string][]byte
	ListMap      []map[string]*list.List
	SetMap       []map[string]*data_struct.Set
	ZSkipListMap []map[string]*data_struct.ZSkipList
	tw           *timingwheel.TimingWheel
	ExecQueue    []chan CmdPackage
	funcMap      map[string]func(index uint32, args ...[]byte) interface{}
	UseCpuNum    uint32
	hash32       hash.Hash32
}

type CmdPackage struct {
	Args [][]byte
	Ch   chan interface{}
}

var db *DB

func init() {
	db = new(DB)
	cpuNum := runtime.NumCPU()
	if cpuNum > 2 {
		cpuNum = cpuNum / 2
	}

	db.StrMap = make([]map[string][]byte, cpuNum)
	db.ListMap = make([]map[string]*list.List, cpuNum)
	db.SetMap = make([]map[string]*data_struct.Set, cpuNum)
	db.ZSkipListMap = make([]map[string]*data_struct.ZSkipList, cpuNum)
	db.ExecQueue = make([]chan CmdPackage, cpuNum)
	for i := 0; i < cpuNum; i++ {
		db.StrMap[i] = make(map[string][]byte)
		db.ListMap[i] = make(map[string]*list.List)
		db.SetMap[i] = make(map[string]*data_struct.Set)
		db.ZSkipListMap[i] = make(map[string]*data_struct.ZSkipList)
		db.ExecQueue[i] = make(chan CmdPackage, 1024*16)
	}

	db.tw = event.GetGlobalTimingWheel()
	db.funcMap = map[string]func(index uint32, args ...[]byte) interface{}{
		"set": db.Set, "get": db.Get, "expire": db.Expire, "setex": db.SetEx,
		"setnx": db.SetNx, "del": db.Del, "rpush": db.RPush, "llen": db.LLen,
		"rpop": db.RLPop, "lpop": db.RLPop, "sadd": db.SAdd, "sismember": db.SisMember,
		"smembers": db.SMembers,
	}
	db.UseCpuNum = uint32(cpuNum)
	db.hash32 = fnv.New32()
}

func GetDB() *DB {
	return db
}

func (d *DB) Work() {
	var i uint32
	for i = 0; i < d.UseCpuNum; i++ {
		go d.workOneCore(i)
	}
}

func (d *DB) workOneCore(index uint32) {
	runtime.LockOSThread()
	for {
		cp := <-db.ExecQueue[index]
		f, ok := db.funcMap[string(cp.Args[0])]
		if !ok {
			continue
		}
		val := f(index, cp.Args...)
		if cp.Ch != nil {
			cp.Ch <- val
		}
	}
}

func (d *DB) Hash(bytes []byte) uint32 {
	if d.UseCpuNum == 1 {
		return 0
	}
	//d.hash32.Reset()
	//_, err := d.hash32.Write(bytes)
	//if err != nil {
	//	panic(err)
	//}
	//h := d.hash32.Sum32() % d.UseCpuNum
	//
	//return h

	return crc32.ChecksumIEEE(bytes) % d.UseCpuNum
}

func (d *DB) Set(index uint32, args ...[]byte) interface{} {
	d.StrMap[index][string(args[1])] = args[2]
	return nil
}

func (d *DB) Get(index uint32, args ...[]byte) interface{} {
	val, ok := d.StrMap[index][string(args[1])]
	if !ok {
		return nil
	}
	return val
}

func (d *DB) Expire(index uint32, args ...[]byte) interface{} {
	_, ok := d.StrMap[index][string(args[1])]
	if ok {
		t, _ := strconv.Atoi(string(args[2]))
		d.tw.AfterFunc(time.Duration(t)*time.Second, func() {
			delete(d.StrMap[index], string(args[1]))
		})
		return 1
	}
	return 0
}

func (d *DB) SetEx(index uint32, args ...[]byte) interface{} {
	d.StrMap[index][string(args[1])] = args[3]
	t, _ := strconv.Atoi(string(args[2]))
	d.tw.AfterFunc(time.Duration(t)*time.Second, func() {
		delete(d.StrMap[index], string(args[1]))
	})
	return nil
}

func (d *DB) SetNx(index uint32, args ...[]byte) interface{} {
	_, ok := d.StrMap[index][string(args[1])]
	if ok {
		return 0
	}
	d.StrMap[index][string(args[1])] = args[2]
	return 1
}

func (d *DB) Del(index uint32, args ...[]byte) interface{} {
	_, ok := d.StrMap[index][string(args[1])]
	delete(d.StrMap[index], string(args[1]))

	if !ok {
		return 0
	} else {
		return 1
	}
}

func (d *DB) RPush(index uint32, args ...[]byte) interface{} {
	var l *list.List
	if n, ok := d.ListMap[index][string(args[1])]; ok {
		l = n
	} else {
		l = list.New()
		d.ListMap[index][string(args[1])] = l
	}
	for i := 2; i < len(args); i++ {
		l.PushBack(args[i])
	}
	return l.Len()
}

func (d *DB) LLen(index uint32, args ...[]byte) interface{} {
	if n, ok := d.ListMap[index][string(args[1])]; ok {
		return n.Len()
	} else {
		return 0
	}
}

func (d *DB) RLPop(index uint32, args ...[]byte) interface{} {
	var l *list.List
	var ok bool
	if l, ok = d.ListMap[index][string(args[1])]; !ok {
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
		delete(d.ListMap[index], string(args[1]))
	}
	return res
}

func (d *DB) SAdd(index uint32, args ...[]byte) interface{} {
	var set *data_struct.Set
	var ok bool
	if set, ok = d.SetMap[index][string(args[1])]; !ok {
		set = data_struct.NewSet()
		d.SetMap[index][string(args[1])] = set
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

func (d *DB) SMembers(index uint32, args ...[]byte) interface{} {
	var set *data_struct.Set
	var ok bool
	if set, ok = d.SetMap[index][string(args[1])]; !ok {
		return nil
	}
	return set.GetAllBytes()
}

func (d *DB) SisMember(index uint32, args ...[]byte) interface{} {
	if set, ok := d.SetMap[index][string(args[1])]; ok {
		if set.Has(string(args[2])) {
			return 1
		}
	}
	return 0
}
