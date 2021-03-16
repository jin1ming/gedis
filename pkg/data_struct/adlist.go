package data_struct

import (
	"reflect"
)

type ListNode struct {
	Prev  *ListNode
	Next  *ListNode
	Value interface{}
}

type List struct {
	Head *ListNode
	Tail *ListNode
	Len  uint64
}

func (l *List) Length() uint64 {
	return l.Len
}

func (l *List) AddNodeHead(value interface{}) {
	node := &ListNode{
		Prev:  nil,
		Next:  l.Head,
		Value: value,
	}

	l.Head = node
	if l.Len == 0 {
		l.Tail = node
	}
	l.Len++
}

func (l *List) AddNodeTail(value interface{}) {
	node := &ListNode{
		Prev:  l.Tail,
		Next:  nil,
		Value: value,
	}

	if l.Tail != nil {
		node.Prev.Next = node
		l.Tail = node
	}
	if l.Len == 0 {
		l.Head = node
	}
	l.Len++
}

func (l *List) InsertNode(oldNode *ListNode, value interface{}, after bool) {
	node := &ListNode{
		Prev:  nil,
		Next:  nil,
		Value: value,
	}

	if after {
		node.Prev = oldNode
		node.Next = oldNode.Next
		if l.Tail == oldNode {
			l.Tail = node
		}
	} else {
		node.Next = oldNode
		node.Prev = oldNode.Prev
		if l.Head == oldNode {
			l.Head = node
		}
	}

	if node.Prev != nil {
		node.Prev.Next = node
	}
	if node.Next != nil {
		node.Next.Prev = node
	}

	l.Len++
}

func (l *List) DelNode(node *ListNode) {
	if node.Prev != nil {
		node.Prev.Next = node.Next
	} else {
		l.Head = node.Next
	}

	if node.Next != nil {
		node.Next.Prev = node.Prev
	} else {
		l.Tail = node.Prev
	}

	l.Len--
}

func (l *List) SearchKey(key interface{}) *ListNode {
	iter := l.Head

	for iter != nil {
		if reflect.DeepEqual(iter.Value, key) {
			return iter
		}
	}

	return nil
}

func (l *List) Index(index int64) *ListNode {
	var node *ListNode

	if index < 0 {
		index = -index - 1 // TODO: Why?
		node = l.Tail
		for node != nil {
			node = node.Next
			index--
		}
	} else {
		node = l.Head
		for node != nil && index > 0 {
			node = node.Next
			index--
		}
	}

	return node
}

func (l *List) RotateHeadToTail() {
	if l.Len < 2 {
		return
	}

	head := l.Head

	l.Head = head.Next
	l.Head.Prev = nil

	l.Tail.Next = head
	head.Next = nil
	head.Prev = l.Tail
	l.Tail = head
}

func (l *List) RotateTailToHead() {
	if l.Len < 2 {
		return
	}

	tail := l.Tail
	l.Tail = tail.Prev
	l.Tail.Next = nil

	l.Head.Prev = tail
	tail.Prev = nil
	tail.Next = l.Head
	l.Head = tail
}

func (l *List) Join(o *List) {
	if o.Len == 0 {
		return
	}

	o.Head.Prev = l.Tail

	if l.Tail != nil {
		l.Tail.Next = o.Head
	} else {
		l.Head = o.Head
	}

	l.Tail = o.Tail
	l.Len += o.Len

	o.Head, o.Tail = nil, nil
	o.Len = 0
	o.Len = 0
}
