// Package vk is a pure-Go Vulkan binding built on purego. It loads the Vulkan
// loader at runtime with dlopen and resolves commands through
// vkGetInstanceProcAddr and vkGetDeviceProcAddr, the same model volk uses in C.
// No cgo and no C compiler are required.
package vk

import (
	"fmt"
	"unsafe"

	"github.com/ebitengine/purego"
)

// Dispatchable handles are pointer-sized. Non-dispatchable handles are uint64
// on every platform per the Vulkan spec.
type (
	Instance       uintptr
	PhysicalDevice uintptr
	Device         uintptr
	Queue          uintptr
	CommandBuffer  uintptr
)

// Result is a VkResult code.
type Result int32

const (
	Success            Result = 0
	NotReady           Result = 1
	Timeout            Result = 2
	EventSet           Result = 3
	EventReset         Result = 4
	Incomplete         Result = 5
	ErrorOutOfHostMem  Result = -1
	ErrorOutOfDeviceMem Result = -2
	ErrorInitFailed    Result = -3
	ErrorDeviceLost    Result = -4
	ErrorExtNotPresent Result = -7
	ErrorFeatureNotPresent Result = -8
	ErrorIncompatibleDriver Result = -9
	SuboptimalKHR      Result = 1000001003
	ErrorOutOfDateKHR  Result = -1000001004
)

// Ok reports whether r is VK_SUCCESS.
func (r Result) Ok() bool { return r == Success }

func (r Result) String() string {
	switch r {
	case Success:
		return "VK_SUCCESS"
	case NotReady:
		return "VK_NOT_READY"
	case Timeout:
		return "VK_TIMEOUT"
	case Incomplete:
		return "VK_INCOMPLETE"
	case ErrorOutOfHostMem:
		return "VK_ERROR_OUT_OF_HOST_MEMORY"
	case ErrorOutOfDeviceMem:
		return "VK_ERROR_OUT_OF_DEVICE_MEMORY"
	case ErrorInitFailed:
		return "VK_ERROR_INITIALIZATION_FAILED"
	case ErrorDeviceLost:
		return "VK_ERROR_DEVICE_LOST"
	case ErrorExtNotPresent:
		return "VK_ERROR_EXTENSION_NOT_PRESENT"
	case ErrorFeatureNotPresent:
		return "VK_ERROR_FEATURE_NOT_PRESENT"
	case ErrorIncompatibleDriver:
		return "VK_ERROR_INCOMPATIBLE_DRIVER"
	case SuboptimalKHR:
		return "VK_SUBOPTIMAL_KHR"
	case ErrorOutOfDateKHR:
		return "VK_ERROR_OUT_OF_DATE_KHR"
	default:
		return fmt.Sprintf("VkResult(%d)", int32(r))
	}
}

// asError returns nil for VK_SUCCESS, otherwise an error naming the result.
func (r Result) asError(op string) error {
	if r == Success {
		return nil
	}
	return fmt.Errorf("%s: %s", op, r)
}

// StructureType values (VkStructureType). Only the ones the binding uses are
// listed; add more as commands are bound.
const (
	stApplicationInfo            uint32 = 0
	stInstanceCreateInfo        uint32 = 1
	stDeviceQueueCreateInfo     uint32 = 2
	stDeviceCreateInfo          uint32 = 3
)

// MakeAPIVersion builds a packed Vulkan version number.
func MakeAPIVersion(variant, major, minor, patch uint32) uint32 {
	return (variant << 29) | (major << 22) | (minor << 12) | patch
}

// API version constants.
var (
	APIVersion10 = MakeAPIVersion(0, 1, 0, 0)
	APIVersion11 = MakeAPIVersion(0, 1, 1, 0)
	APIVersion12 = MakeAPIVersion(0, 1, 2, 0)
	APIVersion13 = MakeAPIVersion(0, 1, 3, 0)
)

// VersionMajor, VersionMinor, VersionPatch unpack a packed version.
func VersionMajor(v uint32) uint32 { return (v >> 22) & 0x7F }
func VersionMinor(v uint32) uint32 { return (v >> 12) & 0x3FF }
func VersionPatch(v uint32) uint32 { return v & 0xFFF }

var (
	libVulkan             uintptr
	vkGetInstanceProcAddr func(instance uintptr, name string) uintptr
	vkGetDeviceProcAddr   func(device uintptr, name string) uintptr
)

// Load opens the Vulkan loader and resolves global commands. It is idempotent.
func Load() error {
	if libVulkan != 0 {
		return nil
	}
	var h uintptr
	var err error
	for _, name := range []string{"libvulkan.so.1", "libvulkan.so"} {
		h, err = purego.Dlopen(name, purego.RTLD_NOW|purego.RTLD_GLOBAL)
		if err == nil && h != 0 {
			break
		}
	}
	if h == 0 {
		return fmt.Errorf("vk: load vulkan loader: %w", err)
	}
	libVulkan = h
	purego.RegisterLibFunc(&vkGetInstanceProcAddr, h, "vkGetInstanceProcAddr")
	loadGlobalCommands()
	return nil
}

// bindInstanceProc resolves a command through vkGetInstanceProcAddr and binds it
// to fptr. instance may be 0 for global commands. It panics if the command is
// missing, since a missing required command is a programming or driver error.
func bindInstanceProc(fptr any, instance uintptr, name string) {
	addr := vkGetInstanceProcAddr(instance, name)
	if addr == 0 {
		panic("vk: missing instance command " + name)
	}
	purego.RegisterFunc(fptr, addr)
}

// bindDeviceProc resolves a device-level command through vkGetDeviceProcAddr.
func bindDeviceProc(fptr any, device uintptr, name string) {
	addr := vkGetDeviceProcAddr(device, name)
	if addr == 0 {
		panic("vk: missing device command " + name)
	}
	purego.RegisterFunc(fptr, addr)
}

// cstr returns a pointer to a NUL-terminated copy of s. The caller must keep the
// returned value reachable (via runtime.KeepAlive) for the duration of any C
// call that reads it.
func cstr(s string) *byte {
	b := make([]byte, len(s)+1)
	copy(b, s)
	return &b[0]
}

// goStr converts a NUL-terminated byte array to a Go string.
func goStr(b []byte) string {
	for i, c := range b {
		if c == 0 {
			return string(b[:i])
		}
	}
	return string(b)
}

// ptr returns an unsafe.Pointer to v, or nil for a nil interface.
func ptr[T any](v *T) unsafe.Pointer { return unsafe.Pointer(v) }
