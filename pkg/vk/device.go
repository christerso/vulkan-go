package vk

import (
	"fmt"
	"github.com/christerso/vulkan-go/pkg/vulkan"
)

// PhysicalDevice represents a Vulkan physical device
type PhysicalDevice struct {
	handle     vulkan.PhysicalDevice
	properties PhysicalDeviceProperties
	features   PhysicalDeviceFeatures
	memProps   PhysicalDeviceMemoryProperties
	queueFams  []QueueFamilyProperties
}

// LogicalDevice represents a Vulkan logical device
type LogicalDevice struct {
	handle         vulkan.Device
	physicalDevice *PhysicalDevice
	queues         map[QueueFamily]*Queue
	extensions     []string
}

// DeviceConfig holds configuration for creating a logical device
type DeviceConfig struct {
	RequiredExtensions []string
	OptionalExtensions []string
	RequiredFeatures   PhysicalDeviceFeatures
	QueueCreateInfos   []DeviceQueueCreateInfo
}

// DeviceQueueCreateInfo specifies queue creation parameters
type DeviceQueueCreateInfo struct {
	QueueFamilyIndex uint32
	QueueCount       uint32
	QueuePriorities  []float32
}

// QueueFamily represents different types of queue families
type QueueFamily uint32

const (
	QueueFamilyGraphics QueueFamily = iota
	QueueFamilyCompute
	QueueFamilyTransfer
	QueueFamilyPresent
)

// PhysicalDeviceProperties contains basic properties of a physical device
type PhysicalDeviceProperties struct {
	APIVersion    uint32
	DriverVersion uint32
	VendorID      uint32
	DeviceID      uint32
	DeviceType    DeviceType
	DeviceName    string
	Limits        PhysicalDeviceLimits
}

// PhysicalDeviceLimits contains device limits
type PhysicalDeviceLimits struct {
	MaxImageDimension1D                             uint32
	MaxImageDimension2D                             uint32
	MaxImageDimension3D                             uint32
	MaxImageDimensionCube                           uint32
	MaxImageArrayLayers                             uint32
	MaxTexelBufferElements                          uint32
	MaxUniformBufferRange                           uint32
	MaxStorageBufferRange                           uint32
	MaxPushConstantsSize                            uint32
	MaxMemoryAllocationCount                        uint32
	MaxSamplerAllocationCount                       uint32
	BufferImageGranularity                          uint64
	SparseAddressSpaceSize                          uint64
	MaxBoundDescriptorSets                          uint32
	MaxPerStageDescriptorSamplers                   uint32
	MaxPerStageDescriptorUniformBuffers             uint32
	MaxPerStageDescriptorStorageBuffers             uint32
	MaxPerStageDescriptorSampledImages              uint32
	MaxPerStageDescriptorStorageImages              uint32
	MaxPerStageDescriptorInputAttachments           uint32
	MaxPerStageResources                            uint32
	MaxDescriptorSetSamplers                        uint32
	MaxDescriptorSetUniformBuffers                  uint32
	MaxDescriptorSetUniformBuffersDynamic           uint32
	MaxDescriptorSetStorageBuffers                  uint32
	MaxDescriptorSetStorageBuffersDynamic           uint32
	MaxDescriptorSetSampledImages                   uint32
	MaxDescriptorSetStorageImages                   uint32
	MaxDescriptorSetInputAttachments                uint32
	MaxVertexInputAttributes                        uint32
	MaxVertexInputBindings                          uint32
	MaxVertexInputAttributeOffset                   uint32
	MaxVertexInputBindingStride                     uint32
	MaxVertexOutputComponents                       uint32
	MaxTessellationGenerationLevel                  uint32
	MaxTessellationPatchSize                        uint32
	MaxTessellationControlPerVertexInputComponents  uint32
	MaxTessellationControlPerVertexOutputComponents uint32
	MaxTessellationControlPerPatchOutputComponents  uint32
	MaxTessellationControlTotalOutputComponents     uint32
	MaxTessellationEvaluationInputComponents        uint32
	MaxTessellationEvaluationOutputComponents       uint32
	MaxGeometryShaderInvocations                    uint32
	MaxGeometryInputComponents                      uint32
	MaxGeometryOutputComponents                     uint32
	MaxGeometryOutputVertices                       uint32
	MaxGeometryTotalOutputComponents                uint32
	MaxFragmentInputComponents                      uint32
	MaxFragmentOutputAttachments                    uint32
	MaxFragmentDualSrcAttachments                   uint32
	MaxFragmentCombinedOutputResources              uint32
	MaxComputeSharedMemorySize                      uint32
	MaxComputeWorkGroupCount                        [3]uint32
	MaxComputeWorkGroupInvocations                  uint32
	MaxComputeWorkGroupSize                         [3]uint32
	SubPixelPrecisionBits                           uint32
	SubTexelPrecisionBits                           uint32
	MipmapPrecisionBits                             uint32
	MaxDrawIndexedIndexValue                        uint32
	MaxDrawIndirectCount                            uint32
	MaxSamplerLodBias                               float32
	MaxSamplerAnisotropy                            float32
	MaxViewports                                    uint32
	MaxViewportDimensions                           [2]uint32
	ViewportBoundsRange                             [2]float32
	ViewportSubPixelBits                            uint32
	MinMemoryMapAlignment                           uint64
	MinTexelBufferOffsetAlignment                   uint64
	MinUniformBufferOffsetAlignment                 uint64
	MinStorageBufferOffsetAlignment                 uint64
	MinTexelOffset                                  int32
	MaxTexelOffset                                  uint32
	MinTexelGatherOffset                            int32
	MaxTexelGatherOffset                            uint32
	MinInterpolationOffset                          float32
	MaxInterpolationOffset                          float32
	SubPixelInterpolationOffsetBits                 uint32
	MaxFramebufferWidth                             uint32
	MaxFramebufferHeight                            uint32
	MaxFramebufferLayers                            uint32
	FramebufferColorSampleCounts                    SampleCountFlags
	FramebufferDepthSampleCounts                    SampleCountFlags
	FramebufferStencilSampleCounts                  SampleCountFlags
	FramebufferNoAttachmentsSampleCounts            SampleCountFlags
	MaxColorAttachments                             uint32
	SampledImageColorSampleCounts                   SampleCountFlags
	SampledImageIntegerSampleCounts                 SampleCountFlags
	SampledImageDepthSampleCounts                   SampleCountFlags
	SampledImageStencilSampleCounts                 SampleCountFlags
	StorageImageSampleCounts                        SampleCountFlags
	MaxSampleMaskWords                              uint32
	TimestampComputeAndGraphics                     bool
	TimestampPeriod                                 float32
	MaxClipDistances                                uint32
	MaxCullDistances                                uint32
	MaxCombinedClipAndCullDistances                 uint32
	DiscreteQueuePriorities                         uint32
	PointSizeRange                                  [2]float32
	LineWidthRange                                  [2]float32
	PointSizeGranularity                            float32
	LineWidthGranularity                            float32
	StrictLines                                     bool
	StandardSampleLocations                         bool
	OptimalBufferCopyOffsetAlignment                uint64
	OptimalBufferCopyRowPitchAlignment              uint64
	NonCoherentAtomSize                             uint64
}

