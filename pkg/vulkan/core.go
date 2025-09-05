// Package vulkan provides working Vulkan API bindings for Go
package vulkan

/*
#cgo windows CFLAGS: -IC:/VulkanSDK/1.4.321.0/Include
#cgo windows LDFLAGS: -LC:/VulkanSDK/1.4.321.0/Lib -lvulkan-1
#cgo linux CFLAGS: -I${VULKAN_SDK}/include
#cgo linux LDFLAGS: -L${VULKAN_SDK}/lib -lvulkan  
#cgo darwin CFLAGS: -I${VULKAN_SDK}/include
#cgo darwin LDFLAGS: -L${VULKAN_SDK}/lib -lMoltenVK

#define VK_USE_PLATFORM_WIN32_KHR 1
#include <vulkan/vulkan.h>
#include <stdlib.h>

uint32_t getVulkanVersion() {
    return VK_API_VERSION_1_3;
}
*/
import "C"
import (
	"fmt"
	"unsafe"
)

// Core Vulkan types - using C types directly
type (
	Instance       C.VkInstance
	PhysicalDevice C.VkPhysicalDevice
	Device         C.VkDevice
	Queue          C.VkQueue
	CommandBuffer  C.VkCommandBuffer
	
	Flags      C.VkFlags
	DeviceSize C.VkDeviceSize
	Bool32     C.VkBool32
	
	Result C.VkResult
)

// Result constants
const (
	SUCCESS                        = Result(C.VK_SUCCESS)
	NOT_READY                      = Result(C.VK_NOT_READY)
	TIMEOUT                        = Result(C.VK_TIMEOUT)
	EVENT_SET                      = Result(C.VK_EVENT_SET)
	EVENT_RESET                    = Result(C.VK_EVENT_RESET)
	INCOMPLETE                     = Result(C.VK_INCOMPLETE)
	ERROR_OUT_OF_HOST_MEMORY       = Result(C.VK_ERROR_OUT_OF_HOST_MEMORY)
	ERROR_OUT_OF_DEVICE_MEMORY     = Result(C.VK_ERROR_OUT_OF_DEVICE_MEMORY)
	ERROR_INITIALIZATION_FAILED    = Result(C.VK_ERROR_INITIALIZATION_FAILED)
	ERROR_DEVICE_LOST              = Result(C.VK_ERROR_DEVICE_LOST)
	ERROR_MEMORY_MAP_FAILED        = Result(C.VK_ERROR_MEMORY_MAP_FAILED)
	ERROR_LAYER_NOT_PRESENT        = Result(C.VK_ERROR_LAYER_NOT_PRESENT)
	ERROR_EXTENSION_NOT_PRESENT    = Result(C.VK_ERROR_EXTENSION_NOT_PRESENT)
	ERROR_FEATURE_NOT_PRESENT      = Result(C.VK_ERROR_FEATURE_NOT_PRESENT)
	ERROR_INCOMPATIBLE_DRIVER      = Result(C.VK_ERROR_INCOMPATIBLE_DRIVER)
	ERROR_TOO_MANY_OBJECTS         = Result(C.VK_ERROR_TOO_MANY_OBJECTS)
	ERROR_FORMAT_NOT_SUPPORTED     = Result(C.VK_ERROR_FORMAT_NOT_SUPPORTED)
	ERROR_FRAGMENTED_POOL          = Result(C.VK_ERROR_FRAGMENTED_POOL)
	ERROR_UNKNOWN                  = Result(C.VK_ERROR_UNKNOWN)
)

// Application info for instance creation
type ApplicationInfo struct {
	PApplicationName   *C.char
	ApplicationVersion uint32
	PEngineName       *C.char
	EngineVersion     uint32
	ApiVersion        uint32
}

// Instance create info
type InstanceCreateInfo struct {
	PApplicationInfo        *ApplicationInfo
	EnabledLayerCount       uint32
	PpEnabledLayerNames     **C.char
	EnabledExtensionCount   uint32
	PpEnabledExtensionNames **C.char
}

// Error interface for Result
func (r Result) Error() string {
	switch r {
	case SUCCESS:
		return "VK_SUCCESS"
	case NOT_READY:
		return "VK_NOT_READY"
	case TIMEOUT:
		return "VK_TIMEOUT"
	case EVENT_SET:
		return "VK_EVENT_SET"
	case EVENT_RESET:
		return "VK_EVENT_RESET"
	case INCOMPLETE:
		return "VK_INCOMPLETE"
	case ERROR_OUT_OF_HOST_MEMORY:
		return "VK_ERROR_OUT_OF_HOST_MEMORY"
	case ERROR_OUT_OF_DEVICE_MEMORY:
		return "VK_ERROR_OUT_OF_DEVICE_MEMORY"
	case ERROR_INITIALIZATION_FAILED:
		return "VK_ERROR_INITIALIZATION_FAILED"
	case ERROR_DEVICE_LOST:
		return "VK_ERROR_DEVICE_LOST"
	case ERROR_MEMORY_MAP_FAILED:
		return "VK_ERROR_MEMORY_MAP_FAILED"
	case ERROR_LAYER_NOT_PRESENT:
		return "VK_ERROR_LAYER_NOT_PRESENT"
	case ERROR_EXTENSION_NOT_PRESENT:
		return "VK_ERROR_EXTENSION_NOT_PRESENT"
	case ERROR_FEATURE_NOT_PRESENT:
		return "VK_ERROR_FEATURE_NOT_PRESENT"
	case ERROR_INCOMPATIBLE_DRIVER:
		return "VK_ERROR_INCOMPATIBLE_DRIVER"
	case ERROR_TOO_MANY_OBJECTS:
		return "VK_ERROR_TOO_MANY_OBJECTS"
	case ERROR_FORMAT_NOT_SUPPORTED:
		return "VK_ERROR_FORMAT_NOT_SUPPORTED"
	case ERROR_FRAGMENTED_POOL:
		return "VK_ERROR_FRAGMENTED_POOL"
	case ERROR_UNKNOWN:
		return "VK_ERROR_UNKNOWN"
	default:
		return fmt.Sprintf("VkResult(%d)", int(r))
	}
}

