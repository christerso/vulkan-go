package vk

import (
	"fmt"
	"runtime"
	"unsafe"

	vulkan "github.com/christerso/vulkan-go/vulkan"
)

// MemoryType mirrors VkMemoryType.
type MemoryType struct {
	PropertyFlags uint32
	HeapIndex     uint32
}

// MemoryHeap mirrors VkMemoryHeap (padded to 16 bytes by the uint64 size).
type MemoryHeap struct {
	Size  DeviceSize
	Flags uint32
	_     uint32
}

// MemoryRequirements mirrors VkMemoryRequirements.
type MemoryRequirements struct {
	Size           DeviceSize
	Alignment      DeviceSize
	MemoryTypeBits uint32
}

// memoryTypeIndex finds a memory type supporting typeBits with the given
// property flags.
func (pd PhysicalDevice) memoryTypeIndex(typeBits, props uint32) (uint32, error) {
	var mp vulkan.VkPhysicalDeviceMemoryProperties
	vulkan.VkGetPhysicalDeviceMemoryProperties(vulkan.VkPhysicalDevice(pd), unsafe.Pointer(&mp))
	for i := uint32(0); i < mp.MemoryTypeCount; i++ {
		if typeBits&(1<<i) != 0 && mp.MemoryTypes[i].PropertyFlags&props == props {
			return i, nil
		}
	}
	return 0, fmt.Errorf("vk: no memory type for bits %#x props %#x", typeBits, props)
}

// allocate allocates device memory of the given requirements and properties.
func (d Device) allocate(pd PhysicalDevice, req MemoryRequirements, props uint32) (DeviceMemory, error) {
	idx, err := pd.memoryTypeIndex(req.MemoryTypeBits, props)
	if err != nil {
		return 0, err
	}
	ai := vulkan.VkMemoryAllocateInfo{
		SType:           vulkan.VkStructureType(stMemoryAllocateInfo),
		AllocationSize:  vulkan.VkDeviceSize(req.Size),
		MemoryTypeIndex: idx,
	}
	var mem vulkan.VkDeviceMemory
	res := Result(vulkan.VkAllocateMemory(vulkan.VkDevice(d), unsafe.Pointer(&ai), nil, unsafe.Pointer(&mem)))
	runtime.KeepAlive(&ai)
	return DeviceMemory(mem), res.asError("vkAllocateMemory")
}

// FreeMemory frees device memory.
func (d Device) FreeMemory(mem DeviceMemory) {
	if mem != 0 {
		vulkan.VkFreeMemory(vulkan.VkDevice(d), vulkan.VkDeviceMemory(mem), nil)
	}
}

// Map maps device memory and returns a pointer to the start.
func (d Device) Map(mem DeviceMemory, size DeviceSize) (unsafe.Pointer, error) {
	var p unsafe.Pointer
	res := Result(vulkan.VkMapMemory(vulkan.VkDevice(d), vulkan.VkDeviceMemory(mem), 0, vulkan.VkDeviceSize(size), 0, unsafe.Pointer(&p)))
	return p, res.asError("vkMapMemory")
}

// Unmap unmaps device memory.
func (d Device) Unmap(mem DeviceMemory) { vulkan.VkUnmapMemory(vulkan.VkDevice(d), vulkan.VkDeviceMemory(mem)) }

// AllocBuffer bundles a buffer handle, its memory, and size.
type AllocBuffer struct {
	Buffer Buffer
	Memory DeviceMemory
	Size   DeviceSize
	Mapped unsafe.Pointer // non-nil for host-visible buffers created with Map=true
}

// BufferConfig describes a buffer allocation.
type BufferConfig struct {
	Size       DeviceSize
	Usage      uint32
	Properties uint32 // memory property flags
	Map        bool   // keep host-visible memory persistently mapped
}

