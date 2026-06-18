package vk

import (
	"runtime"
	"unsafe"

	vulkan "github.com/christerso/vulkan-go/vulkan"
)

// Command buffer level (VkCommandBufferLevel).
const commandBufferLevelPrimary uint32 = 0

// CopyBuffer records a full-size buffer copy of size bytes.
func (c CommandBuffer) CopyBuffer(src, dst Buffer, size DeviceSize) {
	region := vulkan.VkBufferCopy{Size: vulkan.VkDeviceSize(size)}
	vulkan.VkCmdCopyBuffer(vulkan.VkCommandBuffer(c), vulkan.VkBuffer(src), vulkan.VkBuffer(dst), 1, unsafe.Pointer(&region))
	runtime.KeepAlive(&region)
}

// CreateCommandPool creates a command pool allowing individual buffer resets.
func (d Device) CreateCommandPool(family uint32) (CommandPool, error) {
	ci := vulkan.VkCommandPoolCreateInfo{
		SType:            vulkan.VkStructureType(stCommandPoolCreateInfo),
		Flags:            CommandPoolResetCommandBuffer,
		QueueFamilyIndex: family,
	}
	var pool vulkan.VkCommandPool
	res := Result(vulkan.VkCreateCommandPool(vulkan.VkDevice(d), unsafe.Pointer(&ci), nil, unsafe.Pointer(&pool)))
	runtime.KeepAlive(&ci)
	return CommandPool(pool), res.asError("vkCreateCommandPool")
}

// DestroyCommandPool destroys a command pool and its buffers.
func (d Device) DestroyCommandPool(pool CommandPool) {
	if pool != 0 {
		vulkan.VkDestroyCommandPool(vulkan.VkDevice(d), vulkan.VkCommandPool(pool), nil)
	}
}

// AllocateCommandBuffers allocates count primary command buffers.
func (d Device) AllocateCommandBuffers(pool CommandPool, count uint32) ([]CommandBuffer, error) {
	ai := vulkan.VkCommandBufferAllocateInfo{
		SType:              vulkan.VkStructureType(stCommandBufferAllocateInfo),
		CommandPool:        vulkan.VkCommandPool(pool),
		Level:              vulkan.VkCommandBufferLevel(commandBufferLevelPrimary),
		CommandBufferCount: count,
	}
	buffers := make([]CommandBuffer, count)
	res := Result(vulkan.VkAllocateCommandBuffers(vulkan.VkDevice(d), unsafe.Pointer(&ai), unsafe.Pointer(&buffers[0])))
	runtime.KeepAlive(&ai)
	return buffers, res.asError("vkAllocateCommandBuffers")
}

// Begin starts recording. flags is a VkCommandBufferUsageFlags value.
func (c CommandBuffer) Begin(flags uint32) error {
	bi := vulkan.VkCommandBufferBeginInfo{SType: vulkan.VkStructureType(stCommandBufferBeginInfo), Flags: flags}
	res := Result(vulkan.VkBeginCommandBuffer(vulkan.VkCommandBuffer(c), unsafe.Pointer(&bi)))
	runtime.KeepAlive(&bi)
	return res.asError("vkBeginCommandBuffer")
}

// End finishes recording.
func (c CommandBuffer) End() error {
	return Result(vulkan.VkEndCommandBuffer(vulkan.VkCommandBuffer(c))).asError("vkEndCommandBuffer")
}

// Reset resets the command buffer.
func (c CommandBuffer) Reset() error {
	return Result(vulkan.VkResetCommandBuffer(vulkan.VkCommandBuffer(c), 0)).asError("vkResetCommandBuffer")
}

// BeginRenderPass begins a render pass with inline contents. clears holds one
// ClearValue per attachment (color then depth).
func (c CommandBuffer) BeginRenderPass(rp RenderPass, fb Framebuffer, area Rect2D, clears []ClearValue) {
	bi := vulkan.VkRenderPassBeginInfo{
		SType:           vulkan.VkStructureType(stRenderPassBeginInfo),
		RenderPass:      vulkan.VkRenderPass(rp),
		Framebuffer:     vulkan.VkFramebuffer(fb),
		RenderArea:      vulkan.VkRect2D{Offset: vulkan.VkOffset2D{X: area.Offset.X, Y: area.Offset.Y}, Extent: vulkan.VkExtent2D{Width: area.Extent.Width, Height: area.Extent.Height}},
		ClearValueCount: uint32(len(clears)),
	}
	if len(clears) > 0 {
		bi.PClearValues = unsafe.Pointer(&clears[0])
	}
	vulkan.VkCmdBeginRenderPass(vulkan.VkCommandBuffer(c), unsafe.Pointer(&bi), vulkan.VkSubpassContents(SubpassContentsInline))
	runtime.KeepAlive(&bi)
	runtime.KeepAlive(clears)
}

