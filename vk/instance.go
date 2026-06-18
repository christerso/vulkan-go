package vk

import (
	"runtime"
	"unsafe"
)

// Global and instance-level command pointers.
var (
	vkCreateInstance            func(pCreateInfo, pAllocator, pInstance unsafe.Pointer) Result
	vkDestroyInstance           func(instance Instance, pAllocator unsafe.Pointer)
	vkEnumeratePhysicalDevices  func(instance Instance, pCount *uint32, pDevices *PhysicalDevice) Result
	vkGetPhysicalDeviceProperties func(pd PhysicalDevice, pProps unsafe.Pointer)
	vkGetPhysicalDeviceQueueFamilyProperties func(pd PhysicalDevice, pCount *uint32, pProps *QueueFamilyProperties)
)

func loadGlobalCommands() {
	bindInstanceProc(&vkCreateInstance, 0, "vkCreateInstance")
}

func loadInstanceCommands(instance Instance) {
	h := uintptr(instance)
	bindInstanceProc(&vkDestroyInstance, h, "vkDestroyInstance")
	bindInstanceProc(&vkEnumeratePhysicalDevices, h, "vkEnumeratePhysicalDevices")
	bindInstanceProc(&vkGetPhysicalDeviceProperties, h, "vkGetPhysicalDeviceProperties")
	bindInstanceProc(&vkGetPhysicalDeviceQueueFamilyProperties, h, "vkGetPhysicalDeviceQueueFamilyProperties")
	bindInstanceProc(&vkGetDeviceProcAddr, h, "vkGetDeviceProcAddr")
	loadDeviceLevelInstanceCommands(instance)
	loadSurfaceInstanceCommands(instance)
}

// applicationInfo mirrors VkApplicationInfo. Field order and natural alignment
// match the C struct on amd64.
type applicationInfo struct {
	sType              uint32
	pNext              unsafe.Pointer
	pApplicationName   *byte
	applicationVersion uint32
	pEngineName        *byte
	engineVersion      uint32
	apiVersion         uint32
}

// instanceCreateInfo mirrors VkInstanceCreateInfo.
type instanceCreateInfo struct {
	sType                   uint32
	pNext                   unsafe.Pointer
	flags                   uint32
	pApplicationInfo        *applicationInfo
	enabledLayerCount       uint32
	ppEnabledLayerNames     **byte
	enabledExtensionCount   uint32
	ppEnabledExtensionNames **byte
}

// InstanceConfig describes how to create an instance.
type InstanceConfig struct {
	ApplicationName string
	EngineName      string
	APIVersion      uint32   // 0 selects Vulkan 1.3
	Layers          []string // e.g. "VK_LAYER_KHRONOS_validation"
	Extensions      []string // e.g. surface extensions from the window backend
}

// CreateInstance creates a Vulkan instance. Load must be called first.
func CreateInstance(cfg InstanceConfig) (Instance, error) {
	apiVer := cfg.APIVersion
	if apiVer == 0 {
		apiVer = APIVersion13
	}
	appName := cstr(cfg.ApplicationName)
	engName := cstr(cfg.EngineName)
	app := applicationInfo{
		sType:            stApplicationInfo,
		pApplicationName: appName,
		pEngineName:      engName,
		apiVersion:       apiVer,
	}

	layers, layersPin := cstrArray(cfg.Layers)
	exts, extsPin := cstrArray(cfg.Extensions)

	ci := instanceCreateInfo{
		sType:                 stInstanceCreateInfo,
		pApplicationInfo:      &app,
		enabledLayerCount:     uint32(len(cfg.Layers)),
		ppEnabledLayerNames:   layers,
		enabledExtensionCount: uint32(len(cfg.Extensions)),
		ppEnabledExtensionNames: exts,
	}

	var inst Instance
	res := vkCreateInstance(unsafe.Pointer(&ci), nil, unsafe.Pointer(&inst))
	runtime.KeepAlive(appName)
	runtime.KeepAlive(engName)
	runtime.KeepAlive(layersPin)
	runtime.KeepAlive(extsPin)
	runtime.KeepAlive(&app)
	runtime.KeepAlive(&ci)
	if err := res.asError("vkCreateInstance"); err != nil {
		return 0, err
	}
	loadInstanceCommands(inst)
	return inst, nil
}

