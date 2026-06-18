package vk

import (
	"fmt"
	"runtime"
	"unsafe"

	vulkan "github.com/christerso/vulkan-go/vulkan"
)

// Queue family flag bits (VkQueueFlagBits).
const (
	QueueGraphicsBit uint32 = 0x1
	QueueComputeBit  uint32 = 0x2
	QueueTransferBit uint32 = 0x4
)

// QueueFamilyProperties mirrors VkQueueFamilyProperties.
type QueueFamilyProperties struct {
	QueueFlags                  uint32
	QueueCount                  uint32
	TimestampValidBits          uint32
	MinImageTransferGranularity [3]uint32
}

// QueueFamilies returns the queue family properties of the physical device.
func (pd PhysicalDevice) QueueFamilies() []QueueFamilyProperties {
	var count uint32
	vulkan.VkGetPhysicalDeviceQueueFamilyProperties(vulkan.VkPhysicalDevice(pd), unsafe.Pointer(&count), nil)
	if count == 0 {
		return nil
	}
	families := make([]QueueFamilyProperties, count)
	vulkan.VkGetPhysicalDeviceQueueFamilyProperties(vulkan.VkPhysicalDevice(pd), unsafe.Pointer(&count), unsafe.Pointer(&families[0]))
	return families
}

// GraphicsFamily returns the index of the first queue family supporting
// graphics, or an error if none exists.
func (pd PhysicalDevice) GraphicsFamily() (uint32, error) {
	for i, f := range pd.QueueFamilies() {
		if f.QueueFlags&QueueGraphicsBit != 0 {
			return uint32(i), nil
		}
	}
	return 0, fmt.Errorf("vk: no graphics queue family")
}

// DeviceConfig describes how to create a logical device.
type DeviceConfig struct {
	GraphicsFamily uint32
	Extensions     []string // e.g. "VK_KHR_swapchain"
}

// CreateDevice creates a logical device with a single graphics queue.
func (pd PhysicalDevice) CreateDevice(cfg DeviceConfig) (Device, Queue, error) {
	priority := float32(1.0)
	qci := vulkan.VkDeviceQueueCreateInfo{
		SType:            vulkan.VkStructureType(stDeviceQueueCreateInfo),
		QueueFamilyIndex: cfg.GraphicsFamily,
		QueueCount:       1,
		PQueuePriorities: unsafe.Pointer(&priority),
	}
	exts, extsPin := cstrArray(cfg.Extensions)
	dci := vulkan.VkDeviceCreateInfo{
		SType:                   vulkan.VkStructureType(stDeviceCreateInfo),
		QueueCreateInfoCount:    1,
		PQueueCreateInfos:       unsafe.Pointer(&qci),
		EnabledExtensionCount:   uint32(len(cfg.Extensions)),
		PpEnabledExtensionNames: unsafe.Pointer(exts),
	}
	var device vulkan.VkDevice
	res := Result(vulkan.VkCreateDevice(vulkan.VkPhysicalDevice(pd), unsafe.Pointer(&dci), nil, unsafe.Pointer(&device)))
	runtime.KeepAlive(&priority)
	runtime.KeepAlive(&qci)
	runtime.KeepAlive(&dci)
	runtime.KeepAlive(extsPin)
	if err := res.asError("vkCreateDevice"); err != nil {
		return 0, 0, err
	}
	vulkan.LoadDevice(device)
	var queue vulkan.VkQueue
	vulkan.VkGetDeviceQueue(device, cfg.GraphicsFamily, 0, unsafe.Pointer(&queue))
	return Device(device), Queue(queue), nil
}

// Destroy destroys the logical device.
func (d Device) Destroy() {
	if d != 0 {
		vulkan.VkDestroyDevice(vulkan.VkDevice(d), nil)
	}
}
