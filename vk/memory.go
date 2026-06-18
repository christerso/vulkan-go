package vk

import (
	"fmt"
	"runtime"
	"unsafe"
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

type physicalDeviceMemoryProperties struct {
	memoryTypeCount uint32
	memoryTypes     [32]MemoryType
	memoryHeapCount uint32
	memoryHeaps     [16]MemoryHeap
}

// MemoryRequirements mirrors VkMemoryRequirements.
type MemoryRequirements struct {
	Size           DeviceSize
	Alignment      DeviceSize
	MemoryTypeBits uint32
}

type memoryAllocateInfo struct {
	sType           uint32
	pNext           unsafe.Pointer
	allocationSize  DeviceSize
	memoryTypeIndex uint32
}

var (
	vkGetPhysicalDeviceMemoryProperties func(pd PhysicalDevice, pProps *physicalDeviceMemoryProperties)
	vkAllocateMemory                    func(device Device, pInfo, pAllocator unsafe.Pointer, pMem *DeviceMemory) Result
	vkFreeMemory                        func(device Device, mem DeviceMemory, pAllocator unsafe.Pointer)
	vkMapMemory                         func(device Device, mem DeviceMemory, offset, size DeviceSize, flags uint32, ppData *unsafe.Pointer) Result
	vkUnmapMemory                       func(device Device, mem DeviceMemory)
	vkGetBufferMemoryRequirements       func(device Device, buffer Buffer, pReq *MemoryRequirements)
	vkBindBufferMemory                  func(device Device, buffer Buffer, mem DeviceMemory, offset DeviceSize) Result
	vkGetImageMemoryRequirements        func(device Device, image Image, pReq *MemoryRequirements)
	vkBindImageMemory                   func(device Device, image Image, mem DeviceMemory, offset DeviceSize) Result
	vkCreateBuffer                      func(device Device, pInfo, pAllocator unsafe.Pointer, pBuffer *Buffer) Result
	vkDestroyBuffer                     func(device Device, buffer Buffer, pAllocator unsafe.Pointer)
	vkCreateImage                       func(device Device, pInfo, pAllocator unsafe.Pointer, pImage *Image) Result
	vkDestroyImage                      func(device Device, image Image, pAllocator unsafe.Pointer)
	vkCreateImageView                   func(device Device, pInfo, pAllocator unsafe.Pointer, pView *ImageView) Result
	vkDestroyImageView                  func(device Device, view ImageView, pAllocator unsafe.Pointer)
)

func loadMemoryCommands(device Device) {
	h := uintptr(device)
	bindDeviceProc(&vkAllocateMemory, h, "vkAllocateMemory")
	bindDeviceProc(&vkFreeMemory, h, "vkFreeMemory")
	bindDeviceProc(&vkMapMemory, h, "vkMapMemory")
	bindDeviceProc(&vkUnmapMemory, h, "vkUnmapMemory")
	bindDeviceProc(&vkGetBufferMemoryRequirements, h, "vkGetBufferMemoryRequirements")
	bindDeviceProc(&vkBindBufferMemory, h, "vkBindBufferMemory")
	bindDeviceProc(&vkGetImageMemoryRequirements, h, "vkGetImageMemoryRequirements")
	bindDeviceProc(&vkBindImageMemory, h, "vkBindImageMemory")
	bindDeviceProc(&vkCreateBuffer, h, "vkCreateBuffer")
	bindDeviceProc(&vkDestroyBuffer, h, "vkDestroyBuffer")
	bindDeviceProc(&vkCreateImage, h, "vkCreateImage")
	bindDeviceProc(&vkDestroyImage, h, "vkDestroyImage")
	bindDeviceProc(&vkCreateImageView, h, "vkCreateImageView")
	bindDeviceProc(&vkDestroyImageView, h, "vkDestroyImageView")
}

// memoryTypeIndex finds a memory type supporting typeBits with the given
// property flags.
func (pd PhysicalDevice) memoryTypeIndex(typeBits, props uint32) (uint32, error) {
	var mp physicalDeviceMemoryProperties
	vkGetPhysicalDeviceMemoryProperties(pd, &mp)
	for i := uint32(0); i < mp.memoryTypeCount; i++ {
		if typeBits&(1<<i) != 0 && mp.memoryTypes[i].PropertyFlags&props == props {
			return i, nil
		}
	}
	return 0, fmt.Errorf("vk: no memory type for bits %#x props %#x", typeBits, props)
}

// Allocation pairs a device memory object with the physical device used to pick
// its type, so callers can free it.
func (d Device) allocate(pd PhysicalDevice, req MemoryRequirements, props uint32) (DeviceMemory, error) {
	idx, err := pd.memoryTypeIndex(req.MemoryTypeBits, props)
	if err != nil {
		return 0, err
	}
	ai := memoryAllocateInfo{
		sType:           stMemoryAllocateInfo,
		allocationSize:  req.Size,
		memoryTypeIndex: idx,
	}
	var mem DeviceMemory
	res := vkAllocateMemory(d, unsafe.Pointer(&ai), nil, &mem)
	runtime.KeepAlive(&ai)
	return mem, res.asError("vkAllocateMemory")
}

// FreeMemory frees device memory.
func (d Device) FreeMemory(mem DeviceMemory) {
	if mem != 0 {
		vkFreeMemory(d, mem, nil)
	}
}

// Map maps device memory and returns a pointer to the start.
func (d Device) Map(mem DeviceMemory, size DeviceSize) (unsafe.Pointer, error) {
	var p unsafe.Pointer
	res := vkMapMemory(d, mem, 0, size, 0, &p)
	return p, res.asError("vkMapMemory")
}

// Unmap unmaps device memory.
func (d Device) Unmap(mem DeviceMemory) { vkUnmapMemory(d, mem) }

type bufferCreateInfo struct {
	sType                 uint32
	pNext                 unsafe.Pointer
	flags                 uint32
	size                  DeviceSize
	usage                 uint32
	sharingMode           uint32
	queueFamilyIndexCount uint32
	pQueueFamilyIndices   *uint32
}

// Buffer bundles a buffer handle, its memory, and size.
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
	ci := bufferCreateInfo{
		sType:       stBufferCreateInfo,
		size:        cfg.Size,
		usage:       cfg.Usage,
		sharingMode: SharingModeExclusive,
	}
	var buf Buffer
	res := vkCreateBuffer(d, unsafe.Pointer(&ci), nil, &buf)
	runtime.KeepAlive(&ci)
	if err := res.asError("vkCreateBuffer"); err != nil {
		return AllocBuffer{}, err
	}
	var req MemoryRequirements
	vkGetBufferMemoryRequirements(d, buf, &req)
	mem, err := d.allocate(pd, req, cfg.Properties)
	if err != nil {
		vkDestroyBuffer(d, buf, nil)
		return AllocBuffer{}, err
	}
	if res := vkBindBufferMemory(d, buf, mem, 0); !res.Ok() {
		vkDestroyBuffer(d, buf, nil)
		vkFreeMemory(d, mem, nil)
		return AllocBuffer{}, res.asError("vkBindBufferMemory")
	}
	ab := AllocBuffer{Buffer: buf, Memory: mem, Size: cfg.Size}
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
		vkDestroyBuffer(d, b.Buffer, nil)
	}
	if b.Memory != 0 {
		vkFreeMemory(d, b.Memory, nil)
	}
}

