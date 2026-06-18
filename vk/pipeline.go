package vk

import (
	"runtime"
	"unsafe"
)

// ---- shader module ----

type shaderModuleCreateInfo struct {
	sType    uint32
	pNext    unsafe.Pointer
	flags    uint32
	codeSize uintptr
	pCode    *uint32
}

// ---- render pass ----

type attachmentDescription struct {
	flags          uint32
	format         Format
	samples        uint32
	loadOp         uint32
	storeOp        uint32
	stencilLoadOp  uint32
	stencilStoreOp uint32
	initialLayout  ImageLayout
	finalLayout    ImageLayout
}

type attachmentReference struct {
	attachment uint32
	layout     ImageLayout
}

type subpassDescription struct {
	flags                   uint32
	pipelineBindPoint       uint32
	inputAttachmentCount    uint32
	pInputAttachments       *attachmentReference
	colorAttachmentCount    uint32
	pColorAttachments       *attachmentReference
	pResolveAttachments     *attachmentReference
	pDepthStencilAttachment *attachmentReference
	preserveAttachmentCount uint32
	pPreserveAttachments    *uint32
}

type subpassDependency struct {
	srcSubpass      uint32
	dstSubpass      uint32
	srcStageMask    uint32
	dstStageMask    uint32
	srcAccessMask   uint32
	dstAccessMask   uint32
	dependencyFlags uint32
}

type renderPassCreateInfo struct {
	sType           uint32
	pNext           unsafe.Pointer
	flags           uint32
	attachmentCount uint32
	pAttachments    *attachmentDescription
	subpassCount    uint32
	pSubpasses      *subpassDescription
	dependencyCount uint32
	pDependencies   *subpassDependency
}

type framebufferCreateInfo struct {
	sType           uint32
	pNext           unsafe.Pointer
	flags           uint32
	renderPass      RenderPass
	attachmentCount uint32
	pAttachments    *ImageView
	width           uint32
	height          uint32
	layers          uint32
}

// ---- pipeline state ----

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

type pipelineShaderStageCreateInfo struct {
	sType               uint32
	pNext               unsafe.Pointer
	flags               uint32
	stage               uint32
	module              ShaderModule
	pName               *byte
	pSpecializationInfo unsafe.Pointer
}

type pipelineVertexInputStateCreateInfo struct {
	sType                           uint32
	pNext                           unsafe.Pointer
	flags                           uint32
	vertexBindingDescriptionCount   uint32
	pVertexBindingDescriptions      *VertexInputBinding
	vertexAttributeDescriptionCount uint32
	pVertexAttributeDescriptions    *VertexInputAttribute
}

type pipelineInputAssemblyStateCreateInfo struct {
	sType                  uint32
	pNext                  unsafe.Pointer
	flags                  uint32
	topology               uint32
	primitiveRestartEnable uint32
}

type pipelineViewportStateCreateInfo struct {
	sType         uint32
	pNext         unsafe.Pointer
	flags         uint32
	viewportCount uint32
	pViewports    *Viewport
	scissorCount  uint32
	pScissors     *Rect2D
}

type pipelineRasterizationStateCreateInfo struct {
	sType                   uint32
	pNext                   unsafe.Pointer
	flags                   uint32
	depthClampEnable        uint32
	rasterizerDiscardEnable uint32
	polygonMode             uint32
	cullMode                uint32
	frontFace               uint32
	depthBiasEnable         uint32
	depthBiasConstantFactor float32
	depthBiasClamp          float32
	depthBiasSlopeFactor    float32
	lineWidth               float32
}

type pipelineMultisampleStateCreateInfo struct {
	sType                 uint32
	pNext                 unsafe.Pointer
	flags                 uint32
	rasterizationSamples  uint32
	sampleShadingEnable   uint32
	minSampleShading      float32
	pSampleMask           *uint32
	alphaToCoverageEnable uint32
	alphaToOneEnable      uint32
}

type stencilOpState struct {
	failOp      uint32
	passOp      uint32
	depthFailOp uint32
	compareOp   uint32
	compareMask uint32
	writeMask   uint32
	reference   uint32
}

type pipelineDepthStencilStateCreateInfo struct {
	sType                 uint32
	pNext                 unsafe.Pointer
	flags                 uint32
	depthTestEnable       uint32
	depthWriteEnable      uint32
	depthCompareOp        uint32
	depthBoundsTestEnable uint32
	stencilTestEnable     uint32
	front                 stencilOpState
	back                  stencilOpState
	minDepthBounds        float32
	maxDepthBounds        float32
}

