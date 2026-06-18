package vk

import "unsafe"

// Color space (VkColorSpaceKHR).
const ColorSpaceSRGBNonlinear uint32 = 0

// SurfaceFormat mirrors VkSurfaceFormatKHR.
type SurfaceFormat struct {
	Format     Format
	ColorSpace uint32
}

// SurfaceCapabilities mirrors VkSurfaceCapabilitiesKHR.
type SurfaceCapabilities struct {
	MinImageCount           uint32
	MaxImageCount           uint32
	CurrentExtent           Extent2D
	MinImageExtent          Extent2D
	MaxImageExtent          Extent2D
	MaxImageArrayLayers     uint32
	SupportedTransforms     uint32
	CurrentTransform        uint32
	SupportedCompositeAlpha uint32
	SupportedUsageFlags     uint32
}

var (
	vkDestroySurfaceKHR                      func(instance Instance, surface SurfaceKHR, pAllocator unsafe.Pointer)
	vkGetPhysicalDeviceSurfaceSupportKHR     func(pd PhysicalDevice, family uint32, surface SurfaceKHR, pSupported *uint32) Result
	vkGetPhysicalDeviceSurfaceCapabilitiesKHR func(pd PhysicalDevice, surface SurfaceKHR, pCaps *SurfaceCapabilities) Result
	vkGetPhysicalDeviceSurfaceFormatsKHR     func(pd PhysicalDevice, surface SurfaceKHR, pCount *uint32, pFormats *SurfaceFormat) Result
	vkGetPhysicalDeviceSurfacePresentModesKHR func(pd PhysicalDevice, surface SurfaceKHR, pCount *uint32, pModes *PresentMode) Result
)

func loadSurfaceInstanceCommands(instance Instance) {
	h := uintptr(instance)
	bindInstanceProc(&vkDestroySurfaceKHR, h, "vkDestroySurfaceKHR")
	bindInstanceProc(&vkGetPhysicalDeviceSurfaceSupportKHR, h, "vkGetPhysicalDeviceSurfaceSupportKHR")
	bindInstanceProc(&vkGetPhysicalDeviceSurfaceCapabilitiesKHR, h, "vkGetPhysicalDeviceSurfaceCapabilitiesKHR")
	bindInstanceProc(&vkGetPhysicalDeviceSurfaceFormatsKHR, h, "vkGetPhysicalDeviceSurfaceFormatsKHR")
	bindInstanceProc(&vkGetPhysicalDeviceSurfacePresentModesKHR, h, "vkGetPhysicalDeviceSurfacePresentModesKHR")
}

// DestroySurface destroys a surface created by the window backend.
func (i Instance) DestroySurface(s SurfaceKHR) {
	if s != 0 {
		vkDestroySurfaceKHR(i, s, nil)
	}
}

// SurfaceSupport reports whether the queue family can present to the surface.
func (pd PhysicalDevice) SurfaceSupport(family uint32, s SurfaceKHR) bool {
	var supported uint32
	vkGetPhysicalDeviceSurfaceSupportKHR(pd, family, s, &supported)
	return supported != 0
}

// SurfaceCapabilities returns the surface capabilities.
func (pd PhysicalDevice) SurfaceCapabilities(s SurfaceKHR) (SurfaceCapabilities, error) {
	var caps SurfaceCapabilities
	res := vkGetPhysicalDeviceSurfaceCapabilitiesKHR(pd, s, &caps)
	return caps, res.asError("vkGetPhysicalDeviceSurfaceCapabilitiesKHR")
}

// SurfaceFormats returns the supported surface formats.
func (pd PhysicalDevice) SurfaceFormats(s SurfaceKHR) ([]SurfaceFormat, error) {
	var count uint32
	if res := vkGetPhysicalDeviceSurfaceFormatsKHR(pd, s, &count, nil); !res.Ok() {
		return nil, res.asError("vkGetPhysicalDeviceSurfaceFormatsKHR(count)")
	}
	if count == 0 {
		return nil, nil
	}
	formats := make([]SurfaceFormat, count)
	res := vkGetPhysicalDeviceSurfaceFormatsKHR(pd, s, &count, &formats[0])
	return formats, res.asError("vkGetPhysicalDeviceSurfaceFormatsKHR(list)")
}

// SurfacePresentModes returns the supported present modes.
func (pd PhysicalDevice) SurfacePresentModes(s SurfaceKHR) ([]PresentMode, error) {
	var count uint32
	if res := vkGetPhysicalDeviceSurfacePresentModesKHR(pd, s, &count, nil); !res.Ok() {
		return nil, res.asError("vkGetPhysicalDeviceSurfacePresentModesKHR(count)")
	}
	if count == 0 {
		return nil, nil
	}
	modes := make([]PresentMode, count)
	res := vkGetPhysicalDeviceSurfacePresentModesKHR(pd, s, &count, &modes[0])
	return modes, res.asError("vkGetPhysicalDeviceSurfacePresentModesKHR(list)")
}

// PresentFamily returns the first queue family that can present to the surface.
func (pd PhysicalDevice) PresentFamily(s SurfaceKHR) (uint32, bool) {
	for i := range pd.QueueFamilies() {
		if pd.SurfaceSupport(uint32(i), s) {
			return uint32(i), true
		}
	}
	return 0, false
}