func (r Result) IsError() bool {
	return r < 0
}

func (r Result) Must() {
	if r.IsError() {
		panic(r)
	}
}

// Core Vulkan functions
func Init() error {
	fmt.Println("Vulkan core API initialized")
	return nil
}

func Destroy() {
	fmt.Println("Vulkan core API destroyed")
}

func GetVersion() uint32 {
	return uint32(C.getVulkanVersion())
}

func CreateInstance(createInfo *InstanceCreateInfo, allocator unsafe.Pointer, instance *Instance) Result {
	cCreateInfo := C.VkInstanceCreateInfo{
		sType: C.VK_STRUCTURE_TYPE_INSTANCE_CREATE_INFO,
		pNext: nil,
		flags: 0,
	}
	
	var cAppInfo C.VkApplicationInfo
	if createInfo.PApplicationInfo != nil {
		cAppInfo = C.VkApplicationInfo{
			sType:              C.VK_STRUCTURE_TYPE_APPLICATION_INFO,
			pNext:              nil,
			pApplicationName:   createInfo.PApplicationInfo.PApplicationName,
			applicationVersion: C.uint32_t(createInfo.PApplicationInfo.ApplicationVersion),
			pEngineName:       createInfo.PApplicationInfo.PEngineName,
			engineVersion:     C.uint32_t(createInfo.PApplicationInfo.EngineVersion),
			apiVersion:        C.uint32_t(createInfo.PApplicationInfo.ApiVersion),
		}
		cCreateInfo.pApplicationInfo = &cAppInfo
	}
	
	cCreateInfo.enabledLayerCount = C.uint32_t(createInfo.EnabledLayerCount)
	cCreateInfo.ppEnabledLayerNames = createInfo.PpEnabledLayerNames
	cCreateInfo.enabledExtensionCount = C.uint32_t(createInfo.EnabledExtensionCount)
	cCreateInfo.ppEnabledExtensionNames = createInfo.PpEnabledExtensionNames
	
	result := C.vkCreateInstance(&cCreateInfo, (*C.VkAllocationCallbacks)(allocator), (*C.VkInstance)(unsafe.Pointer(instance)))
	return Result(result)
}

func DestroyInstance(instance Instance, allocator unsafe.Pointer) {
	C.vkDestroyInstance(C.VkInstance(instance), (*C.VkAllocationCallbacks)(allocator))
}

func EnumeratePhysicalDevices(instance Instance, deviceCount *uint32, devices *PhysicalDevice) Result {
	result := C.vkEnumeratePhysicalDevices(
		C.VkInstance(instance),
		(*C.uint32_t)(unsafe.Pointer(deviceCount)),
		(*C.VkPhysicalDevice)(unsafe.Pointer(devices)))
	return Result(result)
}

func GetPhysicalDeviceProperties(physicalDevice PhysicalDevice, properties unsafe.Pointer) {
	C.vkGetPhysicalDeviceProperties(C.VkPhysicalDevice(physicalDevice), (*C.VkPhysicalDeviceProperties)(properties))
}

func GetPhysicalDeviceQueueFamilyProperties(physicalDevice PhysicalDevice, queueFamilyCount *uint32, queueFamilies unsafe.Pointer) {
	C.vkGetPhysicalDeviceQueueFamilyProperties(
		C.VkPhysicalDevice(physicalDevice),
		(*C.uint32_t)(unsafe.Pointer(queueFamilyCount)),
		(*C.VkQueueFamilyProperties)(queueFamilies))
}

func CreateDevice(physicalDevice PhysicalDevice, createInfo unsafe.Pointer, allocator unsafe.Pointer, device *Device) Result {
	result := C.vkCreateDevice(
		C.VkPhysicalDevice(physicalDevice),
		(*C.VkDeviceCreateInfo)(createInfo),
		(*C.VkAllocationCallbacks)(allocator),
		(*C.VkDevice)(unsafe.Pointer(device)))
	return Result(result)
}

func DestroyDevice(device Device, allocator unsafe.Pointer) {
	C.vkDestroyDevice(C.VkDevice(device), (*C.VkAllocationCallbacks)(allocator))
}

func GetDeviceQueue(device Device, queueFamilyIndex uint32, queueIndex uint32, queue *Queue) {
	C.vkGetDeviceQueue(
		C.VkDevice(device),
		C.uint32_t(queueFamilyIndex),
		C.uint32_t(queueIndex),
		(*C.VkQueue)(unsafe.Pointer(queue)))
}

func DeviceWaitIdle(device Device) Result {
	result := C.vkDeviceWaitIdle(C.VkDevice(device))
	return Result(result)
}

// String utilities
func CString(s string) *C.char {
	return C.CString(s)
}

func FreeCString(cstr *C.char) {
	C.free(unsafe.Pointer(cstr))
}

func CStringSlice(strs []string) []*C.char {
	if len(strs) == 0 {
		return nil
	}
	
	cstrs := make([]*C.char, len(strs))
	for i, str := range strs {
		cstrs[i] = CString(str)
	}
	return cstrs
}

func FreeCStringSlice(cstrs []*C.char) {
	for _, cstr := range cstrs {
		FreeCString(cstr)
	}
}