type pipelineColorBlendAttachmentState struct {
	blendEnable         uint32
	srcColorBlendFactor uint32
	dstColorBlendFactor uint32
	colorBlendOp        uint32
	srcAlphaBlendFactor uint32
	dstAlphaBlendFactor uint32
	alphaBlendOp        uint32
	colorWriteMask      uint32
}

type pipelineColorBlendStateCreateInfo struct {
	sType           uint32
	pNext           unsafe.Pointer
	flags           uint32
	logicOpEnable   uint32
	logicOp         uint32
	attachmentCount uint32
	pAttachments    *pipelineColorBlendAttachmentState
	blendConstants  [4]float32
}

type pipelineDynamicStateCreateInfo struct {
	sType             uint32
	pNext             unsafe.Pointer
	flags             uint32
	dynamicStateCount uint32
	pDynamicStates    *uint32
}

type pushConstantRange struct {
	stageFlags uint32
	offset     uint32
	size       uint32
}

type pipelineLayoutCreateInfo struct {
	sType                  uint32
	pNext                  unsafe.Pointer
	flags                  uint32
	setLayoutCount         uint32
	pSetLayouts            *DescriptorSetLayout
	pushConstantRangeCount uint32
	pPushConstantRanges    *pushConstantRange
}

type graphicsPipelineCreateInfo struct {
	sType               uint32
	pNext               unsafe.Pointer
	flags               uint32
	stageCount          uint32
	pStages             *pipelineShaderStageCreateInfo
	pVertexInputState   *pipelineVertexInputStateCreateInfo
	pInputAssemblyState *pipelineInputAssemblyStateCreateInfo
	pTessellationState  unsafe.Pointer
	pViewportState      *pipelineViewportStateCreateInfo
	pRasterizationState *pipelineRasterizationStateCreateInfo
	pMultisampleState   *pipelineMultisampleStateCreateInfo
	pDepthStencilState  *pipelineDepthStencilStateCreateInfo
	pColorBlendState    *pipelineColorBlendStateCreateInfo
	pDynamicState       *pipelineDynamicStateCreateInfo
	layout              PipelineLayout
	renderPass          RenderPass
	subpass             uint32
	basePipelineHandle  Pipeline
	basePipelineIndex   int32
}

// ---- descriptors ----

type descriptorSetLayoutBinding struct {
	binding            uint32
	descriptorType     DescriptorType
	descriptorCount    uint32
	stageFlags         uint32
	pImmutableSamplers *Sampler
}

type descriptorSetLayoutCreateInfo struct {
	sType        uint32
	pNext        unsafe.Pointer
	flags        uint32
	bindingCount uint32
	pBindings    *descriptorSetLayoutBinding
}

type descriptorPoolSize struct {
	typ             DescriptorType
	descriptorCount uint32
}

type descriptorPoolCreateInfo struct {
	sType         uint32
	pNext         unsafe.Pointer
	flags         uint32
	maxSets       uint32
	poolSizeCount uint32
	pPoolSizes    *descriptorPoolSize
}

type descriptorSetAllocateInfo struct {
	sType              uint32
	pNext              unsafe.Pointer
	descriptorPool     DescriptorPool
	descriptorSetCount uint32
	pSetLayouts        *DescriptorSetLayout
}

type descriptorBufferInfo struct {
	buffer Buffer
	offset DeviceSize
	rang   DeviceSize
}

type writeDescriptorSet struct {
	sType            uint32
	pNext            unsafe.Pointer
	dstSet           DescriptorSet
	dstBinding       uint32
	dstArrayElement  uint32
	descriptorCount  uint32
	descriptorType   DescriptorType
	pImageInfo       unsafe.Pointer
	pBufferInfo      *descriptorBufferInfo
	pTexelBufferView unsafe.Pointer
}

