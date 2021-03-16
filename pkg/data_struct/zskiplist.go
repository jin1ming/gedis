package data_struct

import (
	"math/rand"
	"time"
)

const (
	ZSKIPLIST_MAXLEVEL = 32
	ZSKIPLIST_P        = 0.25
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
	Len          uint64
	// 所有节点层数的最大值
	level int
}

/* Struct to hold an inclusive/exclusive range spec by score comparison. */
type zRangeSpec struct {
	min, max     float64
	minEx, maxEx bool /* are min or max exclusive? */
}

/* Struct to hold an inclusive/exclusive range spec by lexicographic comparison. */
type zLexRangeSpec struct {
	min, max Sds
	minEx, maxEx bool
}

func NewZSkipList() *zSkipList {
	return &zSkipList{
		header: &zSkipListNode{
			ele:      Sds{},
			score:    0,
			backward: nil,
			level:    make([]zSkipListLevel, ZSKIPLIST_MAXLEVEL),
		},
		tail:  nil,
		Len:   0,
		level: 1, // 表头，哨兵节点，不记录主体数据
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
func (zsl *zSkipList) Insert(score float64, ele Sds) *zSkipListNode {
	update := make([]*zSkipListNode, ZSKIPLIST_MAXLEVEL)
	var x *zSkipListNode
	rank := make([]uint64, ZSKIPLIST_MAXLEVEL)
	var i, level int

	/* 跳表中元素是有序的，需要确定插入的位置 */
	x = zsl.header
	for i = zsl.level - 1; i >= 0; i-- {
		if i == zsl.level-1 {
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

/* Internal function used by Delete, DeleteRangeByScore and
 * DeleteRangeByRank. */
func (zsl *zSkipList) deleteNode(x *zSkipListNode, update []*zSkipListNode) {
	var i int
	for i = 0; i < zsl.level; i++ {
		if update[i].level[i].forward == x {
			update[i].level[i].span += x.level[i].span - 1
			update[i].level[i].forward = x.level[i].forward
		} else {
			update[i].level[i].span -= 1
		}
	}
	if x.level[0].forward != nil {
		x.level[0].forward.backward = x.backward
	} else {
		zsl.tail = x.backward
	}
	for zsl.level > 1 && zsl.header.level[zsl.level-1].forward == nil {
		zsl.level--
	}
	zsl.Len--
}

/* Delete an element with matching score/element from the skiplist.
 * The function returns 1 if the node was found and deleted, otherwise
 * 0 is returned.
 *
 * If 'node' is NULL the deleted node is freed by zslFreeNode(), otherwise
 * it is not freed (but just unlinked) and *node is set to the node pointer,
 * so that it is possible for the caller to reuse the node (including the
 * referenced SDS string at node->ele). */
func (zsl *zSkipList) Delete(score float64, ele Sds) (bool, *zSkipListNode) {
	update := make([]*zSkipListNode, ZSKIPLIST_MAXLEVEL)
	var node *zSkipListNode
	var i int

	x := zsl.header
	for i = zsl.level - 1; i >= 0; i-- {
		for x.level[i].forward != nil &&
			(x.level[i].forward.score < score ||
				(x.level[i].forward.score == score &&
					SdsCompare(x.level[i].forward.ele, ele) < 0)) {
			x = x.level[i].forward
		}
		update[i] = x
	}
	/* We may have multiple elements with the same score, what we need
	 * is to find the element with both the right score and object. */
	x = x.level[0].forward
	if x != nil && score == x.score && SdsCompare(x.ele, ele) == 0 {
		zsl.deleteNode(x, update)
		node = x
		return true, node
	}

	return false, nil /* not found */
}

/* Update the score of an element inside the sorted set skiplist.
 * Note that the element must exist and must match 'score'.
 * This function does not update the score in the hash table side, the
 * caller should take care of it.
 *
 * Note that this function attempts to just update the node, in case after
 * the score update, the node would be exactly at the same position.
 * Otherwise the skiplist is modified by removing and re-adding a new
 * element, which is more costly.
 *
 * The function returns the updated element skiplist node pointer. */
func (zsl *zSkipList) UpdateScore(curScore float64, ele Sds, newScore float64) *zSkipListNode {
	update := make([]*zSkipListNode, ZSKIPLIST_MAXLEVEL)
	var x *zSkipListNode
	var i int

	/* We need to seek to element to update to start: this is useful anyway,
	 * we'll have to update or remove it. */
	x = zsl.header
	for i = zsl.level - 1; i >= 0; i-- {
		for x.level[i].forward != nil &&
			(x.level[i].forward.score < curScore ||
				(x.level[i].forward.score == curScore &&
					SdsCompare(x.level[i].forward.ele, ele) < 0)) {
			x = x.level[i].forward
		}
		update[i] = x
	}

	/* Jump to our element: note that this function assumes that the
	 * element with the matching score exists. */
	x = x.level[0].forward
	if !(x != nil && curScore == x.score && SdsCompare(x.ele, ele) == 0) {
		panic("UpdateScore failed! " +
			"\"x != nil && curScore == x.score && " +
			"SdsCompare(x.ele, ele) == 0\" is false")
	}

	/* If the node, after the score update, would be still exactly
	 * at the same position, we can just update the score without
	 * actually removing and re-inserting the element in the skiplist. */
	if (x.backward == nil || x.backward.score < newScore) &&
		(x.level[0].forward == nil || x.level[0].forward.score > newScore) {
		x.score = newScore
		return x
	}

	/* No way to reuse the old node: we need to remove and insert a new
	 * one at a different place. */
	zsl.deleteNode(x, update)
	newNode := zsl.Insert(newScore, x.ele)
	return newNode
}

/* 判断value是否大于（或大于等于）最小值 */
func (spec *zRangeSpec) isGteMin(value float64) bool {
	if spec.minEx {
		return value > spec.min
	} else {
		return value >= spec.min
	}
}

/* 判断value是否小于（或小于等于）最大值 */
func (spec *zRangeSpec) isLteMax(value float64) bool {
	if spec.maxEx {
		return value < spec.max
	} else {
		return value <= spec.max
	}
}

func (spec *zLexRangeSpec) isGteMin(value Sds) bool {
	if spec.minEx {
		return SdsCompare(value, spec.min) > 0
	}
	return SdsCompare(value, spec.max) >= 0
}

func (spec *zLexRangeSpec) isLteMax(value Sds) bool {
	if spec.maxEx {
		return SdsCompare(value, spec.max) < 0
	}
	return SdsCompare(value, spec.max) <= 0
}

/* Returns if there is a part of the zset is in range. */
func (zsl *zSkipList) IsInRange(r *zRangeSpec) bool {
	if r.min > r.max ||
		(r.min == r.max && (r.minEx || r.maxEx)) {
		return false
	}
	x := zsl.tail
	if x == nil || !r.isGteMin(x.score) {
		return false
	}
	x = zsl.header.level[0].forward
	if x == nil || !r.isLteMax(x.score) {
		return false
	}
	return true
}

func (zsl *zSkipList) IsInLexRange(r *zLexRangeSpec) bool {
	cmp := SdsCompare(r.min, r.max)
	if cmp > 0 || (cmp == 0 && (r.minEx || r.maxEx)) {
		return false
	}
	x := zsl.tail
	if x == nil || !r.isGteMin(x.ele) {
		return false
	}
	x = zsl.header.level[0].forward
	if x == nil || r.isLteMax(x.ele) {
		return false
	}
	return true
}

/* Find the first node that is contained in the specified range.
 * Returns NULL when no element is contained in the range. */
func (zsl *zSkipList) GetFirstInRange(r *zRangeSpec) *zSkipListNode {
	var i int

	if !zsl.IsInRange(r) {
		return nil
	}

	x := zsl.header
	for i = zsl.level - 1; i >= 0; i-- {
		for x.level[i].forward != nil &&
			!r.isGteMin(x.level[i].forward.score) {
			x = x.level[i].forward
		}
	}

	x = x.level[0].forward
	if x == nil {
		panic("GetFirstInRange failed! x is nil.")
	}

	if !r.isLteMax(x.score) {
		return nil
	}
	return x
}

/* Find the last node that is contained in the specified range.
 * Returns NULL when no element is contained in the range. */
func (zsl *zSkipList) GetLastInRange(r *zRangeSpec) *zSkipListNode {
	var i int

	if !zsl.IsInRange(r) {
		return nil
	}

	x := zsl.header
	for i = zsl.level - 1; i >= 0; i-- {
		for x.level[i].forward != nil &&
			r.isLteMax(x.level[i].forward.score) {
			x = x.level[i].forward
		}
	}

	if x == nil {
		panic("GetLastInRange failed! x is nil.")
	}

	if !r.isGteMin(x.score) {
		return nil
	}
	return x
}

/* Delete all the elements with score between min and max from the skiplist.
 * Both min and max can be inclusive or exclusive (see range->minex and
 * range->maxex). When inclusive a score >= min && score <= max is deleted.
 * Note that this function takes the reference to the hash table view of the
 * sorted set, in order to remove the elements from the hash table too. */
func (zsl *zSkipList) DeleteRangeByLex(r *zLexRangeSpec, d *Dict) uint64 {
	update := make([]*zSkipListNode, ZSKIPLIST_MAXLEVEL)
	var removed uint64
	var i int

	x := zsl.header
	for i = zsl.level - 1; i >= 0; i-- {
		for x.level[i].forward != nil &&
			!r.isGteMin(x.level[i].forward.ele) {
			x = x.level[i].forward
		}
		update[i] = x
	}

	x = x.level[0].forward

	for x != nil && r.isLteMax(x.ele) {
		next := x.level[0].forward
		zsl.deleteNode(x, update)
		d.Delete(x.ele)
		removed++
		x = next
	}
	return removed
}


/* Delete all the elements with rank between start and end from the skiplist.
 * Start and end are inclusive. Note that start and end need to be 1-based */
func (zsl *zSkipList) DeleteRangeByRank(start, end uint64, d *Dict) uint64 {
	update := make([]*zSkipListNode, ZSKIPLIST_MAXLEVEL)
	var traversed, removed uint64
	var i int

	x := zsl.header
	for i = zsl.level - 1; i >= 0; i-- {
		for x.level[i].forward != nil && (traversed + x.level[i].span < start) {
			traversed += x.level[i].span
			x = x.level[i].forward
		}
		update[i] = x
	}

	traversed++
	x = x.level[0].forward
	for x != nil && traversed <= end {
		next := x.level[0].forward
		zsl.deleteNode(x, update)
		d.Delete(x.ele)
		removed++
		traversed++
		x = next
	}
	return removed
}

/* Find the rank for an element by both score and key.
 * Returns 0 when the element cannot be found, rank otherwise.
 * Note that the rank is 1-based due to the span of zsl->header to the
 * first element. */
func (zsl *zSkipList) GetRank(score float64, ele Sds) uint64 {
	var rank uint64
	var i int

	x := zsl.header
	for i = zsl.level - 1; i >= 0; i-- {
		for x.level[i].forward != nil &&
			(x.level[i].forward.score < score ||
				(x.level[i].forward.score == score &&
					SdsCompare(x.level[i].forward.ele, ele) <= 0)) {
			rank += x.level[i].span
			x = x.level[i].forward
		}
	}
	if SdsCompare(x.ele, ele) == 0 {
		return rank
	}
	return 0
}

/* Finds an element by its rank. The rank argument needs to be 1-based. */
func (zsl *zSkipList) GetElementByRank(rank uint64) *zSkipListNode {
	var traversed uint64
	var i int

	x := zsl.header
	for i = zsl.level - 1; i >= 0; i-- {
		for x.level[i].forward != nil && (traversed + x.level[i].span) <= rank {
			traversed += x.level[i].span
			x = x.level[i].forward
		}
		if traversed == rank {
			return x
		}
	}
	return nil
}