package toolkit

import (
	"reflect"
	"unsafe"
)

// SliceToString converts byte slice to a string without memory allocation.
func SliceToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// StringToSlice converts string to a byte slice without memory allocation.
func StringToSlice(s string) (b []byte) {
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bh := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	bh.Data, bh.Len, bh.Cap = sh.Data, sh.Len, sh.Len
	return b
}
func Int64ToArray(num int64) (ret [8]byte) {
	ptr := (*int64)(unsafe.Pointer(&ret[0]))
	*ptr = num
	return
}
func ArrayToSlice(arr uintptr, len int) (ret []byte) {
	ptr := (*reflect.SliceHeader)(unsafe.Pointer(&ret))
	ptr.Len = len
	ptr.Cap = len
	ptr.Data = arr
	return
}
