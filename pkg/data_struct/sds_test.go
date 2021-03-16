package data_struct

import (
	"testing"
)

func TestSds(t *testing.T) {
	s := NewSds("0")
	s.Append("123")
	if s.String() != "0123" {
		t.FailNow()
	}
}

func BenchmarkSds(b *testing.B) {
	str := NewSds("")
	for i := 0; i < b.N; i++ {
		str.Append("010101")
	}
}

func BenchmarkString(b *testing.B) {
	var str string
	for i := 0; i < b.N; i++ {
		str += "010101"
	}
}
