package vk

import (
	"runtime"
	"unsafe"

	vulkan "github.com/christerso/vulkan-go/vulkan"
)

// Filter values (VkFilter).
const (
	FilterNearest uint32 = 0
	FilterLinear  uint32 = 1
)

// Sampler mipmap mode (VkSamplerMipmapMode).
const (
	SamplerMipmapModeNearest uint32 = 0
	SamplerMipmapModeLinear  uint32 = 1
)

// Sampler address mode (VkSamplerAddressMode).
const (
	SamplerAddressModeRepeat         uint32 = 0
	SamplerAddressModeMirroredRepeat uint32 = 1
	SamplerAddressModeClampToEdge    uint32 = 2
	SamplerAddressModeClampToBorder  uint32 = 3
)

// Border color (VkBorderColor).
const (
	BorderColorFloatOpaqueBlack uint32 = 1
	BorderColorIntOpaqueBlack   uint32 = 3
)

// Image usage additions used by texture creation are already declared in
// enums.go (ImageUsageTransferDst, ImageUsageSampled).

// SamplerConfig describes a sampler. The zero value yields a linear-filtered,
// repeat-addressed sampler, which is what CreateSampler defaults to when fields
// are left zero (FilterNearest/RepeatAddress are themselves zero-valued, so
// CreateSampler applies sensible non-zero defaults explicitly).
type SamplerConfig struct {
	MagFilter    uint32 // FilterLinear (default) / FilterNearest
	MinFilter    uint32 // FilterLinear (default) / FilterNearest
	MipmapMode   uint32 // SamplerMipmapModeLinear (default)
	AddressModeU uint32 // SamplerAddressModeRepeat (default)
	AddressModeV uint32 // SamplerAddressModeRepeat (default)
	AddressModeW uint32 // SamplerAddressModeRepeat (default)
	MaxLod       float32
}

// CreateSampler creates a sampler. An empty SamplerConfig{} gives linear
// filtering with repeat addressing, which is the common case for textures.
func (d Device) CreateSampler(cfg SamplerConfig) (Sampler, error) {
	// Defaults: linear filtering, repeat addressing. Callers can override any
	// field; since FilterLinear/RepeatRepeat are non-zero/zero respectively, we
	// translate the zero value of the whole struct to the linear+repeat default.
	c := cfg
	if c == (SamplerConfig{}) {
		c = SamplerConfig{
			MagFilter:    FilterLinear,
			MinFilter:    FilterLinear,
			MipmapMode:   SamplerMipmapModeLinear,
			AddressModeU: SamplerAddressModeRepeat,
			AddressModeV: SamplerAddressModeRepeat,
			AddressModeW: SamplerAddressModeRepeat,
		}
	}
	ci := vulkan.VkSamplerCreateInfo{
		SType:        vulkan.VkStructureType(stSamplerCreateInfo),
		MagFilter:    vulkan.VkFilter(c.MagFilter),
		MinFilter:    vulkan.VkFilter(c.MinFilter),
		MipmapMode:   vulkan.VkSamplerMipmapMode(c.MipmapMode),
		AddressModeU: vulkan.VkSamplerAddressMode(c.AddressModeU),
		AddressModeV: vulkan.VkSamplerAddressMode(c.AddressModeV),
		AddressModeW: vulkan.VkSamplerAddressMode(c.AddressModeW),
		MaxLod:       c.MaxLod,
		BorderColor:  vulkan.VkBorderColor(BorderColorIntOpaqueBlack),
		CompareOp:    vulkan.VkCompareOp(CompareNever),
	}
	var s vulkan.VkSampler
	res := Result(vulkan.VkCreateSampler(vulkan.VkDevice(d), unsafe.Pointer(&ci), nil, unsafe.Pointer(&s)))
	runtime.KeepAlive(&ci)
	return Sampler(s), res.asError("vkCreateSampler")
}

// DestroySampler destroys a sampler.
func (d Device) DestroySampler(s Sampler) {
	if s != 0 {
		vulkan.VkDestroySampler(vulkan.VkDevice(d), vulkan.VkSampler(s), nil)
	}
}

// ImageBarrier records a vkCmdPipelineBarrier with a single image memory
// barrier transitioning img from oldLayout to newLayout. The stage and access
// masks and the image aspect are supplied by the caller, matching the Vulkan
// VkImageMemoryBarrier / vkCmdPipelineBarrier parameters.
func (c CommandBuffer) ImageBarrier(img Image, oldLayout, newLayout ImageLayout, srcStage, dstStage, srcAccess, dstAccess uint32, aspect uint32) {
	const queueFamilyIgnored uint32 = 0xFFFFFFFF
	bar := vulkan.VkImageMemoryBarrier{
		SType:               vulkan.VkStructureType(stImageMemoryBarrier),
		SrcAccessMask:       srcAccess,
		DstAccessMask:       dstAccess,
		OldLayout:           vulkan.VkImageLayout(oldLayout),
		NewLayout:           vulkan.VkImageLayout(newLayout),
		SrcQueueFamilyIndex: queueFamilyIgnored,
		DstQueueFamilyIndex: queueFamilyIgnored,
		Image:               vulkan.VkImage(img),
		SubresourceRange: vulkan.VkImageSubresourceRange{
			AspectMask: aspect,
			LevelCount: 1,
			LayerCount: 1,
		},
	}
	vulkan.VkCmdPipelineBarrier(vulkan.VkCommandBuffer(c), srcStage, dstStage, 0, 0, nil, 0, nil, 1, unsafe.Pointer(&bar))
	runtime.KeepAlive(&bar)
}