// CreateBuffer creates a buffer, allocates memory for it, and binds them.
func (d Device) CreateBuffer(pd PhysicalDevice, cfg BufferConfig) (AllocBuffer, error) {
	ci := vulkan.VkBufferCreateInfo{
		SType:       vulkan.VkStructureType(stBufferCreateInfo),
		Size:        vulkan.VkDeviceSize(cfg.Size),
		Usage:       cfg.Usage,
		SharingMode: vulkan.VkSharingMode(SharingModeExclusive),
	}
	var buf vulkan.VkBuffer
	res := Result(vulkan.VkCreateBuffer(vulkan.VkDevice(d), unsafe.Pointer(&ci), nil, unsafe.Pointer(&buf)))
	runtime.KeepAlive(&ci)
	if err := res.asError("vkCreateBuffer"); err != nil {
		return AllocBuffer{}, err
	}
	var req MemoryRequirements
	vulkan.VkGetBufferMemoryRequirements(vulkan.VkDevice(d), buf, unsafe.Pointer(&req))
	mem, err := d.allocate(pd, req, cfg.Properties)
	if err != nil {
		vulkan.VkDestroyBuffer(vulkan.VkDevice(d), buf, nil)
		return AllocBuffer{}, err
	}
	if res := Result(vulkan.VkBindBufferMemory(vulkan.VkDevice(d), buf, vulkan.VkDeviceMemory(mem), 0)); !res.Ok() {
		vulkan.VkDestroyBuffer(vulkan.VkDevice(d), buf, nil)
		vulkan.VkFreeMemory(vulkan.VkDevice(d), vulkan.VkDeviceMemory(mem), nil)
		return AllocBuffer{}, res.asError("vkBindBufferMemory")
	}
	ab := AllocBuffer{Buffer: Buffer(buf), Memory: mem, Size: cfg.Size}
	if cfg.Map {
		p, err := d.Map(mem, cfg.Size)
		if err != nil {
			d.DestroyBuffer(ab)
			return AllocBuffer{}, err
		}
		ab.Mapped = p
	}
	return ab, nil
}

// DestroyBuffer frees a buffer and its memory.
func (d Device) DestroyBuffer(b AllocBuffer) {
	if b.Buffer != 0 {
		vulkan.VkDestroyBuffer(vulkan.VkDevice(d), vulkan.VkBuffer(b.Buffer), nil)
	}
	if b.Memory != 0 {
		vulkan.VkFreeMemory(vulkan.VkDevice(d), vulkan.VkDeviceMemory(b.Memory), nil)
	}
}

// AllocImage bundles an image with its memory.
type AllocImage struct {
	Image  Image
	Memory DeviceMemory
}

// CreateImage2D creates a 2D image and binds device-local memory.
func (d Device) CreateImage2D(pd PhysicalDevice, format Format, extent Extent2D, usage uint32) (AllocImage, error) {
	ci := vulkan.VkImageCreateInfo{
		SType:         vulkan.VkStructureType(stImageCreateInfo),
		ImageType:     vulkan.VkImageType(ImageType2D),
		Format:        vulkan.VkFormat(format),
		Extent:        vulkan.VkExtent3D{Width: extent.Width, Height: extent.Height, Depth: 1},
		MipLevels:     1,
		ArrayLayers:   1,
		Samples:       SampleCount1,
		Tiling:        vulkan.VkImageTiling(ImageTilingOptimal),
		Usage:         usage,
		SharingMode:   vulkan.VkSharingMode(SharingModeExclusive),
		InitialLayout: vulkan.VkImageLayout(LayoutUndefined),
	}
	var img vulkan.VkImage
	res := Result(vulkan.VkCreateImage(vulkan.VkDevice(d), unsafe.Pointer(&ci), nil, unsafe.Pointer(&img)))
	runtime.KeepAlive(&ci)
	if err := res.asError("vkCreateImage"); err != nil {
		return AllocImage{}, err
	}
	var req MemoryRequirements
	vulkan.VkGetImageMemoryRequirements(vulkan.VkDevice(d), img, unsafe.Pointer(&req))
	mem, err := d.allocate(pd, req, MemoryDeviceLocal)
	if err != nil {
		vulkan.VkDestroyImage(vulkan.VkDevice(d), img, nil)
		return AllocImage{}, err
	}
	if res := Result(vulkan.VkBindImageMemory(vulkan.VkDevice(d), img, vulkan.VkDeviceMemory(mem), 0)); !res.Ok() {
		vulkan.VkDestroyImage(vulkan.VkDevice(d), img, nil)
		vulkan.VkFreeMemory(vulkan.VkDevice(d), vulkan.VkDeviceMemory(mem), nil)
		return AllocImage{}, res.asError("vkBindImageMemory")
	}
	return AllocImage{Image: Image(img), Memory: mem}, nil
}