type imageCreateInfo struct {
	sType                 uint32
	pNext                 unsafe.Pointer
	flags                 uint32
	imageType             uint32
	format                Format
	extent                Extent3D
	mipLevels             uint32
	arrayLayers           uint32
	samples               uint32
	tiling                uint32
	usage                 uint32
	sharingMode           uint32
	queueFamilyIndexCount uint32
	pQueueFamilyIndices   *uint32
	initialLayout         ImageLayout
}

// AllocImage bundles an image with its memory.
type AllocImage struct {
	Image  Image
	Memory DeviceMemory
}

// CreateImage2D creates a 2D image and binds device-local memory.
func (d Device) CreateImage2D(pd PhysicalDevice, format Format, extent Extent2D, usage uint32) (AllocImage, error) {
	ci := imageCreateInfo{
		sType:         stImageCreateInfo,
		imageType:     ImageType2D,
		format:        format,
		extent:        Extent3D{Width: extent.Width, Height: extent.Height, Depth: 1},
		mipLevels:     1,
		arrayLayers:   1,
		samples:       SampleCount1,
		tiling:        ImageTilingOptimal,
		usage:         usage,
		sharingMode:   SharingModeExclusive,
		initialLayout: LayoutUndefined,
	}
	var img Image
	res := vkCreateImage(d, unsafe.Pointer(&ci), nil, &img)
	runtime.KeepAlive(&ci)
	if err := res.asError("vkCreateImage"); err != nil {
		return AllocImage{}, err
	}
	var req MemoryRequirements
	vkGetImageMemoryRequirements(d, img, &req)
	mem, err := d.allocate(pd, req, MemoryDeviceLocal)
	if err != nil {
		vkDestroyImage(d, img, nil)
		return AllocImage{}, err
	}
	if res := vkBindImageMemory(d, img, mem, 0); !res.Ok() {
		vkDestroyImage(d, img, nil)
		vkFreeMemory(d, mem, nil)
		return AllocImage{}, res.asError("vkBindImageMemory")
	}
	return AllocImage{Image: img, Memory: mem}, nil
}

// DestroyImage frees an image and its memory.
func (d Device) DestroyImage(a AllocImage) {
	if a.Image != 0 {
		vkDestroyImage(d, a.Image, nil)
	}
	if a.Memory != 0 {
		vkFreeMemory(d, a.Memory, nil)
	}
}

type imageSubresourceRange struct {
	aspectMask     uint32
	baseMipLevel   uint32
	levelCount     uint32
	baseArrayLayer uint32
	layerCount     uint32
}

type componentMapping struct {
	r, g, b, a uint32
}

type imageViewCreateInfo struct {
	sType            uint32
	pNext            unsafe.Pointer
	flags            uint32
	image            Image
	viewType         uint32
	format           Format
	components       componentMapping
	subresourceRange imageSubresourceRange
}

// CreateImageView creates a 2D image view over the given aspect.
func (d Device) CreateImageView(img Image, format Format, aspect uint32) (ImageView, error) {
	ci := imageViewCreateInfo{
		sType:    stImageViewCreateInfo,
		image:    img,
		viewType: ImageViewType2D,
		format:   format,
		subresourceRange: imageSubresourceRange{
			aspectMask: aspect,
			levelCount: 1,
			layerCount: 1,
		},
	}
	var view ImageView
	res := vkCreateImageView(d, unsafe.Pointer(&ci), nil, &view)
	runtime.KeepAlive(&ci)
	return view, res.asError("vkCreateImageView")
}

// DestroyImageView destroys an image view.
func (d Device) DestroyImageView(v ImageView) {
	if v != 0 {
		vkDestroyImageView(d, v, nil)
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
