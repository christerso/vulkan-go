package vk

import (
	"fmt"
	"runtime"
	"unsafe"
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

var (
	vkCreateDevice   func(pd PhysicalDevice, pCreateInfo, pAllocator, pDevice unsafe.Pointer) Result
	vkDestroyDevice  func(device Device, pAllocator unsafe.Pointer)
	vkGetDeviceQueue func(device Device, family, index uint32, pQueue *Queue)
)

// loadDeviceLevelInstanceCommands binds commands that operate on a physical
// device but are resolved through the instance.
func loadDeviceLevelInstanceCommands(instance Instance) {
	h := uintptr(instance)
	bindInstanceProc(&vkCreateDevice, h, "vkCreateDevice")
	bindInstanceProc(&vkGetPhysicalDeviceMemoryProperties, h, "vkGetPhysicalDeviceMemoryProperties")
}

func loadDeviceCommands(device Device) {
	h := uintptr(device)
	bindDeviceProc(&vkDestroyDevice, h, "vkDestroyDevice")
	bindDeviceProc(&vkGetDeviceQueue, h, "vkGetDeviceQueue")
	bindDeviceProc(&vkDeviceWaitIdle, h, "vkDeviceWaitIdle")
	bindDeviceProc(&vkQueueWaitIdle, h, "vkQueueWaitIdle")
	loadSwapchainCommands(device)
	loadMemoryCommands(device)
	loadPipelineCommands(device)
	loadCommandCommands(device)
	loadSyncCommands(device)
}

// QueueFamilies returns the queue family properties of the physical device.
func (pd PhysicalDevice) QueueFamilies() []QueueFamilyProperties {
	var count uint32
	vkGetPhysicalDeviceQueueFamilyProperties(pd, &count, nil)
	if count == 0 {
		return nil
	}
	families := make([]QueueFamilyProperties, count)
	vkGetPhysicalDeviceQueueFamilyProperties(pd, &count, &families[0])
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

type deviceQueueCreateInfo struct {
	sType            uint32
	pNext            unsafe.Pointer
	flags            uint32
	queueFamilyIndex uint32
	queueCount       uint32
	pQueuePriorities *float32
}

type deviceCreateInfo struct {
	sType                   uint32
	pNext                   unsafe.Pointer
	flags                   uint32
	queueCreateInfoCount    uint32
	pQueueCreateInfos       *deviceQueueCreateInfo
	enabledLayerCount       uint32
	ppEnabledLayerNames     **byte
	enabledExtensionCount   uint32
	ppEnabledExtensionNames **byte
	pEnabledFeatures        unsafe.Pointer
}

// DeviceConfig describes how to create a logical device.
type DeviceConfig struct {
	GraphicsFamily uint32
	Extensions     []string // e.g. "VK_KHR_swapchain"
}

// CreateDevice creates a logical device with a single graphics queue.
func (pd PhysicalDevice) CreateDevice(cfg DeviceConfig) (Device, Queue, error) {
	priority := float32(1.0)
	qci := deviceQueueCreateInfo{
		sType:            stDeviceQueueCreateInfo,
		queueFamilyIndex: cfg.GraphicsFamily,
		queueCount:       1,
		pQueuePriorities: &priority,
	}
	exts, extsPin := cstrArray(cfg.Extensions)
	dci := deviceCreateInfo{
		sType:                   stDeviceCreateInfo,
		queueCreateInfoCount:    1,
		pQueueCreateInfos:       &qci,
		enabledExtensionCount:   uint32(len(cfg.Extensions)),
		ppEnabledExtensionNames: exts,
	}
	var device Device
	res := vkCreateDevice(pd, unsafe.Pointer(&dci), nil, unsafe.Pointer(&device))
	runtime.KeepAlive(&priority)
	runtime.KeepAlive(&qci)
	runtime.KeepAlive(&dci)
	runtime.KeepAlive(extsPin)
	if err := res.asError("vkCreateDevice"); err != nil {
		return 0, 0, err
	}
	loadDeviceCommands(device)
	var queue Queue
	vkGetDeviceQueue(device, cfg.GraphicsFamily, 0, &queue)
	return device, queue, nil
}

// Destroy destroys the logical device.
func (d Device) Destroy() {
	if d != 0 {
		vkDestroyDevice(d, nil)
	}
}
