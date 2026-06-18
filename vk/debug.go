package vk

import (
	"fmt"
	"os"
	"runtime"
	"unsafe"

	"github.com/ebitengine/purego"
)

// Debug messenger severity and type bits.
const (
	debugSeverityVerbose uint32 = 0x00000001
	debugSeverityInfo    uint32 = 0x00000010
	debugSeverityWarning uint32 = 0x00000100
	debugSeverityError   uint32 = 0x00001000

	debugTypeGeneral     uint32 = 0x00000001
	debugTypeValidation  uint32 = 0x00000002
	debugTypePerformance uint32 = 0x00000004
)

// ExtDebugUtils is the instance extension name for the debug messenger.
const ExtDebugUtils = "VK_EXT_debug_utils"

// ValidationLayer is the standard Khronos validation layer name.
const ValidationLayer = "VK_LAYER_KHRONOS_validation"

type debugUtilsMessengerCreateInfo struct {
	sType           uint32
	pNext           unsafe.Pointer
	flags           uint32
	messageSeverity uint32
	messageType     uint32
	pfnUserCallback uintptr
	pUserData       unsafe.Pointer
}

// DebugMessenger wraps a VkDebugUtilsMessengerEXT.
type DebugMessenger struct {
	instance  Instance
	handle    uint64
	callback  uintptr
}

var (
	vkCreateDebugUtilsMessengerEXT  func(instance Instance, pInfo, pAllocator unsafe.Pointer, pMessenger *uint64) Result
	vkDestroyDebugUtilsMessengerEXT func(instance Instance, messenger uint64, pAllocator unsafe.Pointer)
)

// debugCallback is the Go function exposed to Vulkan as a C callback. It reads
// pMessage (offset 40 in VkDebugUtilsMessengerCallbackDataEXT) and prints
// warnings and errors to stderr.
func debugCallback(severity uint32, _ uint32, data uintptr, _ uintptr) uintptr {
	if severity >= debugSeverityWarning && data != 0 {
		msgPtr := *(*uintptr)(unsafe.Pointer(data + 40))
		level := "WARN"
		if severity >= debugSeverityError {
			level = "ERROR"
		}
		fmt.Fprintf(os.Stderr, "[vk %s] %s\n", level, cStringAt(msgPtr))
	}
	return 0 // VK_FALSE: do not abort the call
}

// CreateDebugMessenger installs a debug messenger that reports validation
// warnings and errors. The instance must have been created with the debug utils
// extension enabled.
func (i Instance) CreateDebugMessenger() (DebugMessenger, error) {
	bindInstanceProc(&vkCreateDebugUtilsMessengerEXT, uintptr(i), "vkCreateDebugUtilsMessengerEXT")
	bindInstanceProc(&vkDestroyDebugUtilsMessengerEXT, uintptr(i), "vkDestroyDebugUtilsMessengerEXT")
	cb := purego.NewCallback(debugCallback)
	ci := debugUtilsMessengerCreateInfo{
		sType:           stDebugUtilsMessengerCreateInfoEXT,
		messageSeverity: debugSeverityWarning | debugSeverityError,
		messageType:     debugTypeGeneral | debugTypeValidation | debugTypePerformance,
		pfnUserCallback: cb,
	}
	var handle uint64
	res := vkCreateDebugUtilsMessengerEXT(i, unsafe.Pointer(&ci), nil, &handle)
	runtime.KeepAlive(&ci)
	if err := res.asError("vkCreateDebugUtilsMessengerEXT"); err != nil {
		return DebugMessenger{}, err
	}
	return DebugMessenger{instance: i, handle: handle, callback: cb}, nil
}

// Destroy removes the debug messenger.
func (m DebugMessenger) Destroy() {
	if m.handle != 0 {
		vkDestroyDebugUtilsMessengerEXT(m.instance, m.handle, nil)
	}
}
