package internal

import "unsafe"

var ptrSize int = 0

func PtrSize() int {
	if ptrSize != 0 {
		return ptrSize
	}

	u_ptr_size := unsafe.Sizeof(uintptr(0))
	if u_ptr_size == 4 {
		ptrSize = 4
	} else if u_ptr_size == 8 {
		ptrSize = 8
	} else {
		panic("cannot determine architecture")
	}

	return ptrSize
}
