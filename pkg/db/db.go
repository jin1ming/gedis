package db

import (
	"container/list"
	"github.com/jin1ming/Gedis/pkg/data_struct"
)

type DB struct {
	StrMap  map[string][]byte
	ListMap map[string]*list.List
	SetMap  map[string]*data_struct.Set
}

var db *DB

func init() {
	// TODO: 持久化文件读取
	db = new(DB)
	db.StrMap = make(map[string][]byte)
	db.ListMap = make(map[string]*list.List)
	db.SetMap = make(map[string]*data_struct.Set)
}

func GetDB() *DB {
	return db
}
