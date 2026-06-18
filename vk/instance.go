package vk

import (
	"runtime"
	"unsafe"

	vulkan "github.com/christerso/vulkan-go/vulkan"
)

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
	app := vulkan.VkApplicationInfo{
		SType:            vulkan.VkStructureType(stApplicationInfo),
		PApplicationName: unsafe.Pointer(appName),
		PEngineName:      unsafe.Pointer(engName),
		ApiVersion:       apiVer,
	}

	layers, layersPin := cstrArray(cfg.Layers)
	exts, extsPin := cstrArray(cfg.Extensions)

	ci := vulkan.VkInstanceCreateInfo{
		SType:                   vulkan.VkStructureType(stInstanceCreateInfo),
		PApplicationInfo:        unsafe.Pointer(&app),
		EnabledLayerCount:       uint32(len(cfg.Layers)),
		PpEnabledLayerNames:     unsafe.Pointer(layers),
		EnabledExtensionCount:   uint32(len(cfg.Extensions)),
		PpEnabledExtensionNames: unsafe.Pointer(exts),
	}

	var inst vulkan.VkInstance
	res := Result(vulkan.VkCreateInstance(unsafe.Pointer(&ci), nil, unsafe.Pointer(&inst)))
	runtime.KeepAlive(appName)
	runtime.KeepAlive(engName)
	runtime.KeepAlive(layersPin)
	runtime.KeepAlive(extsPin)
	runtime.KeepAlive(&app)
	runtime.KeepAlive(&ci)
	if err := res.asError("vkCreateInstance"); err != nil {
		return 0, err
	}
	vulkan.LoadInstance(inst)
	return Instance(inst), nil
}

// Destroy destroys the instance.
func (i Instance) Destroy() {
	if i != 0 {
		vulkan.VkDestroyInstance(vulkan.VkInstance(i), nil)
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
	if res := Result(vulkan.VkEnumeratePhysicalDevices(vulkan.VkInstance(i), unsafe.Pointer(&count), nil)); !res.Ok() {
		return nil, res.asError("vkEnumeratePhysicalDevices(count)")
	}
	if count == 0 {
		return nil, nil
	}
	devices := make([]PhysicalDevice, count)
	if res := Result(vulkan.VkEnumeratePhysicalDevices(vulkan.VkInstance(i), unsafe.Pointer(&count), unsafe.Pointer(&devices[0]))); !res.Ok() {
		return nil, res.asError("vkEnumeratePhysicalDevices(list)")
	}
	return devices, nil
}

// Info returns the decoded properties of the physical device.
func (pd PhysicalDevice) Info() DeviceInfo {
	var props vulkan.VkPhysicalDeviceProperties
	vulkan.VkGetPhysicalDeviceProperties(vulkan.VkPhysicalDevice(pd), unsafe.Pointer(&props))
	return DeviceInfo{
		Name:          goStr(props.DeviceName[:]),
		Type:          PhysicalDeviceType(props.DeviceType),
		APIVersion:    props.ApiVersion,
		DriverVersion: props.DriverVersion,
		VendorID:      props.VendorID,
		DeviceID:      props.DeviceID,
	}
}
