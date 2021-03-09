package types

import (
	"math/rand"
	"time"
)

const (
	ZSKIPLIST_MAXLEVEL = 32
	ZSKIPLIST_P = 0.25
)

type zSkipListLevel struct {
	// 前进指针
	forward *zSkipListNode
	// 当前层跨越的节点数量，用于计算元素排名
	span uint64
}

type zSkipListNode struct {
	// redis4之前是robj，实际还是string robj，它可能存放long(int64)，
	// 但zadd命令在将数据插入到skiplist里面之前先进行了解码，故存放的一定是sds
	ele Sds
	// 分值
	score float64
	// 后退指针
	backward *zSkipListNode
	// 存放指向各层链表后一个节点的指针（后向指针）
	level []zSkipListLevel
}

type zSkipList struct {
	// 只有第一层链表是双向链表
	header, tail *zSkipListNode
	Len uint64
	// 所有节点层数的最大值
	level int
}

func Create() *zSkipList {
	return &zSkipList{
		header: &zSkipListNode{
			ele:      Sds{},
			score:    0,
			backward: nil,
			level:    make([]zSkipListLevel, ZSKIPLIST_MAXLEVEL),
		},
		tail:   nil,
		Len:    0,
		level:  1, // 表头，哨兵节点，不记录主体数据
	}
}

/* Returns a random level for the new skiplist node we are going to create.
 * The return value of this function is between 1 and ZSKIPLIST_MAXLEVEL
 * (both inclusive), with a powerlaw-alike distribution where higher
 * levels are less likely to be returned. */
/* 随机生成一个1-32之间的值作为level数组的大小
 * 一个节点有第i层的指针，那么它有第i+1层概率的指针是ZSKIPLIST_P
 * 节点的最大层数不允许超过ZSKIPLIST_MAXLEVEL */
func RandomLevel() int {
	level := 1
	rand.Seed(time.Now().UnixNano())
	tmp := ZSKIPLIST_P * 0xffff
	for rand.Intn(0xffff) < int(tmp) {
		level++
	}
	if level > ZSKIPLIST_MAXLEVEL {
		return ZSKIPLIST_MAXLEVEL
	}
	return level
}

/* Insert a new node in the skiplist. Assumes the element does not already
 * exist (up to the caller to enforce that). The skiplist takes ownership
 * of the passed Sds string 'ele'. */
func (zsl *zSkipList)Insert(score float64, ele Sds) *zSkipListNode {
	update := make([]*zSkipListNode, ZSKIPLIST_MAXLEVEL)
	var x *zSkipListNode
	rank := make([]uint64, ZSKIPLIST_MAXLEVEL)
	var i, level int

	/* 跳表中元素是有序的，需要确定插入的位置 */
	x = zsl.header
	for i = zsl.level - 1; i >= 0; i-- {
		if i == zsl.level - 1 {
			rank[i] = 0
		} else {
			rank[i] = rank[i+1]
		}
		for x.level[i].forward != nil &&
			(x.level[i].forward.score < score ||
				(x.level[i].forward.score == score &&
				SdsCompare(x.level[i].forward.ele, ele) < 0)) {
			rank[i] += x.level[i].span
			x = x.level[i].forward
		}
		update[i] = x
	}
	/* we assume the element is not already inside, since we allow duplicated
	 * scores, reinserting the same element should never happen since the
	 * caller of zslInsert() should test in the hash table if the element is
	 * already inside or not. */
	level = RandomLevel()
	if level > zsl.level {
		for i = zsl.level; i < level; i++ {
			rank[i] = 0
			update[i] = zsl.header
			update[i].level[i].span = zsl.Len
		}
		zsl.level = level
	}
	x = &zSkipListNode{
		ele:      ele,
		score:    score,
		backward: nil,
		level:    make([]zSkipListLevel, level),
	}
	for i := 0; i < level; i++ {
		x.level[i].forward = update[i].level[i].forward
		update[i].level[i].forward = x

		/* update span covered by update[i] as x is inserted here */
		x.level[i].span = update[i].level[i].span - (rank[0] - rank[i])
		update[i].level[i].span = rank[0] - rank[i] + 1
	}

	for i := level; i < zsl.level; i++ {
		update[i].level[i].span++
	}

	if update[0] != zsl.header {
		x.backward = update[0]
	}
	if x.level[0].forward != nil {
		x.level[0].forward.backward = x
	} else {
		zsl.tail = x
	}
	zsl.Len++
	return x
}