package vk

import (
	"runtime"
	"unsafe"
)

type swapchainCreateInfo struct {
	sType                 uint32
	pNext                 unsafe.Pointer
	flags                 uint32
	surface               SurfaceKHR
	minImageCount         uint32
	imageFormat           Format
	imageColorSpace       uint32
	imageExtent           Extent2D
	imageArrayLayers      uint32
	imageUsage            uint32
	imageSharingMode      uint32
	queueFamilyIndexCount uint32
	pQueueFamilyIndices   *uint32
	preTransform          uint32
	compositeAlpha        uint32
	presentMode           PresentMode
	clipped               uint32
	oldSwapchain          SwapchainKHR
}

type presentInfo struct {
	sType              uint32
	pNext              unsafe.Pointer
	waitSemaphoreCount uint32
	pWaitSemaphores    *Semaphore
	swapchainCount     uint32
	pSwapchains        *SwapchainKHR
	pImageIndices      *uint32
	pResults           *Result
}

var (
	vkCreateSwapchainKHR    func(device Device, pCreateInfo, pAllocator unsafe.Pointer, pSwapchain *SwapchainKHR) Result
	vkDestroySwapchainKHR   func(device Device, swapchain SwapchainKHR, pAllocator unsafe.Pointer)
	vkGetSwapchainImagesKHR func(device Device, swapchain SwapchainKHR, pCount *uint32, pImages *Image) Result
	vkAcquireNextImageKHR   func(device Device, swapchain SwapchainKHR, timeout uint64, semaphore Semaphore, fence Fence, pIndex *uint32) Result
	vkQueuePresentKHR       func(queue Queue, pPresentInfo unsafe.Pointer) Result
)

func loadSwapchainCommands(device Device) {
	h := uintptr(device)
	bindDeviceProc(&vkCreateSwapchainKHR, h, "vkCreateSwapchainKHR")
	bindDeviceProc(&vkDestroySwapchainKHR, h, "vkDestroySwapchainKHR")
	bindDeviceProc(&vkGetSwapchainImagesKHR, h, "vkGetSwapchainImagesKHR")
	bindDeviceProc(&vkAcquireNextImageKHR, h, "vkAcquireNextImageKHR")
	bindDeviceProc(&vkQueuePresentKHR, h, "vkQueuePresentKHR")
}

// SwapchainConfig describes swapchain creation.
type SwapchainConfig struct {
	Surface      SurfaceKHR
	MinImageCount uint32
	Format       Format
	ColorSpace   uint32
	Extent       Extent2D
	PresentMode  PresentMode
	PreTransform uint32
	Old          SwapchainKHR
}

// CreateSwapchain creates a swapchain for color attachment output.
func (d Device) CreateSwapchain(cfg SwapchainConfig) (SwapchainKHR, error) {
	ci := swapchainCreateInfo{
		sType:            stSwapchainCreateInfoKHR,
		surface:          cfg.Surface,
		minImageCount:    cfg.MinImageCount,
		imageFormat:      cfg.Format,
		imageColorSpace:  cfg.ColorSpace,
		imageExtent:      cfg.Extent,
		imageArrayLayers: 1,
		imageUsage:       ImageUsageColorAttachment,
		imageSharingMode: SharingModeExclusive,
		preTransform:     cfg.PreTransform,
		compositeAlpha:   CompositeAlphaOpaque,
		presentMode:      cfg.PresentMode,
		clipped:          1,
		oldSwapchain:     cfg.Old,
	}
	var sc SwapchainKHR
	res := vkCreateSwapchainKHR(d, unsafe.Pointer(&ci), nil, &sc)
	runtime.KeepAlive(&ci)
	return sc, res.asError("vkCreateSwapchainKHR")
}

// DestroySwapchain destroys a swapchain.
func (d Device) DestroySwapchain(sc SwapchainKHR) {
	if sc != 0 {
		vkDestroySwapchainKHR(d, sc, nil)
	}
}

// SwapchainImages returns the images owned by the swapchain.
func (d Device) SwapchainImages(sc SwapchainKHR) ([]Image, error) {
	var count uint32
	if res := vkGetSwapchainImagesKHR(d, sc, &count, nil); !res.Ok() {
		return nil, res.asError("vkGetSwapchainImagesKHR(count)")
	}
	images := make([]Image, count)
	res := vkGetSwapchainImagesKHR(d, sc, &count, &images[0])
	return images, res.asError("vkGetSwapchainImagesKHR(list)")
}

// AcquireNextImage acquires the next swapchain image, signaling sem when ready.
// It returns the image index and the raw result (which may be SuboptimalKHR or
// ErrorOutOfDateKHR and still needs handling by the caller).
func (d Device) AcquireNextImage(sc SwapchainKHR, sem Semaphore, timeout uint64) (uint32, Result) {
	var index uint32
	res := vkAcquireNextImageKHR(d, sc, timeout, sem, 0, &index)
	return index, res
}

// Present queues the image for presentation, waiting on sem.
func (q Queue) Present(sc SwapchainKHR, imageIndex uint32, wait Semaphore) Result {
	scs := sc
	idx := imageIndex
	pi := presentInfo{
		sType:          stPresentInfoKHR,
		swapchainCount: 1,
		pSwapchains:    &scs,
		pImageIndices:  &idx,
	}
	if wait != 0 {
		w := wait
		pi.waitSemaphoreCount = 1
		pi.pWaitSemaphores = &w
		res := vkQueuePresentKHR(q, unsafe.Pointer(&pi))
		runtime.KeepAlive(&w)
		runtime.KeepAlive(&pi)
		return res
	}
	res := vkQueuePresentKHR(q, unsafe.Pointer(&pi))
	runtime.KeepAlive(&pi)
	return res
}
