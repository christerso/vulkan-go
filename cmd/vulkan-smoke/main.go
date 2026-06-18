// Command vulkan-smoke exercises the generated raw `vulkan` package end to end:
// it loads the loader, creates an instance with the Khronos validation layer and
// the debug-utils messenger, enumerates physical devices, prints the GPU name,
// picks a graphics queue family, creates a device, and fetches a queue. A silent
// validation layer proves the generated struct ABI matches the C layout.
package main

import (
	"fmt"
	"os"
	"unsafe"

	"github.com/ebitengine/purego"
	vk "github.com/christerso/vulkan-go/vulkan"
)

// cstr returns a pointer to a NUL-terminated copy of s.
func cstr(s string) *byte {
	b := make([]byte, len(s)+1)
	copy(b, s)
	return &b[0]
}

// goStr reads a NUL-terminated C string from a byte slice.
func goStr(b []byte) string {
	for i, c := range b {
		if c == 0 {
			return string(b[:i])
		}
	}
	return string(b)
}

var validationErrors int

// debugCallback is registered as the VkDebugUtilsMessengerEXT callback. Vulkan
// calls it with (severity, types, pCallbackData, pUserData). It prints every
// message to stderr and counts warnings/errors.
func debugCallback(severity uint32, _ uint32, data uintptr, _ uintptr) uintptr {
	// Only surface warnings and errors. INFO/VERBOSE is loader trace noise and
	// would drown the signal; a real ABI mismatch shows up as a WARN or ERROR.
	if severity < uint32(vk.VK_DEBUG_UTILS_MESSAGE_SEVERITY_WARNING_BIT_EXT) || data == 0 {
		return 0
	}
	d := (*vk.VkDebugUtilsMessengerCallbackDataEXT)(unsafe.Pointer(data))
	msg := ""
	if d.PMessage != nil {
		msg = goStr(unsafe.Slice((*byte)(d.PMessage), 4096))
	}
	level := "WARN"
	if severity >= uint32(vk.VK_DEBUG_UTILS_MESSAGE_SEVERITY_ERROR_BIT_EXT) {
		level = "ERROR"
	}
	validationErrors++
	fmt.Fprintf(os.Stderr, "[VALIDATION %s] %s\n", level, msg)
	return 0 // VK_FALSE
}

func die(msg string, r vk.VkResult) {
	fmt.Fprintf(os.Stderr, "FAIL: %s (VkResult=%d)\n", msg, int32(r))
	os.Exit(1)
}