// Destroy destroys the instance.
func (i Instance) Destroy() {
	if i != 0 {
		vkDestroyInstance(i, nil)
	}
}

// cstrArray builds a C array of C strings (char**). The returned pin value keeps
// the backing memory reachable and must outlive any C call that reads the array.
func cstrArray(ss []string) (**byte, []*byte) {
	if len(ss) == 0 {
		return nil, nil
	}
	ptrs := make([]*byte, len(ss))
	for i, s := range ss {
		ptrs[i] = cstr(s)
	}
	return &ptrs[0], ptrs
}

// PhysicalDeviceType enumerates VkPhysicalDeviceType.
type PhysicalDeviceType uint32

const (
	DeviceTypeOther         PhysicalDeviceType = 0
	DeviceTypeIntegratedGPU PhysicalDeviceType = 1
	DeviceTypeDiscreteGPU   PhysicalDeviceType = 2
	DeviceTypeVirtualGPU    PhysicalDeviceType = 3
	DeviceTypeCPU           PhysicalDeviceType = 4
)

func (t PhysicalDeviceType) String() string {
	switch t {
	case DeviceTypeIntegratedGPU:
		return "Integrated GPU"
	case DeviceTypeDiscreteGPU:
		return "Discrete GPU"
	case DeviceTypeVirtualGPU:
		return "Virtual GPU"
	case DeviceTypeCPU:
		return "CPU"
	default:
		return "Other"
	}
}

// physicalDeviceProperties mirrors the head of VkPhysicalDeviceProperties. The
// trailing limits and sparseProperties members are not modelled yet; the tail
// pad reserves enough space for the driver to write the full struct. The head
// field offsets (through pipelineCacheUUID) match the C layout exactly.
type physicalDeviceProperties struct {
	apiVersion        uint32
	driverVersion     uint32
	vendorID          uint32
	deviceID          uint32
	deviceType        uint32
	deviceName        [256]byte
	pipelineCacheUUID [16]byte
	_                 [824]byte // VkPhysicalDeviceLimits + VkPhysicalDeviceSparseProperties
}

// DeviceInfo is the decoded subset of physical device properties.
type DeviceInfo struct {
	Name          string
	Type          PhysicalDeviceType
	APIVersion    uint32
	DriverVersion uint32
	VendorID      uint32
	DeviceID      uint32
}

// EnumeratePhysicalDevices returns the physical devices on the instance.
func (i Instance) EnumeratePhysicalDevices() ([]PhysicalDevice, error) {
	var count uint32
	if res := vkEnumeratePhysicalDevices(i, &count, nil); !res.Ok() {
		return nil, res.asError("vkEnumeratePhysicalDevices(count)")
	}
	if count == 0 {
		return nil, nil
	}
	devices := make([]PhysicalDevice, count)
	if res := vkEnumeratePhysicalDevices(i, &count, &devices[0]); !res.Ok() {
		return nil, res.asError("vkEnumeratePhysicalDevices(list)")
	}
	return devices, nil
}

// Info returns the decoded properties of the physical device.
func (pd PhysicalDevice) Info() DeviceInfo {
	var props physicalDeviceProperties
	vkGetPhysicalDeviceProperties(pd, unsafe.Pointer(&props))
	return DeviceInfo{
		Name:          goStr(props.deviceName[:]),
		Type:          PhysicalDeviceType(props.deviceType),
		APIVersion:    props.apiVersion,
		DriverVersion: props.driverVersion,
		VendorID:      props.vendorID,
		DeviceID:      props.deviceID,
	}
}
