package vk

// Non-dispatchable handles are uint64 on all platforms.
type (
	SurfaceKHR            uint64
	SwapchainKHR          uint64
	Image                 uint64
	ImageView             uint64
	RenderPass            uint64
	Framebuffer           uint64
	ShaderModule          uint64
	PipelineLayout        uint64
	Pipeline              uint64
	DescriptorSetLayout   uint64
	DescriptorPool        uint64
	DescriptorSet         uint64
	CommandPool           uint64
	Buffer                uint64
	DeviceMemory          uint64
	Semaphore             uint64
	Fence                 uint64
	Sampler               uint64
)

// DeviceSize is VkDeviceSize.
type DeviceSize uint64

// Structure type values used by this binding.
const (
	stSubmitInfo                            uint32 = 4
	stMemoryAllocateInfo                    uint32 = 5
	stFenceCreateInfo                       uint32 = 8
	stSemaphoreCreateInfo                   uint32 = 9
	stBufferCreateInfo                      uint32 = 12
	stImageCreateInfo                       uint32 = 14
	stImageViewCreateInfo                   uint32 = 15
	stShaderModuleCreateInfo                uint32 = 16
	stPipelineShaderStageCreateInfo         uint32 = 18
	stPipelineVertexInputStateCreateInfo    uint32 = 19
	stPipelineInputAssemblyStateCreateInfo  uint32 = 20
	stPipelineViewportStateCreateInfo       uint32 = 22
	stPipelineRasterizationStateCreateInfo  uint32 = 23
	stPipelineMultisampleStateCreateInfo    uint32 = 24
	stPipelineDepthStencilStateCreateInfo   uint32 = 25
	stPipelineColorBlendStateCreateInfo     uint32 = 26
	stPipelineDynamicStateCreateInfo        uint32 = 27
	stGraphicsPipelineCreateInfo            uint32 = 28
	stPipelineLayoutCreateInfo              uint32 = 30
	stDescriptorSetLayoutCreateInfo         uint32 = 32
	stDescriptorPoolCreateInfo              uint32 = 33
	stDescriptorSetAllocateInfo             uint32 = 34
	stWriteDescriptorSet                    uint32 = 35
	stFramebufferCreateInfo                 uint32 = 37
	stRenderPassCreateInfo                  uint32 = 38
	stCommandPoolCreateInfo                 uint32 = 39
	stCommandBufferAllocateInfo             uint32 = 40
	stCommandBufferBeginInfo                uint32 = 42
	stRenderPassBeginInfo                   uint32 = 43
	stImageMemoryBarrier                    uint32 = 45
	stSwapchainCreateInfoKHR                uint32 = 1000001000
	stPresentInfoKHR                        uint32 = 1000001001
	stDebugUtilsMessengerCreateInfoEXT      uint32 = 1000128004
)

// Format values (VkFormat), the subset the binding uses.
type Format uint32

const (
	FormatUndefined         Format = 0
	FormatR8G8B8A8Unorm     Format = 37
	FormatR32Sfloat         Format = 100
	FormatB8G8R8A8Unorm     Format = 44
	FormatB8G8R8A8Srgb      Format = 50
	FormatR32G32Sfloat      Format = 103
	FormatR32G32B32Sfloat   Format = 106
	FormatR32G32B32A32Sfloat Format = 109
	FormatD32Sfloat         Format = 126
	FormatD24UnormS8Uint    Format = 129
)

// Image layouts (VkImageLayout).
type ImageLayout uint32

const (
	LayoutUndefined                     ImageLayout = 0
	LayoutGeneral                       ImageLayout = 1
	LayoutColorAttachmentOptimal        ImageLayout = 2
	LayoutDepthStencilAttachmentOptimal ImageLayout = 3
	LayoutShaderReadOnlyOptimal         ImageLayout = 5
	LayoutTransferDstOptimal            ImageLayout = 7
	LayoutPresentSrcKHR                 ImageLayout = 1000001002
)

// Image usage flag bits (VkImageUsageFlagBits).
const (
	ImageUsageTransferDst            uint32 = 0x00000002
	ImageUsageSampled                uint32 = 0x00000004
	ImageUsageColorAttachment        uint32 = 0x00000010
	ImageUsageDepthStencilAttachment uint32 = 0x00000020
)

// Buffer usage flag bits (VkBufferUsageFlagBits).
const (
	BufferUsageTransferSrc   uint32 = 0x00000001
	BufferUsageTransferDst   uint32 = 0x00000002
	BufferUsageUniformBuffer uint32 = 0x00000010
	BufferUsageStorageBuffer uint32 = 0x00000020
	BufferUsageIndexBuffer   uint32 = 0x00000040
	BufferUsageVertexBuffer  uint32 = 0x00000080
)

// Memory property flag bits (VkMemoryPropertyFlagBits).
const (
	MemoryDeviceLocal  uint32 = 0x00000001
	MemoryHostVisible  uint32 = 0x00000002
	MemoryHostCoherent uint32 = 0x00000004
)

// Sample count (VkSampleCountFlagBits).
const (
	SampleCount1 uint32 = 0x00000001
)

