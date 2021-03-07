package types

import (
	"testing"
)

func TestSDS(t *testing.T) {
	s := NewSDS("0")
	s.Append("123")
	if s.String() != "0123" {
		t.FailNow()
	}
}

func BenchmarkSDS(b *testing.B) {
	str := NewSDS("")
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