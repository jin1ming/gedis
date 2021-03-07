package types

import (
	"strconv"
	"testing"
)

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