// PhysicalDeviceFeatures represents device features that can be enabled
type PhysicalDeviceFeatures struct {
	RobustBufferAccess                      bool
	FullDrawIndexUint32                     bool
	ImageCubeArray                          bool
	IndependentBlend                        bool
	GeometryShader                          bool
	TessellationShader                      bool
	SampleRateShading                       bool
	DualSrcBlend                            bool
	LogicOp                                 bool
	MultiDrawIndirect                       bool
	DrawIndirectFirstInstance               bool
	DepthClamp                              bool
	DepthBiasClamp                          bool
	FillModeNonSolid                        bool
	DepthBounds                             bool
	WideLines                               bool
	LargePoints                             bool
	AlphaToOne                              bool
	MultiViewport                           bool
	SamplerAnisotropy                       bool
	TextureCompressionETC2                  bool
	TextureCompressionASTC_LDR              bool
	TextureCompressionBC                    bool
	OcclusionQueryPrecise                   bool
	PipelineStatisticsQuery                 bool
	VertexPipelineStoresAndAtomics          bool
	FragmentStoresAndAtomics                bool
	ShaderTessellationAndGeometryPointSize  bool
	ShaderImageGatherExtended               bool
	ShaderStorageImageExtendedFormats       bool
	ShaderStorageImageMultisample           bool
	ShaderStorageImageReadWithoutFormat     bool
	ShaderStorageImageWriteWithoutFormat    bool
	ShaderUniformBufferArrayDynamicIndexing bool
	ShaderSampledImageArrayDynamicIndexing  bool
	ShaderStorageBufferArrayDynamicIndexing bool
	ShaderStorageImageArrayDynamicIndexing  bool
	ShaderClipDistance                      bool
	ShaderCullDistance                      bool
	ShaderFloat64                           bool
	ShaderInt64                             bool
	ShaderInt16                             bool
	ShaderResourceResidency                 bool
	ShaderResourceMinLod                    bool
	SparseBinding                           bool
	SparseResidencyBuffer                   bool
	SparseResidencyImage2D                  bool
	SparseResidencyImage3D                  bool
	SparseResidency2Samples                 bool
	SparseResidency4Samples                 bool
	SparseResidency8Samples                 bool
	SparseResidency16Samples                bool
	SparseResidencyAliased                  bool
	VariableMultisampleRate                 bool
	InheritedQueries                        bool
}