var (
	vkCreateShaderModule        func(device Device, pInfo, pAllocator unsafe.Pointer, pModule *ShaderModule) Result
	vkDestroyShaderModule       func(device Device, module ShaderModule, pAllocator unsafe.Pointer)
	vkCreateRenderPass          func(device Device, pInfo, pAllocator unsafe.Pointer, pRP *RenderPass) Result
	vkDestroyRenderPass         func(device Device, rp RenderPass, pAllocator unsafe.Pointer)
	vkCreateFramebuffer         func(device Device, pInfo, pAllocator unsafe.Pointer, pFB *Framebuffer) Result
	vkDestroyFramebuffer        func(device Device, fb Framebuffer, pAllocator unsafe.Pointer)
	vkCreateDescriptorSetLayout func(device Device, pInfo, pAllocator unsafe.Pointer, pLayout *DescriptorSetLayout) Result
	vkDestroyDescriptorSetLayout func(device Device, layout DescriptorSetLayout, pAllocator unsafe.Pointer)
	vkCreatePipelineLayout      func(device Device, pInfo, pAllocator unsafe.Pointer, pLayout *PipelineLayout) Result
	vkDestroyPipelineLayout     func(device Device, layout PipelineLayout, pAllocator unsafe.Pointer)
	vkCreateGraphicsPipelines   func(device Device, cache uint64, count uint32, pInfos, pAllocator unsafe.Pointer, pPipelines *Pipeline) Result
	vkDestroyPipeline           func(device Device, pipeline Pipeline, pAllocator unsafe.Pointer)
	vkCreateDescriptorPool      func(device Device, pInfo, pAllocator unsafe.Pointer, pPool *DescriptorPool) Result
	vkDestroyDescriptorPool     func(device Device, pool DescriptorPool, pAllocator unsafe.Pointer)
	vkAllocateDescriptorSets    func(device Device, pInfo unsafe.Pointer, pSets *DescriptorSet) Result
	vkUpdateDescriptorSets      func(device Device, writeCount uint32, pWrites unsafe.Pointer, copyCount uint32, pCopies unsafe.Pointer)
)

func loadPipelineCommands(device Device) {
	h := uintptr(device)
	bindDeviceProc(&vkCreateShaderModule, h, "vkCreateShaderModule")
	bindDeviceProc(&vkDestroyShaderModule, h, "vkDestroyShaderModule")
	bindDeviceProc(&vkCreateRenderPass, h, "vkCreateRenderPass")
	bindDeviceProc(&vkDestroyRenderPass, h, "vkDestroyRenderPass")
	bindDeviceProc(&vkCreateFramebuffer, h, "vkCreateFramebuffer")
	bindDeviceProc(&vkDestroyFramebuffer, h, "vkDestroyFramebuffer")
	bindDeviceProc(&vkCreateDescriptorSetLayout, h, "vkCreateDescriptorSetLayout")
	bindDeviceProc(&vkDestroyDescriptorSetLayout, h, "vkDestroyDescriptorSetLayout")
	bindDeviceProc(&vkCreatePipelineLayout, h, "vkCreatePipelineLayout")
	bindDeviceProc(&vkDestroyPipelineLayout, h, "vkDestroyPipelineLayout")
	bindDeviceProc(&vkCreateGraphicsPipelines, h, "vkCreateGraphicsPipelines")
	bindDeviceProc(&vkDestroyPipeline, h, "vkDestroyPipeline")
	bindDeviceProc(&vkCreateDescriptorPool, h, "vkCreateDescriptorPool")
	bindDeviceProc(&vkDestroyDescriptorPool, h, "vkDestroyDescriptorPool")
	bindDeviceProc(&vkAllocateDescriptorSets, h, "vkAllocateDescriptorSets")
	bindDeviceProc(&vkUpdateDescriptorSets, h, "vkUpdateDescriptorSets")
}

// CreateShaderModule creates a shader module from SPIR-V bytes. The length must
// be a multiple of four.
func (d Device) CreateShaderModule(code []byte) (ShaderModule, error) {
	ci := shaderModuleCreateInfo{
		sType:    stShaderModuleCreateInfo,
		codeSize: uintptr(len(code)),
		pCode:    (*uint32)(unsafe.Pointer(&code[0])),
	}
	var m ShaderModule
	res := vkCreateShaderModule(d, unsafe.Pointer(&ci), nil, &m)
	runtime.KeepAlive(&ci)
	runtime.KeepAlive(code)
	return m, res.asError("vkCreateShaderModule")
}

// DestroyShaderModule destroys a shader module.
func (d Device) DestroyShaderModule(m ShaderModule) {
	if m != 0 {
		vkDestroyShaderModule(d, m, nil)
	}
}

