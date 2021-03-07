package types

import (
	"strconv"
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
		str.Append(strconv.Itoa(i))
	}
}

func BenchmarkString(b *testing.B) {
	var str string
	for i := 0; i < b.N; i++ {
		str += strconv.Itoa(i)
	}
}