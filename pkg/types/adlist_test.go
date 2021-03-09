package types

import "testing"

func TestList(t *testing.T) {
	list := &List{
		Head: nil,
		Tail: nil,
		Len:  0,
	}
	list.AddNodeHead(1)
	list.AddNodeTail(2)
	if list.Length() != 2 {
		t.Fatal("list.Length() is ", list.Length())
	}
	if list.Head.Value != 1 {
		t.Fatal("list.Head.Value is ", list.Index(2).Value)
	}
	if list.Tail.Value != 2 {
		t.Fatal("list.Tail.Value is ", list.Tail.Value)
	}
	list.InsertNode(list.Head.Next, 5, false)
	if list.Index(2).Value != 2 {
		t.Fatal("list.Index(2).Value is", list.Index(2).Value)
	}
	list.RotateHeadToTail()
	if list.Tail.Value != 1 {
		t.Fatal("list.Tail.Value is ", list.Tail.Value)
	}
	list.RotateTailToHead()
	if list.Tail.Value != 2 {
		t.Fatal("list.Tail.Value", list.Tail.Value)
	}
}