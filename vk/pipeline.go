package vk

import (
	"runtime"
	"unsafe"

	vulkan "github.com/christerso/vulkan-go/vulkan"
)

// VertexInputBinding mirrors VkVertexInputBindingDescription.
type VertexInputBinding struct {
	Binding   uint32
	Stride    uint32
	InputRate uint32
}

// VertexInputAttribute mirrors VkVertexInputAttributeDescription.
type VertexInputAttribute struct {
	Location uint32
	Binding  uint32
	Format   Format
	Offset   uint32
}

// CreateShaderModule creates a shader module from SPIR-V bytes. The length must
// be a multiple of four.
func (d Device) CreateShaderModule(code []byte) (ShaderModule, error) {
	ci := vulkan.VkShaderModuleCreateInfo{
		SType:    vulkan.VkStructureType(stShaderModuleCreateInfo),
		CodeSize: uintptr(len(code)),
		PCode:    unsafe.Pointer(&code[0]),
	}
	var m vulkan.VkShaderModule
	res := Result(vulkan.VkCreateShaderModule(vulkan.VkDevice(d), unsafe.Pointer(&ci), nil, unsafe.Pointer(&m)))
	runtime.KeepAlive(&ci)
	runtime.KeepAlive(code)
	return ShaderModule(m), res.asError("vkCreateShaderModule")
}

// DestroyShaderModule destroys a shader module.
func (d Device) DestroyShaderModule(m ShaderModule) {
	if m != 0 {
		vulkan.VkDestroyShaderModule(vulkan.VkDevice(d), vulkan.VkShaderModule(m), nil)
	}
}

// CreateColorDepthRenderPass creates a render pass with one color attachment
// that is cleared and presented, and one depth attachment that is cleared.
func (d Device) CreateColorDepthRenderPass(colorFormat, depthFormat Format) (RenderPass, error) {
	attachments := []vulkan.VkAttachmentDescription{
		{
			Format:         vulkan.VkFormat(colorFormat),
			Samples:        SampleCount1,
			LoadOp:         vulkan.VkAttachmentLoadOp(AttachmentLoadOpClear),
			StoreOp:        vulkan.VkAttachmentStoreOp(AttachmentStoreOpStore),
			StencilLoadOp:  vulkan.VkAttachmentLoadOp(AttachmentLoadOpDontCare),
			StencilStoreOp: vulkan.VkAttachmentStoreOp(AttachmentStoreOpDontCare),
			InitialLayout:  vulkan.VkImageLayout(LayoutUndefined),
			FinalLayout:    vulkan.VkImageLayout(LayoutPresentSrcKHR),
		},
		{
			Format:         vulkan.VkFormat(depthFormat),
			Samples:        SampleCount1,
			LoadOp:         vulkan.VkAttachmentLoadOp(AttachmentLoadOpClear),
			StoreOp:        vulkan.VkAttachmentStoreOp(AttachmentStoreOpDontCare),
			StencilLoadOp:  vulkan.VkAttachmentLoadOp(AttachmentLoadOpDontCare),
			StencilStoreOp: vulkan.VkAttachmentStoreOp(AttachmentStoreOpDontCare),
			InitialLayout:  vulkan.VkImageLayout(LayoutUndefined),
			FinalLayout:    vulkan.VkImageLayout(LayoutDepthStencilAttachmentOptimal),
		},
	}
	colorRef := vulkan.VkAttachmentReference{Attachment: 0, Layout: vulkan.VkImageLayout(LayoutColorAttachmentOptimal)}
	depthRef := vulkan.VkAttachmentReference{Attachment: 1, Layout: vulkan.VkImageLayout(LayoutDepthStencilAttachmentOptimal)}
	subpass := vulkan.VkSubpassDescription{
		PipelineBindPoint:       vulkan.VkPipelineBindPoint(BindPointGraphics),
		ColorAttachmentCount:    1,
		PColorAttachments:       unsafe.Pointer(&colorRef),
		PDepthStencilAttachment: unsafe.Pointer(&depthRef),
	}
	dep := vulkan.VkSubpassDependency{
		SrcSubpass:    SubpassExternal,
		DstSubpass:    0,
		SrcStageMask:  StageColorAttachmentOutput | StageEarlyFragmentTests,
		DstStageMask:  StageColorAttachmentOutput | StageEarlyFragmentTests,
		SrcAccessMask: 0,
		DstAccessMask: AccessColorAttachmentWrite | AccessDepthStencilAttachmentWrite,
	}
	ci := vulkan.VkRenderPassCreateInfo{
		SType:           vulkan.VkStructureType(stRenderPassCreateInfo),
		AttachmentCount: uint32(len(attachments)),
		PAttachments:    unsafe.Pointer(&attachments[0]),
		SubpassCount:    1,
		PSubpasses:      unsafe.Pointer(&subpass),
		DependencyCount: 1,
		PDependencies:   unsafe.Pointer(&dep),
	}
	var rp vulkan.VkRenderPass
	res := Result(vulkan.VkCreateRenderPass(vulkan.VkDevice(d), unsafe.Pointer(&ci), nil, unsafe.Pointer(&rp)))
	runtime.KeepAlive(&ci)
	runtime.KeepAlive(attachments)
	runtime.KeepAlive(&colorRef)
	runtime.KeepAlive(&depthRef)
	runtime.KeepAlive(&subpass)
	runtime.KeepAlive(&dep)
	return RenderPass(rp), res.asError("vkCreateRenderPass")
}