// CreateColorDepthRenderPass creates a render pass with one color attachment
// that is cleared and presented, and one depth attachment that is cleared.
func (d Device) CreateColorDepthRenderPass(colorFormat, depthFormat Format) (RenderPass, error) {
	attachments := []attachmentDescription{
		{
			format:        colorFormat,
			samples:       SampleCount1,
			loadOp:        AttachmentLoadOpClear,
			storeOp:       AttachmentStoreOpStore,
			stencilLoadOp: AttachmentLoadOpDontCare,
			stencilStoreOp: AttachmentStoreOpDontCare,
			initialLayout: LayoutUndefined,
			finalLayout:   LayoutPresentSrcKHR,
		},
		{
			format:         depthFormat,
			samples:        SampleCount1,
			loadOp:         AttachmentLoadOpClear,
			storeOp:        AttachmentStoreOpDontCare,
			stencilLoadOp:  AttachmentLoadOpDontCare,
			stencilStoreOp: AttachmentStoreOpDontCare,
			initialLayout:  LayoutUndefined,
			finalLayout:    LayoutDepthStencilAttachmentOptimal,
		},
	}
	colorRef := attachmentReference{attachment: 0, layout: LayoutColorAttachmentOptimal}
	depthRef := attachmentReference{attachment: 1, layout: LayoutDepthStencilAttachmentOptimal}
	subpass := subpassDescription{
		pipelineBindPoint:       BindPointGraphics,
		colorAttachmentCount:    1,
		pColorAttachments:       &colorRef,
		pDepthStencilAttachment: &depthRef,
	}
	dep := subpassDependency{
		srcSubpass:    SubpassExternal,
		dstSubpass:    0,
		srcStageMask:  StageColorAttachmentOutput | StageEarlyFragmentTests,
		dstStageMask:  StageColorAttachmentOutput | StageEarlyFragmentTests,
		srcAccessMask: 0,
		dstAccessMask: AccessColorAttachmentWrite | AccessDepthStencilAttachmentWrite,
	}
	ci := renderPassCreateInfo{
		sType:           stRenderPassCreateInfo,
		attachmentCount: uint32(len(attachments)),
		pAttachments:    &attachments[0],
		subpassCount:    1,
		pSubpasses:      &subpass,
		dependencyCount: 1,
		pDependencies:   &dep,
	}
	var rp RenderPass
	res := vkCreateRenderPass(d, unsafe.Pointer(&ci), nil, &rp)
	runtime.KeepAlive(&ci)
	runtime.KeepAlive(attachments)
	runtime.KeepAlive(&colorRef)
	runtime.KeepAlive(&depthRef)
	runtime.KeepAlive(&subpass)
	runtime.KeepAlive(&dep)
	return rp, res.asError("vkCreateRenderPass")
}

// DestroyRenderPass destroys a render pass.
func (d Device) DestroyRenderPass(rp RenderPass) {
	if rp != 0 {
		vkDestroyRenderPass(d, rp, nil)
	}
}

// CreateFramebuffer creates a framebuffer over the given attachments.
func (d Device) CreateFramebuffer(rp RenderPass, attachments []ImageView, extent Extent2D) (Framebuffer, error) {
	ci := framebufferCreateInfo{
		sType:           stFramebufferCreateInfo,
		renderPass:      rp,
		attachmentCount: uint32(len(attachments)),
		pAttachments:    &attachments[0],
		width:           extent.Width,
		height:          extent.Height,
		layers:          1,
	}
	var fb Framebuffer
	res := vkCreateFramebuffer(d, unsafe.Pointer(&ci), nil, &fb)
	runtime.KeepAlive(&ci)
	runtime.KeepAlive(attachments)
	return fb, res.asError("vkCreateFramebuffer")
}