// DestroyImage frees an image and its memory.
func (d Device) DestroyImage(a AllocImage) {
	if a.Image != 0 {
		vulkan.VkDestroyImage(vulkan.VkDevice(d), vulkan.VkImage(a.Image), nil)
	}
	if a.Memory != 0 {
		vulkan.VkFreeMemory(vulkan.VkDevice(d), vulkan.VkDeviceMemory(a.Memory), nil)
	}
}

// CreateImageView creates a 2D image view over the given aspect.
func (d Device) CreateImageView(img Image, format Format, aspect uint32) (ImageView, error) {
	ci := vulkan.VkImageViewCreateInfo{
		SType:    vulkan.VkStructureType(stImageViewCreateInfo),
		Image:    vulkan.VkImage(img),
		ViewType: vulkan.VkImageViewType(ImageViewType2D),
		Format:   vulkan.VkFormat(format),
		SubresourceRange: vulkan.VkImageSubresourceRange{
			AspectMask: aspect,
			LevelCount: 1,
			LayerCount: 1,
		},
	}
	var view vulkan.VkImageView
	res := Result(vulkan.VkCreateImageView(vulkan.VkDevice(d), unsafe.Pointer(&ci), nil, unsafe.Pointer(&view)))
	runtime.KeepAlive(&ci)
	return ImageView(view), res.asError("vkCreateImageView")
}

// DestroyImageView destroys an image view.
func (d Device) DestroyImageView(v ImageView) {
	if v != 0 {
		vulkan.VkDestroyImageView(vulkan.VkDevice(d), vulkan.VkImageView(v), nil)
	}
}

// CopyToMapped copies src into a mapped pointer.
func CopyToMapped(dst unsafe.Pointer, src []byte) {
	copy(unsafe.Slice((*byte)(dst), len(src)), src)
}

// CreateDeviceLocalBuffer uploads data into a new device-local buffer through a
// host-visible staging buffer, using a one-time command submission. usage is
// the device-local buffer usage; transfer-dst is added automatically.
func (d Device) CreateDeviceLocalBuffer(pd PhysicalDevice, q Queue, pool CommandPool, data []byte, usage uint32) (AllocBuffer, error) {
	size := DeviceSize(len(data))
	staging, err := d.CreateBuffer(pd, BufferConfig{
		Size:       size,
		Usage:      BufferUsageTransferSrc,
		Properties: MemoryHostVisible | MemoryHostCoherent,
		Map:        true,
	})
	if err != nil {
		return AllocBuffer{}, err
	}
	defer d.DestroyBuffer(staging)
	CopyToMapped(staging.Mapped, data)

	dst, err := d.CreateBuffer(pd, BufferConfig{
		Size:       size,
		Usage:      usage | BufferUsageTransferDst,
		Properties: MemoryDeviceLocal,
	})
	if err != nil {
		return AllocBuffer{}, err
	}

	cmds, err := d.AllocateCommandBuffers(pool, 1)
	if err != nil {
		d.DestroyBuffer(dst)
		return AllocBuffer{}, err
	}
	cmd := cmds[0]
	if err := cmd.Begin(CommandBufferOneTimeSubmit); err != nil {
		d.DestroyBuffer(dst)
		return AllocBuffer{}, err
	}
	cmd.CopyBuffer(staging.Buffer, dst.Buffer, size)
	if err := cmd.End(); err != nil {
		d.DestroyBuffer(dst)
		return AllocBuffer{}, err
	}
	if err := q.Submit(SubmitConfig{Command: cmd}); err != nil {
		d.DestroyBuffer(dst)
		return AllocBuffer{}, err
	}
	if err := q.WaitIdle(); err != nil {
		d.DestroyBuffer(dst)
		return AllocBuffer{}, err
	}
	return dst, nil
}