// DestroyRenderPass destroys a render pass.
func (d Device) DestroyRenderPass(rp RenderPass) {
	if rp != 0 {
		vulkan.VkDestroyRenderPass(vulkan.VkDevice(d), vulkan.VkRenderPass(rp), nil)
	}
}

// CreateFramebuffer creates a framebuffer over the given attachments.
func (d Device) CreateFramebuffer(rp RenderPass, attachments []ImageView, extent Extent2D) (Framebuffer, error) {
	ci := vulkan.VkFramebufferCreateInfo{
		SType:           vulkan.VkStructureType(stFramebufferCreateInfo),
		RenderPass:      vulkan.VkRenderPass(rp),
		AttachmentCount: uint32(len(attachments)),
		PAttachments:    unsafe.Pointer(&attachments[0]),
		Width:           extent.Width,
		Height:          extent.Height,
		Layers:          1,
	}
	var fb vulkan.VkFramebuffer
	res := Result(vulkan.VkCreateFramebuffer(vulkan.VkDevice(d), unsafe.Pointer(&ci), nil, unsafe.Pointer(&fb)))
	runtime.KeepAlive(&ci)
	runtime.KeepAlive(attachments)
	return Framebuffer(fb), res.asError("vkCreateFramebuffer")
}

// DestroyFramebuffer destroys a framebuffer.
func (d Device) DestroyFramebuffer(fb Framebuffer) {
	if fb != 0 {
		vulkan.VkDestroyFramebuffer(vulkan.VkDevice(d), vulkan.VkFramebuffer(fb), nil)
	}
}

// DescriptorBinding describes one descriptor set layout binding.
type DescriptorBinding struct {
	Binding uint32
	Type    DescriptorType
	Count   uint32
	Stages  uint32
}

// CreateDescriptorSetLayout creates a descriptor set layout.
func (d Device) CreateDescriptorSetLayout(bindings []DescriptorBinding) (DescriptorSetLayout, error) {
	vkb := make([]vulkan.VkDescriptorSetLayoutBinding, len(bindings))
	for i, b := range bindings {
		vkb[i] = vulkan.VkDescriptorSetLayoutBinding{
			Binding:         b.Binding,
			DescriptorType:  vulkan.VkDescriptorType(b.Type),
			DescriptorCount: b.Count,
			StageFlags:      b.Stages,
		}
	}
	ci := vulkan.VkDescriptorSetLayoutCreateInfo{
		SType:        vulkan.VkStructureType(stDescriptorSetLayoutCreateInfo),
		BindingCount: uint32(len(vkb)),
		PBindings:    unsafe.Pointer(&vkb[0]),
	}
	var layout vulkan.VkDescriptorSetLayout
	res := Result(vulkan.VkCreateDescriptorSetLayout(vulkan.VkDevice(d), unsafe.Pointer(&ci), nil, unsafe.Pointer(&layout)))
	runtime.KeepAlive(&ci)
	runtime.KeepAlive(vkb)
	return DescriptorSetLayout(layout), res.asError("vkCreateDescriptorSetLayout")
}

// DestroyDescriptorSetLayout destroys a descriptor set layout.
func (d Device) DestroyDescriptorSetLayout(l DescriptorSetLayout) {
	if l != 0 {
		vulkan.VkDestroyDescriptorSetLayout(vulkan.VkDevice(d), vulkan.VkDescriptorSetLayout(l), nil)
	}
}

