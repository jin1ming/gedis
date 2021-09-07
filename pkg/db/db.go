package db

type DB struct {
	DataMap		map[string]interface{}
}

var db *DB

func init() {
	// TODO: 持久化文件读取
	db = &DB{DataMap: make(map[string]interface{})}
}

func GetDB() *DB {
	return db
}