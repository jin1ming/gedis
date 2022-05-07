package cluster

import (
	"hash/crc32"
	"sort"
	"strconv"
)

type ConHashMap struct {
	hash     func(data []byte) uint32
	replicas int
	keys     []int
	Map      map[int]string
}

func NewConHashMap(replicas int, fn uint32) *ConHashMap {
	return &ConHashMap{
		hash:     crc32.ChecksumIEEE,
		replicas: replicas,
		keys:     nil,
		Map:      make(map[int]string),
	}
}

func (c *ConHashMap) Add(keys ...string) {
	for _, key := range keys {
		for i := 0; i < c.replicas; i++ {
			hash := int(c.hash([]byte(strconv.Itoa(i) + key)))
			c.keys = append(c.keys, hash)
			c.Map[hash] = key
		}
	}
	sort.Ints(c.keys)
}

func (c *ConHashMap) Get(key string) string {
	if len(c.keys) == 0 {
		return ""
	}
	hash := int(c.hash([]byte(key)))
	idx := sort.Search(len(c.keys), func(i int) bool {
		return c.keys[i] >= hash
	})

	return c.Map[c.keys[idx%len(c.keys)]]
}
