package main

import (
	"github.com/christerso/vulkan-go/examples/internal/win"
	"github.com/christerso/vulkan-go/vk"
)

// swapchain bundles the swapchain and its per-image resources.
type swapchain struct {
	handle         vk.SwapchainKHR
	extent         vk.Extent2D
	images         []vk.Image
	views          []vk.ImageView
	depth          vk.AllocImage
	depthView      vk.ImageView
	framebuffers   []vk.Framebuffer
	renderFinished []vk.Semaphore
}

func newSwapchain(device vk.Device, pd vk.PhysicalDevice, surf vk.SurfaceKHR, rp vk.RenderPass,
	format vk.Format, colorSpace uint32, present vk.PresentMode, depthFormat vk.Format, window *win.Window) (*swapchain, error) {

	caps, err := pd.SurfaceCapabilities(surf)
	if err != nil {
		return nil, err
	}
	extent := caps.CurrentExtent
	if extent.Width == 0xFFFFFFFF {
		w, h := window.PixelSize()
		extent = vk.Extent2D{Width: clamp(w, caps.MinImageExtent.Width, caps.MaxImageExtent.Width),
			Height: clamp(h, caps.MinImageExtent.Height, caps.MaxImageExtent.Height)}
	}
	minCount := caps.MinImageCount + 1
	if caps.MaxImageCount > 0 && minCount > caps.MaxImageCount {
		minCount = caps.MaxImageCount
	}

	sc := &swapchain{extent: extent}
	sc.handle, err = device.CreateSwapchain(vk.SwapchainConfig{
		Surface:       surf,
		MinImageCount: minCount,
		Format:        format,
		ColorSpace:    colorSpace,
		Extent:        extent,
		PresentMode:   present,
		PreTransform:  caps.CurrentTransform,
	})
	if err != nil {
		return nil, err
	}

	sc.images, err = device.SwapchainImages(sc.handle)
	if err != nil {
		sc.destroy(device)
		return nil, err
	}
	for _, img := range sc.images {
		view, err := device.CreateImageView(img, format, vk.AspectColor)
		if err != nil {
			sc.destroy(device)
			return nil, err
		}
		sc.views = append(sc.views, view)
	}

	sc.depth, err = device.CreateImage2D(pd, depthFormat, extent, vk.ImageUsageDepthStencilAttachment)
	if err != nil {
		sc.destroy(device)
		return nil, err
	}
	sc.depthView, err = device.CreateImageView(sc.depth.Image, depthFormat, vk.AspectDepth)
	if err != nil {
		sc.destroy(device)
		return nil, err
	}

	for _, view := range sc.views {
		fb, err := device.CreateFramebuffer(rp, []vk.ImageView{view, sc.depthView}, extent)
		if err != nil {
			sc.destroy(device)
			return nil, err
		}
		sc.framebuffers = append(sc.framebuffers, fb)
		sem, err := device.CreateSemaphore()
		if err != nil {
			sc.destroy(device)
			return nil, err
		}
		sc.renderFinished = append(sc.renderFinished, sem)
	}
	return sc, nil
}

func (sc *swapchain) destroy(device vk.Device) {
	for _, sem := range sc.renderFinished {
		device.DestroySemaphore(sem)
	}
	for _, fb := range sc.framebuffers {
		device.DestroyFramebuffer(fb)
	}
	device.DestroyImageView(sc.depthView)
	device.DestroyImage(sc.depth)
	for _, v := range sc.views {
		device.DestroyImageView(v)
	}
	device.DestroySwapchain(sc.handle)
	*sc = swapchain{}
}

func clamp(v, lo, hi uint32) uint32 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
