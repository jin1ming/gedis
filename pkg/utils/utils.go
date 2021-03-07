package utils

import (
	"reflect"
	"unsafe"
)

func StringToBytes(str string) (bytes []byte) {
	s := *(*reflect.StringHeader)(unsafe.Pointer(&str))
	bs := (*reflect.SliceHeader)(unsafe.Pointer(&bytes))
	bs.Data, bs.Len, bs.Cap = s.Data, s.Len, s.Len
	return bytes
}

func BytesToString(bytes *[]byte) string {
	return *(*string)(unsafe.Pointer(&bytes))
}