package vk

import (
	"fmt"
	"os"
	"runtime"
	"unsafe"

	"github.com/ebitengine/purego"

	vulkan "github.com/christerso/vulkan-go/vulkan"
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

// DebugMessenger wraps a VkDebugUtilsMessengerEXT.
type DebugMessenger struct {
	instance Instance
	handle   uint64
	callback uintptr
}

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
// extension enabled. The create/destroy commands are resolved through the
// generated vulkan package (vulkan.LoadInstance binds them).
func (i Instance) CreateDebugMessenger() (DebugMessenger, error) {
	cb := purego.NewCallback(debugCallback)
	ci := vulkan.VkDebugUtilsMessengerCreateInfoEXT{
		SType:           vulkan.VkStructureType(stDebugUtilsMessengerCreateInfoEXT),
		MessageSeverity: debugSeverityWarning | debugSeverityError,
		MessageType:     debugTypeGeneral | debugTypeValidation | debugTypePerformance,
		PfnUserCallback: cb,
	}
	var handle vulkan.VkDebugUtilsMessengerEXT
	res := Result(vulkan.VkCreateDebugUtilsMessengerEXT(vulkan.VkInstance(i), unsafe.Pointer(&ci), nil, unsafe.Pointer(&handle)))
	runtime.KeepAlive(&ci)
	if err := res.asError("vkCreateDebugUtilsMessengerEXT"); err != nil {
		return DebugMessenger{}, err
	}
	return DebugMessenger{instance: i, handle: handle, callback: cb}, nil
}

// Destroy removes the debug messenger.
func (m DebugMessenger) Destroy() {
	if m.handle != 0 {
		vulkan.VkDestroyDebugUtilsMessengerEXT(vulkan.VkInstance(m.instance), m.handle, nil)
	}
}