// Pipeline stage flag bits (VkPipelineStageFlagBits).
const (
	StageTopOfPipe            uint32 = 0x00000001
	StageVertexShader         uint32 = 0x00000008
	StageEarlyFragmentTests   uint32 = 0x00000100
	StageColorAttachmentOutput uint32 = 0x00000400
	StageTransfer             uint32 = 0x00001000
	StageBottomOfPipe         uint32 = 0x00002000
)

// Access flag bits (VkAccessFlagBits).
const (
	AccessColorAttachmentWrite        uint32 = 0x00000100
	AccessDepthStencilAttachmentWrite uint32 = 0x00000400
	AccessTransferWrite               uint32 = 0x00001000
)

// Image aspect flag bits (VkImageAspectFlagBits).
const (
	AspectColor uint32 = 0x00000001
	AspectDepth uint32 = 0x00000002
)

// Shader stage flag bits (VkShaderStageFlagBits).
const (
	ShaderStageVertex   uint32 = 0x00000001
	ShaderStageFragment uint32 = 0x00000010
)

// Descriptor type (VkDescriptorType).
type DescriptorType uint32

const (
	DescriptorUniformBuffer        DescriptorType = 6
	DescriptorStorageBuffer        DescriptorType = 7
	DescriptorCombinedImageSampler DescriptorType = 1
)

// Vertex input rate (VkVertexInputRate).
const (
	VertexInputRateVertex   uint32 = 0
	VertexInputRateInstance uint32 = 1
)

// Primitive topology (VkPrimitiveTopology).
const (
	TopologyTriangleList uint32 = 3
)

// Polygon mode (VkPolygonMode).
const (
	PolygonFill uint32 = 0
	PolygonLine uint32 = 1
)

// Cull mode flag bits (VkCullModeFlagBits).
const (
	CullNone  uint32 = 0
	CullBack  uint32 = 0x00000002
	CullFront uint32 = 0x00000001
)

// Front face (VkFrontFace).
const (
	FrontFaceCounterClockwise uint32 = 0
	FrontFaceClockwise        uint32 = 1
)

// Compare op (VkCompareOp).
const (
	CompareLess          uint32 = 1
	CompareLessOrEqual   uint32 = 3
	CompareGreaterOrEqual uint32 = 5
)

// Index type (VkIndexType).
const (
	IndexTypeUint16 uint32 = 0
	IndexTypeUint32 uint32 = 1
)

// Load/store ops (VkAttachmentLoadOp / VkAttachmentStoreOp).
const (
	AttachmentLoadOpLoad     uint32 = 0
	AttachmentLoadOpClear    uint32 = 1
	AttachmentLoadOpDontCare uint32 = 2
	AttachmentStoreOpStore   uint32 = 0
	AttachmentStoreOpDontCare uint32 = 1
)

// Pipeline bind point (VkPipelineBindPoint).
const (
	BindPointGraphics uint32 = 0
)

// Subpass contents (VkSubpassContents).
const (
	SubpassContentsInline uint32 = 0
)

// Command buffer usage flag bits.
const (
	CommandBufferOneTimeSubmit uint32 = 0x00000001
)

// Command pool create flag bits.
const (
	CommandPoolResetCommandBuffer uint32 = 0x00000002
)

// Fence create flag bits.
const (
	FenceCreateSignaled uint32 = 0x00000001
)

// Sharing mode (VkSharingMode).
const (
	SharingModeExclusive uint32 = 0
)

// Present mode (VkPresentModeKHR).
type PresentMode uint32

const (
	PresentModeImmediate   PresentMode = 0
	PresentModeMailbox     PresentMode = 1
	PresentModeFIFO        PresentMode = 2
	PresentModeFIFORelaxed PresentMode = 3
)

// Composite alpha flag bits (VkCompositeAlphaFlagBitsKHR).
const (
	CompositeAlphaOpaque uint32 = 0x00000001
)

// Surface transform flag bits (VkSurfaceTransformFlagBitsKHR).
const (
	SurfaceTransformIdentity uint32 = 0x00000001
)

// Image view type (VkImageViewType).
const (
	ImageViewType2D uint32 = 1
)

// Image type (VkImageType).
const (
	ImageType2D uint32 = 1
)

// Image tiling (VkImageTiling).
const (
	ImageTilingOptimal uint32 = 0
)

// Dynamic state (VkDynamicState).
const (
	DynamicStateViewport uint32 = 0
	DynamicStateScissor  uint32 = 1
)

// Subpass external constant.
const SubpassExternal uint32 = 0xFFFFFFFF

// WholeSize maps to VK_WHOLE_SIZE.
const WholeSize DeviceSize = 0xFFFFFFFFFFFFFFFF

// Small shared structs.

type Offset2D struct{ X, Y int32 }
type Extent2D struct{ Width, Height uint32 }
type Extent3D struct{ Width, Height, Depth uint32 }
type Rect2D struct {
	Offset Offset2D
	Extent Extent2D
}

// Viewport mirrors VkViewport.
type Viewport struct {
	X, Y, Width, Height, MinDepth, MaxDepth float32
}

// ClearColor and ClearDepthStencil compose a VkClearValue (a 16-byte union).
type ClearValue [16]byte