// DestroyFramebuffer destroys a framebuffer.
func (d Device) DestroyFramebuffer(fb Framebuffer) {
	if fb != 0 {
		vkDestroyFramebuffer(d, fb, nil)
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
	vkb := make([]descriptorSetLayoutBinding, len(bindings))
	for i, b := range bindings {
		vkb[i] = descriptorSetLayoutBinding{
			binding:         b.Binding,
			descriptorType:  b.Type,
			descriptorCount: b.Count,
			stageFlags:      b.Stages,
		}
	}
	ci := descriptorSetLayoutCreateInfo{
		sType:        stDescriptorSetLayoutCreateInfo,
		bindingCount: uint32(len(vkb)),
		pBindings:    &vkb[0],
	}
	var layout DescriptorSetLayout
	res := vkCreateDescriptorSetLayout(d, unsafe.Pointer(&ci), nil, &layout)
	runtime.KeepAlive(&ci)
	runtime.KeepAlive(vkb)
	return layout, res.asError("vkCreateDescriptorSetLayout")
}

// DestroyDescriptorSetLayout destroys a descriptor set layout.
func (d Device) DestroyDescriptorSetLayout(l DescriptorSetLayout) {
	if l != 0 {
		vkDestroyDescriptorSetLayout(d, l, nil)
	}
}

// CreatePipelineLayout creates a pipeline layout from set layouts and an
// optional push constant range (size 0 means none).
func (d Device) CreatePipelineLayout(setLayouts []DescriptorSetLayout, pushStage, pushSize uint32) (PipelineLayout, error) {
	ci := pipelineLayoutCreateInfo{
		sType:          stPipelineLayoutCreateInfo,
		setLayoutCount: uint32(len(setLayouts)),
	}
	if len(setLayouts) > 0 {
		ci.pSetLayouts = &setLayouts[0]
	}
	var pcr pushConstantRange
	if pushSize > 0 {
		pcr = pushConstantRange{stageFlags: pushStage, offset: 0, size: pushSize}
		ci.pushConstantRangeCount = 1
		ci.pPushConstantRanges = &pcr
	}
	var layout PipelineLayout
	res := vkCreatePipelineLayout(d, unsafe.Pointer(&ci), nil, &layout)
	runtime.KeepAlive(&ci)
	runtime.KeepAlive(setLayouts)
	runtime.KeepAlive(&pcr)
	return layout, res.asError("vkCreatePipelineLayout")
}

// DestroyPipelineLayout destroys a pipeline layout.
func (d Device) DestroyPipelineLayout(l PipelineLayout) {
	if l != 0 {
		vkDestroyPipelineLayout(d, l, nil)
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
}

// CreateGraphicsPipeline builds a graphics pipeline with dynamic viewport and
// scissor.
func (d Device) CreateGraphicsPipeline(cfg GraphicsPipelineConfig) (Pipeline, error) {
	entry := cstr("main")
	stages := []pipelineShaderStageCreateInfo{
		{sType: stPipelineShaderStageCreateInfo, stage: ShaderStageVertex, module: cfg.VertexShader, pName: entry},
		{sType: stPipelineShaderStageCreateInfo, stage: ShaderStageFragment, module: cfg.FragShader, pName: entry},
	}

	vi := pipelineVertexInputStateCreateInfo{
		sType:                         stPipelineVertexInputStateCreateInfo,
		vertexBindingDescriptionCount: uint32(len(cfg.Bindings)),
		vertexAttributeDescriptionCount: uint32(len(cfg.Attributes)),
	}
	if len(cfg.Bindings) > 0 {
		vi.pVertexBindingDescriptions = &cfg.Bindings[0]
	}
	if len(cfg.Attributes) > 0 {
		vi.pVertexAttributeDescriptions = &cfg.Attributes[0]
	}

	ia := pipelineInputAssemblyStateCreateInfo{sType: stPipelineInputAssemblyStateCreateInfo, topology: cfg.Topology}
	vp := pipelineViewportStateCreateInfo{sType: stPipelineViewportStateCreateInfo, viewportCount: 1, scissorCount: 1}
	rs := pipelineRasterizationStateCreateInfo{
		sType:       stPipelineRasterizationStateCreateInfo,
		polygonMode: cfg.PolygonMode,
		cullMode:    cfg.CullMode,
		frontFace:   cfg.FrontFace,
		lineWidth:   1.0,
	}
	ms := pipelineMultisampleStateCreateInfo{sType: stPipelineMultisampleStateCreateInfo, rasterizationSamples: SampleCount1}
	ds := pipelineDepthStencilStateCreateInfo{
		sType:          stPipelineDepthStencilStateCreateInfo,
		depthCompareOp: CompareLess,
		maxDepthBounds: 1.0,
	}
	if cfg.DepthTest {
		ds.depthTestEnable = 1
	}
	if cfg.DepthWrite {
		ds.depthWriteEnable = 1
	}
	cb := pipelineColorBlendAttachmentState{colorWriteMask: 0xF}
	cbs := pipelineColorBlendStateCreateInfo{
		sType:           stPipelineColorBlendStateCreateInfo,
		attachmentCount: 1,
		pAttachments:    &cb,
	}
	dynStates := []uint32{DynamicStateViewport, DynamicStateScissor}
	dyn := pipelineDynamicStateCreateInfo{
		sType:             stPipelineDynamicStateCreateInfo,
		dynamicStateCount: uint32(len(dynStates)),
		pDynamicStates:    &dynStates[0],
	}

	gp := graphicsPipelineCreateInfo{
		sType:               stGraphicsPipelineCreateInfo,
		stageCount:          uint32(len(stages)),
		pStages:             &stages[0],
		pVertexInputState:   &vi,
		pInputAssemblyState: &ia,
		pViewportState:      &vp,
		pRasterizationState: &rs,
		pMultisampleState:   &ms,
		pDepthStencilState:  &ds,
		pColorBlendState:    &cbs,
		pDynamicState:       &dyn,
		layout:              cfg.Layout,
		renderPass:          cfg.RenderPass,
		basePipelineIndex:   -1,
	}
	var pipeline Pipeline
	res := vkCreateGraphicsPipelines(d, 0, 1, unsafe.Pointer(&gp), nil, &pipeline)
	runtime.KeepAlive(entry)
	runtime.KeepAlive(stages)
	runtime.KeepAlive(&vi)
	runtime.KeepAlive(cfg.Bindings)
	runtime.KeepAlive(cfg.Attributes)
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
	return pipeline, res.asError("vkCreateGraphicsPipelines")
}

// DestroyPipeline destroys a pipeline.
func (d Device) DestroyPipeline(p Pipeline) {
	if p != 0 {
		vkDestroyPipeline(d, p, nil)
	}
}

// CreateDescriptorPool creates a descriptor pool sized for the given counts.
func (d Device) CreateDescriptorPool(maxSets uint32, sizes map[DescriptorType]uint32) (DescriptorPool, error) {
	poolSizes := make([]descriptorPoolSize, 0, len(sizes))
	for t, c := range sizes {
		poolSizes = append(poolSizes, descriptorPoolSize{typ: t, descriptorCount: c})
	}
	ci := descriptorPoolCreateInfo{
		sType:         stDescriptorPoolCreateInfo,
		maxSets:       maxSets,
		poolSizeCount: uint32(len(poolSizes)),
		pPoolSizes:    &poolSizes[0],
	}
	var pool DescriptorPool
	res := vkCreateDescriptorPool(d, unsafe.Pointer(&ci), nil, &pool)
	runtime.KeepAlive(&ci)
	runtime.KeepAlive(poolSizes)
	return pool, res.asError("vkCreateDescriptorPool")
}

// DestroyDescriptorPool destroys a descriptor pool and its sets.
func (d Device) DestroyDescriptorPool(p DescriptorPool) {
	if p != 0 {
		vkDestroyDescriptorPool(d, p, nil)
	}
}

// AllocateDescriptorSet allocates a single descriptor set with the given layout.
func (d Device) AllocateDescriptorSet(pool DescriptorPool, layout DescriptorSetLayout) (DescriptorSet, error) {
	l := layout
	ai := descriptorSetAllocateInfo{
		sType:              stDescriptorSetAllocateInfo,
		descriptorPool:     pool,
		descriptorSetCount: 1,
		pSetLayouts:        &l,
	}
	var set DescriptorSet
	res := vkAllocateDescriptorSets(d, unsafe.Pointer(&ai), &set)
	runtime.KeepAlive(&ai)
	runtime.KeepAlive(&l)
	return set, res.asError("vkAllocateDescriptorSets")
}

// UpdateBufferDescriptor points a uniform/storage buffer descriptor at a buffer.
func (d Device) UpdateBufferDescriptor(set DescriptorSet, binding uint32, t DescriptorType, buf Buffer, offset, rang DeviceSize) {
	bi := descriptorBufferInfo{buffer: buf, offset: offset, rang: rang}
	w := writeDescriptorSet{
		sType:           stWriteDescriptorSet,
		dstSet:          set,
		dstBinding:      binding,
		descriptorCount: 1,
		descriptorType:  t,
		pBufferInfo:     &bi,
	}
	vkUpdateDescriptorSets(d, 1, unsafe.Pointer(&w), 0, nil)
	runtime.KeepAlive(&bi)
	runtime.KeepAlive(&w)
}
