package data_struct

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestList(t *testing.T) {
	list := &List{
		Head: nil,
		Tail: nil,
		Len:  0,
	}

	Convey("AddNodeHead", t, func() {
		list.AddNodeHead(1)
		So(list.Tail.Value, ShouldEqual, 1)
	})

	Convey("AddNodeTail", t, func() {
		list.AddNodeTail(2)
		So(list.Head.Value, ShouldEqual, 1)
		So(list.Tail.Value, ShouldEqual, 2)
		So(list.Len, ShouldEqual, 2)
	})

	Convey("Index", t, func() {
		So(list.Index(2), ShouldEqual, nil)
		So(list.Index(1).Value, ShouldEqual, 2)
	})

	Convey("InsertNode", t, func() {
		list.InsertNode(list.Head.Next, 5, true)
		So(list.Index(2).Value, ShouldEqual, 5)
	})

	Convey("RotateHeadToTail", t, func() {
		list.RotateHeadToTail()
		So(list.Head.Value, ShouldEqual, 2)
		So(list.Tail.Value, ShouldEqual, 1)
	})

	Convey("RotateTailToHead", t, func() {
		list.RotateTailToHead()
		So(list.Head.Value, ShouldEqual, 1)
		So(list.Tail.Value, ShouldEqual, 5)
	})
}
