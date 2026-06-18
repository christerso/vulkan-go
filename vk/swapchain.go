package vk

import (
	"runtime"
	"unsafe"

	vulkan "github.com/christerso/vulkan-go/vulkan"
)

// SwapchainConfig describes swapchain creation.
type SwapchainConfig struct {
	Surface       SurfaceKHR
	MinImageCount uint32
	Format        Format
	ColorSpace    uint32
	Extent        Extent2D
	PresentMode   PresentMode
	PreTransform  uint32
	Old           SwapchainKHR
}

// CreateSwapchain creates a swapchain for color attachment output.
func (d Device) CreateSwapchain(cfg SwapchainConfig) (SwapchainKHR, error) {
	ci := vulkan.VkSwapchainCreateInfoKHR{
		SType:            vulkan.VkStructureType(stSwapchainCreateInfoKHR),
		Surface:          vulkan.VkSurfaceKHR(cfg.Surface),
		MinImageCount:    cfg.MinImageCount,
		ImageFormat:      vulkan.VkFormat(cfg.Format),
		ImageColorSpace:  vulkan.VkColorSpaceKHR(cfg.ColorSpace),
		ImageExtent:      vulkan.VkExtent2D{Width: cfg.Extent.Width, Height: cfg.Extent.Height},
		ImageArrayLayers: 1,
		ImageUsage:       ImageUsageColorAttachment,
		ImageSharingMode: vulkan.VkSharingMode(SharingModeExclusive),
		PreTransform:     vulkan.VkSurfaceTransformFlagBitsKHR(cfg.PreTransform),
		CompositeAlpha:   vulkan.VkCompositeAlphaFlagBitsKHR(CompositeAlphaOpaque),
		PresentMode:      vulkan.VkPresentModeKHR(cfg.PresentMode),
		Clipped:          1,
		OldSwapchain:     vulkan.VkSwapchainKHR(cfg.Old),
	}
	var sc vulkan.VkSwapchainKHR
	res := Result(vulkan.VkCreateSwapchainKHR(vulkan.VkDevice(d), unsafe.Pointer(&ci), nil, unsafe.Pointer(&sc)))
	runtime.KeepAlive(&ci)
	return SwapchainKHR(sc), res.asError("vkCreateSwapchainKHR")
}

// DestroySwapchain destroys a swapchain.
func (d Device) DestroySwapchain(sc SwapchainKHR) {
	if sc != 0 {
		vulkan.VkDestroySwapchainKHR(vulkan.VkDevice(d), vulkan.VkSwapchainKHR(sc), nil)
	}
}

// SwapchainImages returns the images owned by the swapchain.
func (d Device) SwapchainImages(sc SwapchainKHR) ([]Image, error) {
	var count uint32
	if res := Result(vulkan.VkGetSwapchainImagesKHR(vulkan.VkDevice(d), vulkan.VkSwapchainKHR(sc), unsafe.Pointer(&count), nil)); !res.Ok() {
		return nil, res.asError("vkGetSwapchainImagesKHR(count)")
	}
	images := make([]Image, count)
	res := Result(vulkan.VkGetSwapchainImagesKHR(vulkan.VkDevice(d), vulkan.VkSwapchainKHR(sc), unsafe.Pointer(&count), unsafe.Pointer(&images[0])))
	return images, res.asError("vkGetSwapchainImagesKHR(list)")
}

// AcquireNextImage acquires the next swapchain image, signaling sem when ready.
// It returns the image index and the raw result (which may be SuboptimalKHR or
// ErrorOutOfDateKHR and still needs handling by the caller).
func (d Device) AcquireNextImage(sc SwapchainKHR, sem Semaphore, timeout uint64) (uint32, Result) {
	var index uint32
	res := Result(vulkan.VkAcquireNextImageKHR(vulkan.VkDevice(d), vulkan.VkSwapchainKHR(sc), timeout, vulkan.VkSemaphore(sem), 0, unsafe.Pointer(&index)))
	return index, res
}

// Present queues the image for presentation, waiting on sem.
func (q Queue) Present(sc SwapchainKHR, imageIndex uint32, wait Semaphore) Result {
	scs := vulkan.VkSwapchainKHR(sc)
	idx := imageIndex
	pi := vulkan.VkPresentInfoKHR{
		SType:          vulkan.VkStructureType(stPresentInfoKHR),
		SwapchainCount: 1,
		PSwapchains:    unsafe.Pointer(&scs),
		PImageIndices:  unsafe.Pointer(&idx),
	}
	if wait != 0 {
		w := vulkan.VkSemaphore(wait)
		pi.WaitSemaphoreCount = 1
		pi.PWaitSemaphores = unsafe.Pointer(&w)
		res := Result(vulkan.VkQueuePresentKHR(vulkan.VkQueue(q), unsafe.Pointer(&pi)))
		runtime.KeepAlive(&w)
		runtime.KeepAlive(&pi)
		runtime.KeepAlive(&scs)
		runtime.KeepAlive(&idx)
		return res
	}
	res := Result(vulkan.VkQueuePresentKHR(vulkan.VkQueue(q), unsafe.Pointer(&pi)))
	runtime.KeepAlive(&pi)
	runtime.KeepAlive(&scs)
	runtime.KeepAlive(&idx)
	return res
}