// CreatePipelineLayout creates a pipeline layout from set layouts and an
// optional push constant range (size 0 means none).
func (d Device) CreatePipelineLayout(setLayouts []DescriptorSetLayout, pushStage, pushSize uint32) (PipelineLayout, error) {
	ci := vulkan.VkPipelineLayoutCreateInfo{
		SType:          vulkan.VkStructureType(stPipelineLayoutCreateInfo),
		SetLayoutCount: uint32(len(setLayouts)),
	}
	if len(setLayouts) > 0 {
		ci.PSetLayouts = unsafe.Pointer(&setLayouts[0])
	}
	var pcr vulkan.VkPushConstantRange
	if pushSize > 0 {
		pcr = vulkan.VkPushConstantRange{StageFlags: pushStage, Offset: 0, Size: pushSize}
		ci.PushConstantRangeCount = 1
		ci.PPushConstantRanges = unsafe.Pointer(&pcr)
	}
	var layout vulkan.VkPipelineLayout
	res := Result(vulkan.VkCreatePipelineLayout(vulkan.VkDevice(d), unsafe.Pointer(&ci), nil, unsafe.Pointer(&layout)))
	runtime.KeepAlive(&ci)
	runtime.KeepAlive(setLayouts)
	runtime.KeepAlive(&pcr)
	return PipelineLayout(layout), res.asError("vkCreatePipelineLayout")
}

// DestroyPipelineLayout destroys a pipeline layout.
func (d Device) DestroyPipelineLayout(l PipelineLayout) {
	if l != 0 {
		vulkan.VkDestroyPipelineLayout(vulkan.VkDevice(d), vulkan.VkPipelineLayout(l), nil)
	}
}

// GraphicsPipelineConfig describes a graphics pipeline. Viewport and scissor are
// dynamic state, so the extent is supplied at draw time.
type GraphicsPipelineConfig struct {
	Layout       PipelineLayout
	RenderPass   RenderPass
	VertexShader ShaderModule
	FragShader   ShaderModule
	Bindings     []VertexInputBinding
	Attributes   []VertexInputAttribute
	Topology     uint32
	PolygonMode  uint32
	CullMode     uint32
	FrontFace    uint32
	DepthTest    bool
	DepthWrite   bool
	// Blend, when true, enables standard alpha blending on the color
	// attachment (src=SRC_ALPHA, dst=ONE_MINUS_SRC_ALPHA, op=ADD; alpha
	// src=ONE, dst=ONE_MINUS_SRC_ALPHA). Default false keeps opaque output.
	Blend bool
}