// PhysicalDeviceMemoryProperties describes memory heaps and types
type PhysicalDeviceMemoryProperties struct {
	MemoryTypeCount uint32
	MemoryTypes     [32]MemoryType
	MemoryHeapCount uint32
	MemoryHeaps     [16]MemoryHeap
}

// MemoryType describes properties of a memory type
type MemoryType struct {
	PropertyFlags MemoryPropertyFlags
	HeapIndex     uint32
}

// MemoryHeap describes a memory heap
type MemoryHeap struct {
	Size  uint64
	Flags MemoryHeapFlags
}

// QueueFamilyProperties describes properties of a queue family
type QueueFamilyProperties struct {
	QueueFlags                   QueueFlags
	QueueCount                   uint32
	TimestampValidBits           uint32
	MinImageTransferGranularity  Extent3D
}

// Queue represents a Vulkan queue
type Queue struct {
	handle           vulkan.Queue
	familyIndex      uint32
	queueIndex       uint32
	flags            QueueFlags
}

// Extent3D represents 3D extents
type Extent3D struct {
	Width  uint32
	Height uint32
	Depth  uint32
}

// Flag types
type SampleCountFlags uint32
type MemoryPropertyFlags uint32
type MemoryHeapFlags uint32
type QueueFlags uint32

// Sample count flags
const (
	SampleCount1Bit  SampleCountFlags = 0x00000001
	SampleCount2Bit  SampleCountFlags = 0x00000002
	SampleCount4Bit  SampleCountFlags = 0x00000004
	SampleCount8Bit  SampleCountFlags = 0x00000008
	SampleCount16Bit SampleCountFlags = 0x00000010
	SampleCount32Bit SampleCountFlags = 0x00000020
	SampleCount64Bit SampleCountFlags = 0x00000040
)

// Memory property flags
const (
	MemoryPropertyDeviceLocalBit     MemoryPropertyFlags = 0x00000001
	MemoryPropertyHostVisibleBit     MemoryPropertyFlags = 0x00000002
	MemoryPropertyHostCoherentBit    MemoryPropertyFlags = 0x00000004
	MemoryPropertyHostCachedBit      MemoryPropertyFlags = 0x00000008
	MemoryPropertyLazilyAllocatedBit MemoryPropertyFlags = 0x00000010
)

// Memory heap flags
const (
	MemoryHeapDeviceLocalBit MemoryHeapFlags = 0x00000001
)

// Queue flags
const (
	QueueGraphicsBit       QueueFlags = 0x00000001
	QueueComputeBit        QueueFlags = 0x00000002
	QueueTransferBit       QueueFlags = 0x00000004
	QueueSparseBindingBit  QueueFlags = 0x00000008
	QueueProtectedBit      QueueFlags = 0x00000010
)

// GetProperties returns the properties of the physical device
func (pd *PhysicalDevice) GetProperties() PhysicalDeviceProperties {
	return pd.properties
}

// GetFeatures returns the features supported by the physical device
func (pd *PhysicalDevice) GetFeatures() PhysicalDeviceFeatures {
	return pd.features
}

// GetMemoryProperties returns the memory properties of the physical device
func (pd *PhysicalDevice) GetMemoryProperties() PhysicalDeviceMemoryProperties {
	return pd.memProps
}

// GetQueueFamilyProperties returns the queue family properties
func (pd *PhysicalDevice) GetQueueFamilyProperties() []QueueFamilyProperties {
	return pd.queueFams
}

// FindQueueFamily finds a queue family with the specified flags
func (pd *PhysicalDevice) FindQueueFamily(flags QueueFlags) (uint32, bool) {
	for i, qf := range pd.queueFams {
		if qf.QueueFlags&flags == flags {
			return uint32(i), true
		}
	}
	return 0, false
}

// FindMemoryType finds a memory type with the specified properties
func (pd *PhysicalDevice) FindMemoryType(typeFilter uint32, properties MemoryPropertyFlags) (uint32, bool) {
	for i := uint32(0); i < pd.memProps.MemoryTypeCount; i++ {
		if (typeFilter&(1<<i)) != 0 && 
		   (pd.memProps.MemoryTypes[i].PropertyFlags&properties) == properties {
			return i, true
		}
	}
	return 0, false
}

