package vk

import (
	"unsafe"

	vulkan "github.com/christerso/vulkan-go/vulkan"
)

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

// DestroySurface destroys a surface created by the window backend.
func (i Instance) DestroySurface(s SurfaceKHR) {
	if s != 0 {
		vulkan.VkDestroySurfaceKHR(vulkan.VkInstance(i), vulkan.VkSurfaceKHR(s), nil)
	}
}

// SurfaceSupport reports whether the queue family can present to the surface.
func (pd PhysicalDevice) SurfaceSupport(family uint32, s SurfaceKHR) bool {
	var supported uint32
	vulkan.VkGetPhysicalDeviceSurfaceSupportKHR(vulkan.VkPhysicalDevice(pd), family, vulkan.VkSurfaceKHR(s), unsafe.Pointer(&supported))
	return supported != 0
}

// SurfaceCapabilities returns the surface capabilities.
func (pd PhysicalDevice) SurfaceCapabilities(s SurfaceKHR) (SurfaceCapabilities, error) {
	var caps SurfaceCapabilities
	res := Result(vulkan.VkGetPhysicalDeviceSurfaceCapabilitiesKHR(vulkan.VkPhysicalDevice(pd), vulkan.VkSurfaceKHR(s), unsafe.Pointer(&caps)))
	return caps, res.asError("vkGetPhysicalDeviceSurfaceCapabilitiesKHR")
}

// SurfaceFormats returns the supported surface formats.
func (pd PhysicalDevice) SurfaceFormats(s SurfaceKHR) ([]SurfaceFormat, error) {
	var count uint32
	if res := Result(vulkan.VkGetPhysicalDeviceSurfaceFormatsKHR(vulkan.VkPhysicalDevice(pd), vulkan.VkSurfaceKHR(s), unsafe.Pointer(&count), nil)); !res.Ok() {
		return nil, res.asError("vkGetPhysicalDeviceSurfaceFormatsKHR(count)")
	}
	if count == 0 {
		return nil, nil
	}
	formats := make([]SurfaceFormat, count)
	res := Result(vulkan.VkGetPhysicalDeviceSurfaceFormatsKHR(vulkan.VkPhysicalDevice(pd), vulkan.VkSurfaceKHR(s), unsafe.Pointer(&count), unsafe.Pointer(&formats[0])))
	return formats, res.asError("vkGetPhysicalDeviceSurfaceFormatsKHR(list)")
}

// SurfacePresentModes returns the supported present modes.
func (pd PhysicalDevice) SurfacePresentModes(s SurfaceKHR) ([]PresentMode, error) {
	var count uint32
	if res := Result(vulkan.VkGetPhysicalDeviceSurfacePresentModesKHR(vulkan.VkPhysicalDevice(pd), vulkan.VkSurfaceKHR(s), unsafe.Pointer(&count), nil)); !res.Ok() {
		return nil, res.asError("vkGetPhysicalDeviceSurfacePresentModesKHR(count)")
	}
	if count == 0 {
		return nil, nil
	}
	modes := make([]PresentMode, count)
	res := Result(vulkan.VkGetPhysicalDeviceSurfacePresentModesKHR(vulkan.VkPhysicalDevice(pd), vulkan.VkSurfaceKHR(s), unsafe.Pointer(&count), unsafe.Pointer(&modes[0])))
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