func main() {
	if err := vk.Load(); err != nil {
		fmt.Fprintln(os.Stderr, "FAIL: Load:", err)
		os.Exit(1)
	}

	appName := cstr("vulkan-smoke")
	app := vk.VkApplicationInfo{
		SType:            vk.VK_STRUCTURE_TYPE_APPLICATION_INFO,
		PApplicationName: unsafe.Pointer(appName),
		ApiVersion:       vkMakeVersion(1, 3, 0),
	}

	layer := cstr("VK_LAYER_KHRONOS_validation")
	ext := cstr("VK_EXT_debug_utils")
	layers := []*byte{layer}
	exts := []*byte{ext}

	cb := purego.NewCallback(debugCallback)
	dbg := vk.VkDebugUtilsMessengerCreateInfoEXT{
		SType: vk.VK_STRUCTURE_TYPE_DEBUG_UTILS_MESSENGER_CREATE_INFO_EXT,
		MessageSeverity: uint32(vk.VK_DEBUG_UTILS_MESSAGE_SEVERITY_VERBOSE_BIT_EXT) |
			uint32(vk.VK_DEBUG_UTILS_MESSAGE_SEVERITY_INFO_BIT_EXT) |
			uint32(vk.VK_DEBUG_UTILS_MESSAGE_SEVERITY_WARNING_BIT_EXT) |
			uint32(vk.VK_DEBUG_UTILS_MESSAGE_SEVERITY_ERROR_BIT_EXT),
		MessageType: uint32(vk.VK_DEBUG_UTILS_MESSAGE_TYPE_GENERAL_BIT_EXT) |
			uint32(vk.VK_DEBUG_UTILS_MESSAGE_TYPE_VALIDATION_BIT_EXT) |
			uint32(vk.VK_DEBUG_UTILS_MESSAGE_TYPE_PERFORMANCE_BIT_EXT),
		PfnUserCallback: cb,
	}

	ci := vk.VkInstanceCreateInfo{
		SType:                   vk.VK_STRUCTURE_TYPE_INSTANCE_CREATE_INFO,
		PNext:                   unsafe.Pointer(&dbg), // also validate instance create/destroy
		PApplicationInfo:        unsafe.Pointer(&app),
		EnabledLayerCount:       1,
		PpEnabledLayerNames:     unsafe.Pointer(&layers[0]),
		EnabledExtensionCount:   1,
		PpEnabledExtensionNames: unsafe.Pointer(&exts[0]),
	}

	var instance vk.VkInstance
	if r := vk.VkCreateInstance(unsafe.Pointer(&ci), nil, unsafe.Pointer(&instance)); r != 0 {
		die("vkCreateInstance", r)
	}
	vk.LoadInstance(instance)

	// Create the standalone messenger too.
	var messenger vk.VkDebugUtilsMessengerEXT
	if vk.VkCreateDebugUtilsMessengerEXT != nil {
		if r := vk.VkCreateDebugUtilsMessengerEXT(instance, unsafe.Pointer(&dbg), nil, unsafe.Pointer(&messenger)); r != 0 {
			die("vkCreateDebugUtilsMessengerEXT", r)
		}
	}

	var count uint32
	if r := vk.VkEnumeratePhysicalDevices(instance, unsafe.Pointer(&count), nil); r != 0 {
		die("vkEnumeratePhysicalDevices(count)", r)
	}
	if count == 0 {
		fmt.Fprintln(os.Stderr, "FAIL: no physical devices")
		os.Exit(1)
	}
	devices := make([]vk.VkPhysicalDevice, count)
	if r := vk.VkEnumeratePhysicalDevices(instance, unsafe.Pointer(&count), unsafe.Pointer(&devices[0])); r != 0 {
		die("vkEnumeratePhysicalDevices(list)", r)
	}

	gpu := devices[0]
	var props vk.VkPhysicalDeviceProperties
	vk.VkGetPhysicalDeviceProperties(gpu, unsafe.Pointer(&props))
	name := goStr(props.DeviceName[:])
	fmt.Printf("GPU: %s\n", name)
	fmt.Printf("API: %d.%d.%d  Driver: 0x%X  Vendor: 0x%X  Device: 0x%X\n",
		(props.ApiVersion>>22)&0x7F, (props.ApiVersion>>12)&0x3FF, props.ApiVersion&0xFFF,
		props.DriverVersion, props.VendorID, props.DeviceID)

	// Pick a graphics queue family.
	var qfCount uint32
	vk.VkGetPhysicalDeviceQueueFamilyProperties(gpu, unsafe.Pointer(&qfCount), nil)
	qfams := make([]vk.VkQueueFamilyProperties, qfCount)
	vk.VkGetPhysicalDeviceQueueFamilyProperties(gpu, unsafe.Pointer(&qfCount), unsafe.Pointer(&qfams[0]))
	gfxFamily := uint32(0xFFFFFFFF)
	for i, qf := range qfams {
		if qf.QueueFlags&uint32(vk.VK_QUEUE_GRAPHICS_BIT) != 0 {
			gfxFamily = uint32(i)
			break
		}
	}
	if gfxFamily == 0xFFFFFFFF {
		fmt.Fprintln(os.Stderr, "FAIL: no graphics queue family")
		os.Exit(1)
	}
	fmt.Printf("Graphics queue family: %d\n", gfxFamily)

	priority := float32(1.0)
	qci := vk.VkDeviceQueueCreateInfo{
		SType:            vk.VK_STRUCTURE_TYPE_DEVICE_QUEUE_CREATE_INFO,
		QueueFamilyIndex: gfxFamily,
		QueueCount:       1,
		PQueuePriorities: unsafe.Pointer(&priority),
	}
	dci := vk.VkDeviceCreateInfo{
		SType:                vk.VK_STRUCTURE_TYPE_DEVICE_CREATE_INFO,
		QueueCreateInfoCount: 1,
		PQueueCreateInfos:    unsafe.Pointer(&qci),
	}
	var device vk.VkDevice
	if r := vk.VkCreateDevice(gpu, unsafe.Pointer(&dci), nil, unsafe.Pointer(&device)); r != 0 {
		die("vkCreateDevice", r)
	}
	vk.LoadDevice(device)

	var queue vk.VkQueue
	vk.VkGetDeviceQueue(device, gfxFamily, 0, unsafe.Pointer(&queue))
	if queue == 0 {
		fmt.Fprintln(os.Stderr, "FAIL: null queue")
		os.Exit(1)
	}
	fmt.Printf("Device created, graphics queue acquired: 0x%X\n", queue)

	// Tear down (exercises destroy paths under validation).
	if vk.VkDestroyDevice != nil {
		vk.VkDestroyDevice(device, nil)
	}
	if messenger != 0 && vk.VkDestroyDebugUtilsMessengerEXT != nil {
		vk.VkDestroyDebugUtilsMessengerEXT(instance, messenger, nil)
	}
	if vk.VkDestroyInstance != nil {
		vk.VkDestroyInstance(instance, nil)
	}

	runtimeKeepAlive(appName, layer, ext, &app, &dbg, &ci, &qci, &dci, &priority)

	if validationErrors > 0 {
		fmt.Fprintf(os.Stderr, "FAIL: %d validation warning(s)/error(s)\n", validationErrors)
		os.Exit(1)
	}
	fmt.Println("SUCCESS: device created with silent validation")
}

// vkMakeVersion packs a Vulkan API version.
func vkMakeVersion(major, minor, patch uint32) uint32 {
	return (major << 22) | (minor << 12) | patch
}

// runtimeKeepAlive prevents the GC from reclaiming pinned objects before the C
// calls that read them have completed. It is a no-op at runtime but creates a
// reference the compiler must honor.
//
//go:noinline
func runtimeKeepAlive(...any) {}
