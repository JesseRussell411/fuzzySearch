package main

import "unsafe"

func unsafeBoundlessSliceGet_int(slc []int, i uintptr) int {
	start := uintptr(unsafe.Pointer(&slc[0]))
	target := start + unsafe.Sizeof(int(0))*i
	targetPointer := unsafe.Pointer(target)
	targetValue := *(*int)(targetPointer)
	return targetValue
}

func unsafeBoundlessSliceSet_int(slc []int, i uintptr, value int) {
	start := uintptr(unsafe.Pointer(&slc[0]))
	target := start + unsafe.Sizeof(int(0))*i
	targetPointer := unsafe.Pointer(target)
	*(*int)(targetPointer) = value
}

func unsafeBoundlessStringGet(str string, i uintptr) byte {
	data := unsafe.StringData(str)
	target := uintptr(unsafe.Pointer(data)) + i
	targetPointer := unsafe.Pointer(target)
	targetValue := *(*byte)(targetPointer)
	return targetValue
}