// CreateLogicalDevice creates a logical device from the physical device
func (pd *PhysicalDevice) CreateLogicalDevice(config DeviceConfig) (*LogicalDevice, error) {
	// Check extension support
	availableExtensions, err := pd.enumerateDeviceExtensions()
	if err != nil {
		return nil, fmt.Errorf("failed to enumerate device extensions: %w", err)
	}

	// Verify required extensions are available
	for _, ext := range config.RequiredExtensions {
		if !isExtensionSupported(ext, availableExtensions) {
			return nil, fmt.Errorf("required extension %s is not supported", ext)
		}
	}

	// Add optional extensions that are available
	enabledExtensions := make([]string, len(config.RequiredExtensions))
	copy(enabledExtensions, config.RequiredExtensions)

	for _, ext := range config.OptionalExtensions {
		if isExtensionSupported(ext, availableExtensions) {
			enabledExtensions = append(enabledExtensions, ext)
		}
	}

	// TODO: Implement actual device creation
	device := &LogicalDevice{
		physicalDevice: pd,
		queues:         make(map[QueueFamily]*Queue),
		extensions:     enabledExtensions,
	}

	// Create queues based on queue create infos
	for _, qci := range config.QueueCreateInfos {
		for queueIndex := uint32(0); queueIndex < qci.QueueCount; queueIndex++ {
			queue := &Queue{
				familyIndex: qci.QueueFamilyIndex,
				queueIndex:  queueIndex,
				flags:       pd.queueFams[qci.QueueFamilyIndex].QueueFlags,
			}
			
			// TODO: Get actual queue handle from Vulkan
			
			// Determine queue family type
			var queueFamily QueueFamily
			if queue.flags&QueueGraphicsBit != 0 {
				queueFamily = QueueFamilyGraphics
			} else if queue.flags&QueueComputeBit != 0 {
				queueFamily = QueueFamilyCompute
			} else if queue.flags&QueueTransferBit != 0 {
				queueFamily = QueueFamilyTransfer
			}
			
			device.queues[queueFamily] = queue
		}
	}

	return device, nil
}

// Destroy cleans up the logical device
func (d *LogicalDevice) Destroy() {
	if d.handle != 0 {
		// TODO: Call vkDestroyDevice
		d.handle = 0
	}
	d.queues = nil
}

// Handle returns the underlying Vulkan device handle
func (d *LogicalDevice) Handle() vulkan.Device {
	return d.handle
}

// GetQueue returns a queue of the specified family
func (d *LogicalDevice) GetQueue(family QueueFamily) *Queue {
	return d.queues[family]
}

// WaitIdle waits for all operations on the device to complete
func (d *LogicalDevice) WaitIdle() error {
	// TODO: Call vkDeviceWaitIdle
	return nil
}

// GetPhysicalDevice returns the physical device this logical device was created from
func (d *LogicalDevice) GetPhysicalDevice() *PhysicalDevice {
	return d.physicalDevice
}

// Helper functions

func (pd *PhysicalDevice) enumerateDeviceExtensions() ([]ExtensionProperties, error) {
	// TODO: Implement vkEnumerateDeviceExtensionProperties
	return []ExtensionProperties{}, nil
}

// DefaultDeviceConfig returns a default device configuration
func DefaultDeviceConfig(physicalDevice *PhysicalDevice) DeviceConfig {
	// Find graphics queue family
	graphicsFamily, hasGraphics := physicalDevice.FindQueueFamily(QueueGraphicsBit)
	if !hasGraphics {
		panic("No graphics queue family found")
	}

	return DeviceConfig{
		RequiredExtensions: []string{},
		OptionalExtensions: []string{
			"VK_KHR_swapchain", // For presentation
		},
		RequiredFeatures: PhysicalDeviceFeatures{}, // No specific features required
		QueueCreateInfos: []DeviceQueueCreateInfo{
			{
				QueueFamilyIndex: graphicsFamily,
				QueueCount:       1,
				QueuePriorities:  []float32{1.0},
			},
		},
	}
}

// Queue operations

// Submit submits command buffers to the queue
func (q *Queue) Submit(commandBuffers []*CommandBuffer, fence *Fence) error {
	// TODO: Implement vkQueueSubmit
	return nil
}

// Present presents images to the surface (for present queues)
func (q *Queue) Present(presentInfo *PresentInfo) error {
	// TODO: Implement vkQueuePresentKHR
	return nil
}

// WaitIdle waits for all operations on the queue to complete
func (q *Queue) WaitIdle() error {
	// TODO: Implement vkQueueWaitIdle
	return nil
}

// Placeholder types for future implementation
type CommandBuffer struct{}
type Fence struct{}
type PresentInfo struct{}