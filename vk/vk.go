// Package vk is an ergonomic Go Vulkan API. It is a thin convenience layer on
// top of the generated raw binding in the sibling vulkan package: vk owns no
// FFI loading of its own. Every command goes through vulkan.Vk* function vars,
// and every C struct passed across the boundary is a generated vulkan.Vk*
// struct, so the ABI is guaranteed to match the one the generator validated.
package vk

import (
	"fmt"

	vulkan "github.com/christerso/vulkan-go/vulkan"
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

// Load opens the Vulkan loader and resolves global commands through the
// generated vulkan package. It is idempotent.
func Load() error {
	if err := vulkan.Load(); err != nil {
		return fmt.Errorf("vk: %w", err)
	}
	return nil
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