// CreateGraphicsPipeline builds a graphics pipeline with dynamic viewport and
// scissor.
func (d Device) CreateGraphicsPipeline(cfg GraphicsPipelineConfig) (Pipeline, error) {
	entry := cstr("main")
	stages := []vulkan.VkPipelineShaderStageCreateInfo{
		{SType: vulkan.VkStructureType(stPipelineShaderStageCreateInfo), Stage: ShaderStageVertex, Module: vulkan.VkShaderModule(cfg.VertexShader), PName: unsafe.Pointer(entry)},
		{SType: vulkan.VkStructureType(stPipelineShaderStageCreateInfo), Stage: ShaderStageFragment, Module: vulkan.VkShaderModule(cfg.FragShader), PName: unsafe.Pointer(entry)},
	}

	// Build the generated vertex input binding/attribute arrays.
	vkBindings := make([]vulkan.VkVertexInputBindingDescription, len(cfg.Bindings))
	for i, b := range cfg.Bindings {
		vkBindings[i] = vulkan.VkVertexInputBindingDescription{Binding: b.Binding, Stride: b.Stride, InputRate: vulkan.VkVertexInputRate(b.InputRate)}
	}
	vkAttrs := make([]vulkan.VkVertexInputAttributeDescription, len(cfg.Attributes))
	for i, a := range cfg.Attributes {
		vkAttrs[i] = vulkan.VkVertexInputAttributeDescription{Location: a.Location, Binding: a.Binding, Format: vulkan.VkFormat(a.Format), Offset: a.Offset}
	}

	vi := vulkan.VkPipelineVertexInputStateCreateInfo{
		SType:                           vulkan.VkStructureType(stPipelineVertexInputStateCreateInfo),
		VertexBindingDescriptionCount:   uint32(len(vkBindings)),
		VertexAttributeDescriptionCount: uint32(len(vkAttrs)),
	}
	if len(vkBindings) > 0 {
		vi.PVertexBindingDescriptions = unsafe.Pointer(&vkBindings[0])
	}
	if len(vkAttrs) > 0 {
		vi.PVertexAttributeDescriptions = unsafe.Pointer(&vkAttrs[0])
	}

	ia := vulkan.VkPipelineInputAssemblyStateCreateInfo{SType: vulkan.VkStructureType(stPipelineInputAssemblyStateCreateInfo), Topology: vulkan.VkPrimitiveTopology(cfg.Topology)}
	vp := vulkan.VkPipelineViewportStateCreateInfo{SType: vulkan.VkStructureType(stPipelineViewportStateCreateInfo), ViewportCount: 1, ScissorCount: 1}
	rs := vulkan.VkPipelineRasterizationStateCreateInfo{
		SType:       vulkan.VkStructureType(stPipelineRasterizationStateCreateInfo),
		PolygonMode: vulkan.VkPolygonMode(cfg.PolygonMode),
		CullMode:    cfg.CullMode,
		FrontFace:   vulkan.VkFrontFace(cfg.FrontFace),
		LineWidth:   1.0,
	}
	ms := vulkan.VkPipelineMultisampleStateCreateInfo{SType: vulkan.VkStructureType(stPipelineMultisampleStateCreateInfo), RasterizationSamples: SampleCount1}
	ds := vulkan.VkPipelineDepthStencilStateCreateInfo{
		SType:          vulkan.VkStructureType(stPipelineDepthStencilStateCreateInfo),
		DepthCompareOp: vulkan.VkCompareOp(CompareLess),
		MaxDepthBounds: 1.0,
	}
	if cfg.DepthTest {
		ds.DepthTestEnable = 1
	}
	if cfg.DepthWrite {
		ds.DepthWriteEnable = 1
	}
	cb := vulkan.VkPipelineColorBlendAttachmentState{ColorWriteMask: 0xF}
	if cfg.Blend {
		cb.BlendEnable = 1
		cb.SrcColorBlendFactor = vulkan.VK_BLEND_FACTOR_SRC_ALPHA
		cb.DstColorBlendFactor = vulkan.VK_BLEND_FACTOR_ONE_MINUS_SRC_ALPHA
		cb.ColorBlendOp = vulkan.VK_BLEND_OP_ADD
		cb.SrcAlphaBlendFactor = vulkan.VK_BLEND_FACTOR_ONE
		cb.DstAlphaBlendFactor = vulkan.VK_BLEND_FACTOR_ONE_MINUS_SRC_ALPHA
		cb.AlphaBlendOp = vulkan.VK_BLEND_OP_ADD
	}
	cbs := vulkan.VkPipelineColorBlendStateCreateInfo{
		SType:           vulkan.VkStructureType(stPipelineColorBlendStateCreateInfo),
		AttachmentCount: 1,
		PAttachments:    unsafe.Pointer(&cb),
	}
	dynStates := []vulkan.VkDynamicState{vulkan.VkDynamicState(DynamicStateViewport), vulkan.VkDynamicState(DynamicStateScissor)}
	dyn := vulkan.VkPipelineDynamicStateCreateInfo{
		SType:             vulkan.VkStructureType(stPipelineDynamicStateCreateInfo),
		DynamicStateCount: uint32(len(dynStates)),
		PDynamicStates:    unsafe.Pointer(&dynStates[0]),
	}

	gp := vulkan.VkGraphicsPipelineCreateInfo{
		SType:               vulkan.VkStructureType(stGraphicsPipelineCreateInfo),
		StageCount:          uint32(len(stages)),
		PStages:             unsafe.Pointer(&stages[0]),
		PVertexInputState:   unsafe.Pointer(&vi),
		PInputAssemblyState: unsafe.Pointer(&ia),
		PViewportState:      unsafe.Pointer(&vp),
		PRasterizationState: unsafe.Pointer(&rs),
		PMultisampleState:   unsafe.Pointer(&ms),
		PDepthStencilState:  unsafe.Pointer(&ds),
		PColorBlendState:    unsafe.Pointer(&cbs),
		PDynamicState:       unsafe.Pointer(&dyn),
		Layout:              vulkan.VkPipelineLayout(cfg.Layout),
		RenderPass:          vulkan.VkRenderPass(cfg.RenderPass),
		BasePipelineIndex:   -1,
	}
	var pipeline vulkan.VkPipeline
	res := Result(vulkan.VkCreateGraphicsPipelines(vulkan.VkDevice(d), 0, 1, unsafe.Pointer(&gp), nil, unsafe.Pointer(&pipeline)))
	runtime.KeepAlive(entry)
	runtime.KeepAlive(stages)
	runtime.KeepAlive(&vi)
	runtime.KeepAlive(vkBindings)
	runtime.KeepAlive(vkAttrs)
	runtime.KeepAlive(&ia)
	runtime.KeepAlive(&vp)
	runtime.KeepAlive(&rs)
	runtime.KeepAlive(&ms)
	runtime.KeepAlive(&ds)
	runtime.KeepAlive(&cb)
	runtime.KeepAlive(&cbs)
	runtime.KeepAlive(dynStates)
	runtime.KeepAlive(&dyn)
	runtime.KeepAlive(&gp)
	return Pipeline(pipeline), res.asError("vkCreateGraphicsPipelines")
}

