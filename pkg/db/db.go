package db

import (
	"container/list"
	"github.com/RussellLuo/timingwheel"
	"github.com/jin1ming/Gedis/pkg/data_struct"
	"github.com/jin1ming/Gedis/pkg/event"
	"strconv"
	"time"
)

type DB struct {
	StrMap       map[string][]byte
	ListMap      map[string]*list.List
	SetMap       map[string]*data_struct.Set
	ZSkipListMap map[string]*data_struct.ZSkipList
	tw           *timingwheel.TimingWheel
	returnBuffer chan any
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
}

func GetDB() *DB {
	return db
}

func (d *DB) Set(args ...[]byte) {
	d.StrMap[string(args[1])] = args[2]
}

func (d *DB) Get(args ...[]byte) ([]byte, bool) {
	val, ok := d.StrMap[string(args[1])]
	return val, ok
}

func (d *DB) Expire(args ...[]byte) int {
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

func (d *DB) SetEx(args ...[]byte) {
	d.StrMap[string(args[1])] = args[3]
	t, _ := strconv.Atoi(string(args[2]))
	d.tw.AfterFunc(time.Duration(t)*time.Second, func() {
		delete(d.StrMap, string(args[1]))
	})
}

func (d *DB) SetNx(args ...[]byte) int {
	_, ok := d.StrMap[string(args[1])]
	if ok {
		return 0
	}
	d.StrMap[string(args[1])] = args[2]
	return 1
}

func (d *DB) Del(args ...[]byte) int {
	_, ok := d.StrMap[string(args[1])]
	delete(d.StrMap, string(args[1]))

	if !ok {
		return 0
	} else {
		return 1
	}
}

func (d *DB) RPush(args ...[]byte) int {
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

func (d *DB) LLen(args ...[]byte) int {
	if n, ok := d.ListMap[string(args[1])]; ok {
		return n.Len()
	} else {
		return 0
	}
}

func (d *DB) RLPop(args ...[]byte) []byte {
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

func (d *DB) SAdd(args ...[]byte) int {
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

func (d *DB) SMembers(args ...[]byte) [][]byte {
	var set *data_struct.Set
	var ok bool
	if set, ok = d.SetMap[string(args[1])]; !ok {
		return nil
	}
	return set.GetAllBytes()
}

func (d *DB) SisMember(args ...[]byte) int {
	if set, ok := d.SetMap[string(args[1])]; ok {
		if set.Has(string(args[2])) {
			return 1
		}
	}
	return 0
}