// EndRenderPass ends the current render pass.
func (c CommandBuffer) EndRenderPass() { vulkan.VkCmdEndRenderPass(vulkan.VkCommandBuffer(c)) }

// BindPipeline binds a graphics pipeline.
func (c CommandBuffer) BindPipeline(p Pipeline) {
	vulkan.VkCmdBindPipeline(vulkan.VkCommandBuffer(c), vulkan.VkPipelineBindPoint(BindPointGraphics), vulkan.VkPipeline(p))
}

// SetViewport sets a single viewport.
func (c CommandBuffer) SetViewport(v Viewport) {
	vv := vulkan.VkViewport{X: v.X, Y: v.Y, Width: v.Width, Height: v.Height, MinDepth: v.MinDepth, MaxDepth: v.MaxDepth}
	vulkan.VkCmdSetViewport(vulkan.VkCommandBuffer(c), 0, 1, unsafe.Pointer(&vv))
	runtime.KeepAlive(&vv)
}

// SetScissor sets a single scissor rectangle.
func (c CommandBuffer) SetScissor(r Rect2D) {
	vr := vulkan.VkRect2D{Offset: vulkan.VkOffset2D{X: r.Offset.X, Y: r.Offset.Y}, Extent: vulkan.VkExtent2D{Width: r.Extent.Width, Height: r.Extent.Height}}
	vulkan.VkCmdSetScissor(vulkan.VkCommandBuffer(c), 0, 1, unsafe.Pointer(&vr))
	runtime.KeepAlive(&vr)
}

// BindVertexBuffer binds one vertex buffer at binding 0 with the given offset.
func (c CommandBuffer) BindVertexBuffer(b Buffer, offset DeviceSize) {
	vb := vulkan.VkBuffer(b)
	off := vulkan.VkDeviceSize(offset)
	vulkan.VkCmdBindVertexBuffers(vulkan.VkCommandBuffer(c), 0, 1, unsafe.Pointer(&vb), unsafe.Pointer(&off))
	runtime.KeepAlive(&vb)
	runtime.KeepAlive(&off)
}

// BindVertexBuffers binds buffers starting at firstBinding. The offsets slice
// must match buffers in length.
func (c CommandBuffer) BindVertexBuffers(firstBinding uint32, buffers []Buffer, offsets []DeviceSize) {
	vulkan.VkCmdBindVertexBuffers(vulkan.VkCommandBuffer(c), firstBinding, uint32(len(buffers)), unsafe.Pointer(&buffers[0]), unsafe.Pointer(&offsets[0]))
	runtime.KeepAlive(buffers)
	runtime.KeepAlive(offsets)
}

// BindIndexBuffer binds an index buffer.
func (c CommandBuffer) BindIndexBuffer(b Buffer, offset DeviceSize, indexType uint32) {
	vulkan.VkCmdBindIndexBuffer(vulkan.VkCommandBuffer(c), vulkan.VkBuffer(b), vulkan.VkDeviceSize(offset), vulkan.VkIndexType(indexType))
}

// BindDescriptorSet binds a single descriptor set at firstSet.
func (c CommandBuffer) BindDescriptorSet(layout PipelineLayout, firstSet uint32, set DescriptorSet) {
	vs := vulkan.VkDescriptorSet(set)
	vulkan.VkCmdBindDescriptorSets(vulkan.VkCommandBuffer(c), vulkan.VkPipelineBindPoint(BindPointGraphics), vulkan.VkPipelineLayout(layout), firstSet, 1, unsafe.Pointer(&vs), 0, nil)
	runtime.KeepAlive(&vs)
}

// PushConstants uploads push constant data.
func (c CommandBuffer) PushConstants(layout PipelineLayout, stage, offset uint32, data unsafe.Pointer, size uint32) {
	vulkan.VkCmdPushConstants(vulkan.VkCommandBuffer(c), vulkan.VkPipelineLayout(layout), stage, offset, size, data)
}

// Draw issues a non-indexed draw.
func (c CommandBuffer) Draw(vertexCount, instanceCount, firstVertex, firstInstance uint32) {
	vulkan.VkCmdDraw(vulkan.VkCommandBuffer(c), vertexCount, instanceCount, firstVertex, firstInstance)
}

// DrawIndexed issues an indexed draw.
func (c CommandBuffer) DrawIndexed(indexCount, instanceCount, firstIndex uint32, vertexOffset int32, firstInstance uint32) {
	vulkan.VkCmdDrawIndexed(vulkan.VkCommandBuffer(c), indexCount, instanceCount, firstIndex, vertexOffset, firstInstance)
}
