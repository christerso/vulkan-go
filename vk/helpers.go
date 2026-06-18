package vk

import "unsafe"

// ClearColor builds a color ClearValue.
func ClearColor(r, g, b, a float32) ClearValue {
	var cv ClearValue
	f := (*[4]float32)(unsafe.Pointer(&cv[0]))
	f[0], f[1], f[2], f[3] = r, g, b, a
	return cv
}

// ClearDepthStencil builds a depth/stencil ClearValue.
func ClearDepthStencil(depth float32, stencil uint32) ClearValue {
	var cv ClearValue
	*(*float32)(unsafe.Pointer(&cv[0])) = depth
	*(*uint32)(unsafe.Pointer(&cv[4])) = stencil
	return cv
}

// cStringAt reads a NUL-terminated C string at the given address.
func cStringAt(p uintptr) string {
	if p == 0 {
		return ""
	}
	var n int
	for {
		c := *(*byte)(unsafe.Pointer(p + uintptr(n)))
		if c == 0 {
			break
		}
		n++
	}
	return string(unsafe.Slice((*byte)(unsafe.Pointer(p)), n))
}