// CopyBufferToImage records a copy of the whole buffer into the image's
// transfer-dst layout at the given extent (one mip, one layer, color aspect).
func (c CommandBuffer) CopyBufferToImage(buf Buffer, img Image, width, height uint32) {
	region := vulkan.VkBufferImageCopy{
		ImageSubresource: vulkan.VkImageSubresourceLayers{
			AspectMask: AspectColor,
			LayerCount: 1,
		},
		ImageExtent: vulkan.VkExtent3D{Width: width, Height: height, Depth: 1},
	}
	vulkan.VkCmdCopyBufferToImage(vulkan.VkCommandBuffer(c), vulkan.VkBuffer(buf), vulkan.VkImage(img), vulkan.VkImageLayout(LayoutTransferDstOptimal), 1, unsafe.Pointer(&region))
	runtime.KeepAlive(&region)
}

// CreateTexture2D creates a device-local, sampled 2D image from RGBA bytes and
// returns the image (with its memory) and a color image view. The data is
// uploaded through a host-visible staging buffer and a one-time command buffer
// that transitions UNDEFINED -> TRANSFER_DST, copies, then TRANSFER_DST ->
// SHADER_READ_ONLY. rgba must hold width*height*4 bytes (R8G8B8A8_UNORM).
func (d Device) CreateTexture2D(pd PhysicalDevice, q Queue, pool CommandPool, width, height uint32, rgba []byte) (AllocImage, ImageView, error) {
	img, err := d.CreateImage2D(pd, FormatR8G8B8A8Unorm, Extent2D{Width: width, Height: height}, ImageUsageTransferDst|ImageUsageSampled)
	if err != nil {
		return AllocImage{}, 0, err
	}

	staging, err := d.CreateBuffer(pd, BufferConfig{
		Size:       DeviceSize(len(rgba)),
		Usage:      BufferUsageTransferSrc,
		Properties: MemoryHostVisible | MemoryHostCoherent,
		Map:        true,
	})
	if err != nil {
		d.DestroyImage(img)
		return AllocImage{}, 0, err
	}
	defer d.DestroyBuffer(staging)
	CopyToMapped(staging.Mapped, rgba)

	cmds, err := d.AllocateCommandBuffers(pool, 1)
	if err != nil {
		d.DestroyImage(img)
		return AllocImage{}, 0, err
	}
	cmd := cmds[0]
	if err := cmd.Begin(CommandBufferOneTimeSubmit); err != nil {
		d.DestroyImage(img)
		return AllocImage{}, 0, err
	}
	cmd.ImageBarrier(img.Image, LayoutUndefined, LayoutTransferDstOptimal,
		StageTopOfPipe, StageTransfer, 0, AccessTransferWrite, AspectColor)
	cmd.CopyBufferToImage(staging.Buffer, img.Image, width, height)
	cmd.ImageBarrier(img.Image, LayoutTransferDstOptimal, LayoutShaderReadOnlyOptimal,
		StageTransfer, StageFragmentShader, AccessTransferWrite, AccessShaderRead, AspectColor)
	if err := cmd.End(); err != nil {
		d.DestroyImage(img)
		return AllocImage{}, 0, err
	}
	if err := q.Submit(SubmitConfig{Command: cmd}); err != nil {
		d.DestroyImage(img)
		return AllocImage{}, 0, err
	}
	if err := q.WaitIdle(); err != nil {
		d.DestroyImage(img)
		return AllocImage{}, 0, err
	}

	view, err := d.CreateImageView(img.Image, FormatR8G8B8A8Unorm, AspectColor)
	if err != nil {
		d.DestroyImage(img)
		return AllocImage{}, 0, err
	}
	return img, view, nil
}

// UpdateImageDescriptor points a combined-image-sampler descriptor at the given
// image view and sampler, assuming the image is in SHADER_READ_ONLY layout.
func (d Device) UpdateImageDescriptor(set DescriptorSet, binding uint32, view ImageView, sampler Sampler) {
	ii := vulkan.VkDescriptorImageInfo{
		Sampler:     vulkan.VkSampler(sampler),
		ImageView:   vulkan.VkImageView(view),
		ImageLayout: vulkan.VkImageLayout(LayoutShaderReadOnlyOptimal),
	}
	w := vulkan.VkWriteDescriptorSet{
		SType:           vulkan.VkStructureType(stWriteDescriptorSet),
		DstSet:          vulkan.VkDescriptorSet(set),
		DstBinding:      binding,
		DescriptorCount: 1,
		DescriptorType:  vulkan.VkDescriptorType(DescriptorCombinedImageSampler),
		PImageInfo:      unsafe.Pointer(&ii),
	}
	vulkan.VkUpdateDescriptorSets(vulkan.VkDevice(d), 1, unsafe.Pointer(&w), 0, nil)
	runtime.KeepAlive(&ii)
	runtime.KeepAlive(&w)
}