// DestroyPipeline destroys a pipeline.
func (d Device) DestroyPipeline(p Pipeline) {
	if p != 0 {
		vulkan.VkDestroyPipeline(vulkan.VkDevice(d), vulkan.VkPipeline(p), nil)
	}
}

// CreateDescriptorPool creates a descriptor pool sized for the given counts.
func (d Device) CreateDescriptorPool(maxSets uint32, sizes map[DescriptorType]uint32) (DescriptorPool, error) {
	poolSizes := make([]vulkan.VkDescriptorPoolSize, 0, len(sizes))
	for t, c := range sizes {
		poolSizes = append(poolSizes, vulkan.VkDescriptorPoolSize{Type: vulkan.VkDescriptorType(t), DescriptorCount: c})
	}
	ci := vulkan.VkDescriptorPoolCreateInfo{
		SType:         vulkan.VkStructureType(stDescriptorPoolCreateInfo),
		MaxSets:       maxSets,
		PoolSizeCount: uint32(len(poolSizes)),
		PPoolSizes:    unsafe.Pointer(&poolSizes[0]),
	}
	var pool vulkan.VkDescriptorPool
	res := Result(vulkan.VkCreateDescriptorPool(vulkan.VkDevice(d), unsafe.Pointer(&ci), nil, unsafe.Pointer(&pool)))
	runtime.KeepAlive(&ci)
	runtime.KeepAlive(poolSizes)
	return DescriptorPool(pool), res.asError("vkCreateDescriptorPool")
}

// DestroyDescriptorPool destroys a descriptor pool and its sets.
func (d Device) DestroyDescriptorPool(p DescriptorPool) {
	if p != 0 {
		vulkan.VkDestroyDescriptorPool(vulkan.VkDevice(d), vulkan.VkDescriptorPool(p), nil)
	}
}

// AllocateDescriptorSet allocates a single descriptor set with the given layout.
func (d Device) AllocateDescriptorSet(pool DescriptorPool, layout DescriptorSetLayout) (DescriptorSet, error) {
	l := vulkan.VkDescriptorSetLayout(layout)
	ai := vulkan.VkDescriptorSetAllocateInfo{
		SType:              vulkan.VkStructureType(stDescriptorSetAllocateInfo),
		DescriptorPool:     vulkan.VkDescriptorPool(pool),
		DescriptorSetCount: 1,
		PSetLayouts:        unsafe.Pointer(&l),
	}
	var set vulkan.VkDescriptorSet
	res := Result(vulkan.VkAllocateDescriptorSets(vulkan.VkDevice(d), unsafe.Pointer(&ai), unsafe.Pointer(&set)))
	runtime.KeepAlive(&ai)
	runtime.KeepAlive(&l)
	return DescriptorSet(set), res.asError("vkAllocateDescriptorSets")
}

// UpdateBufferDescriptor points a uniform/storage buffer descriptor at a buffer.
func (d Device) UpdateBufferDescriptor(set DescriptorSet, binding uint32, t DescriptorType, buf Buffer, offset, rang DeviceSize) {
	bi := vulkan.VkDescriptorBufferInfo{Buffer: vulkan.VkBuffer(buf), Offset: vulkan.VkDeviceSize(offset), Range: vulkan.VkDeviceSize(rang)}
	w := vulkan.VkWriteDescriptorSet{
		SType:           vulkan.VkStructureType(stWriteDescriptorSet),
		DstSet:          vulkan.VkDescriptorSet(set),
		DstBinding:      binding,
		DescriptorCount: 1,
		DescriptorType:  vulkan.VkDescriptorType(t),
		PBufferInfo:     unsafe.Pointer(&bi),
	}
	vulkan.VkUpdateDescriptorSets(vulkan.VkDevice(d), 1, unsafe.Pointer(&w), 0, nil)
	runtime.KeepAlive(&bi)
	runtime.KeepAlive(&w)
}
