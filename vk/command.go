package vk

import (
	"runtime"
	"unsafe"
)

// Command buffer level (VkCommandBufferLevel).
const commandBufferLevelPrimary uint32 = 0

type commandPoolCreateInfo struct {
	sType            uint32
	pNext            unsafe.Pointer
	flags            uint32
	queueFamilyIndex uint32
}

type commandBufferAllocateInfo struct {
	sType              uint32
	pNext              unsafe.Pointer
	commandPool        CommandPool
	level              uint32
	commandBufferCount uint32
}

type commandBufferBeginInfo struct {
	sType            uint32
	pNext            unsafe.Pointer
	flags            uint32
	pInheritanceInfo unsafe.Pointer
}

type renderPassBeginInfo struct {
	sType           uint32
	pNext           unsafe.Pointer
	renderPass      RenderPass
	framebuffer     Framebuffer
	renderArea      Rect2D
	clearValueCount uint32
	pClearValues    *ClearValue
}

var (
	vkCreateCommandPool      func(device Device, pInfo, pAllocator unsafe.Pointer, pPool *CommandPool) Result
	vkDestroyCommandPool     func(device Device, pool CommandPool, pAllocator unsafe.Pointer)
	vkAllocateCommandBuffers func(device Device, pInfo unsafe.Pointer, pBuffers *CommandBuffer) Result
	vkBeginCommandBuffer     func(cmd CommandBuffer, pInfo unsafe.Pointer) Result
	vkEndCommandBuffer       func(cmd CommandBuffer) Result
	vkResetCommandBuffer     func(cmd CommandBuffer, flags uint32) Result
	vkCmdBeginRenderPass     func(cmd CommandBuffer, pInfo unsafe.Pointer, contents uint32)
	vkCmdEndRenderPass       func(cmd CommandBuffer)
	vkCmdBindPipeline        func(cmd CommandBuffer, bindPoint uint32, pipeline Pipeline)
	vkCmdSetViewport         func(cmd CommandBuffer, first, count uint32, pViewports *Viewport)
	vkCmdSetScissor          func(cmd CommandBuffer, first, count uint32, pScissors *Rect2D)
	vkCmdBindVertexBuffers   func(cmd CommandBuffer, first, count uint32, pBuffers *Buffer, pOffsets *DeviceSize)
	vkCmdBindIndexBuffer     func(cmd CommandBuffer, buffer Buffer, offset DeviceSize, indexType uint32)
	vkCmdBindDescriptorSets  func(cmd CommandBuffer, bindPoint uint32, layout PipelineLayout, firstSet, count uint32, pSets *DescriptorSet, dynCount uint32, pDyn *uint32)
	vkCmdPushConstants       func(cmd CommandBuffer, layout PipelineLayout, stage uint32, offset, size uint32, pValues unsafe.Pointer)
	vkCmdDraw                func(cmd CommandBuffer, vertexCount, instanceCount, firstVertex, firstInstance uint32)
	vkCmdDrawIndexed         func(cmd CommandBuffer, indexCount, instanceCount, firstIndex uint32, vertexOffset int32, firstInstance uint32)
	vkCmdCopyBuffer          func(cmd CommandBuffer, src, dst Buffer, regionCount uint32, pRegions unsafe.Pointer)
)

type bufferCopy struct {
	srcOffset DeviceSize
	dstOffset DeviceSize
	size      DeviceSize
}

func loadCommandCommands(device Device) {
	h := uintptr(device)
	bindDeviceProc(&vkCreateCommandPool, h, "vkCreateCommandPool")
	bindDeviceProc(&vkDestroyCommandPool, h, "vkDestroyCommandPool")
	bindDeviceProc(&vkAllocateCommandBuffers, h, "vkAllocateCommandBuffers")
	bindDeviceProc(&vkBeginCommandBuffer, h, "vkBeginCommandBuffer")
	bindDeviceProc(&vkEndCommandBuffer, h, "vkEndCommandBuffer")
	bindDeviceProc(&vkResetCommandBuffer, h, "vkResetCommandBuffer")
	bindDeviceProc(&vkCmdBeginRenderPass, h, "vkCmdBeginRenderPass")
	bindDeviceProc(&vkCmdEndRenderPass, h, "vkCmdEndRenderPass")
	bindDeviceProc(&vkCmdBindPipeline, h, "vkCmdBindPipeline")
	bindDeviceProc(&vkCmdSetViewport, h, "vkCmdSetViewport")
	bindDeviceProc(&vkCmdSetScissor, h, "vkCmdSetScissor")
	bindDeviceProc(&vkCmdBindVertexBuffers, h, "vkCmdBindVertexBuffers")
	bindDeviceProc(&vkCmdBindIndexBuffer, h, "vkCmdBindIndexBuffer")
	bindDeviceProc(&vkCmdBindDescriptorSets, h, "vkCmdBindDescriptorSets")
	bindDeviceProc(&vkCmdPushConstants, h, "vkCmdPushConstants")
	bindDeviceProc(&vkCmdDraw, h, "vkCmdDraw")
	bindDeviceProc(&vkCmdDrawIndexed, h, "vkCmdDrawIndexed")
	bindDeviceProc(&vkCmdCopyBuffer, h, "vkCmdCopyBuffer")
}

// CopyBuffer records a full-size buffer copy of size bytes.
func (c CommandBuffer) CopyBuffer(src, dst Buffer, size DeviceSize) {
	region := bufferCopy{size: size}
	vkCmdCopyBuffer(c, src, dst, 1, unsafe.Pointer(&region))
	runtime.KeepAlive(&region)
}

// CreateCommandPool creates a command pool allowing individual buffer resets.
func (d Device) CreateCommandPool(family uint32) (CommandPool, error) {
	ci := commandPoolCreateInfo{
		sType:            stCommandPoolCreateInfo,
		flags:            CommandPoolResetCommandBuffer,
		queueFamilyIndex: family,
	}
	var pool CommandPool
	res := vkCreateCommandPool(d, unsafe.Pointer(&ci), nil, &pool)
	runtime.KeepAlive(&ci)
	return pool, res.asError("vkCreateCommandPool")
}

// DestroyCommandPool destroys a command pool and its buffers.
func (d Device) DestroyCommandPool(pool CommandPool) {
	if pool != 0 {
		vkDestroyCommandPool(d, pool, nil)
	}
}

// AllocateCommandBuffers allocates count primary command buffers.
func (d Device) AllocateCommandBuffers(pool CommandPool, count uint32) ([]CommandBuffer, error) {
	ai := commandBufferAllocateInfo{
		sType:              stCommandBufferAllocateInfo,
		commandPool:        pool,
		level:              commandBufferLevelPrimary,
		commandBufferCount: count,
	}
	buffers := make([]CommandBuffer, count)
	res := vkAllocateCommandBuffers(d, unsafe.Pointer(&ai), &buffers[0])
	runtime.KeepAlive(&ai)
	return buffers, res.asError("vkAllocateCommandBuffers")
}

// Begin starts recording. flags is a VkCommandBufferUsageFlags value.
func (c CommandBuffer) Begin(flags uint32) error {
	bi := commandBufferBeginInfo{sType: stCommandBufferBeginInfo, flags: flags}
	res := vkBeginCommandBuffer(c, unsafe.Pointer(&bi))
	runtime.KeepAlive(&bi)
	return res.asError("vkBeginCommandBuffer")
}

// End finishes recording.
func (c CommandBuffer) End() error { return vkEndCommandBuffer(c).asError("vkEndCommandBuffer") }

// Reset resets the command buffer.
func (c CommandBuffer) Reset() error { return vkResetCommandBuffer(c, 0).asError("vkResetCommandBuffer") }

// BeginRenderPass begins a render pass with inline contents. clears holds one
// ClearValue per attachment (color then depth).
func (c CommandBuffer) BeginRenderPass(rp RenderPass, fb Framebuffer, area Rect2D, clears []ClearValue) {
	bi := renderPassBeginInfo{
		sType:           stRenderPassBeginInfo,
		renderPass:      rp,
		framebuffer:     fb,
		renderArea:      area,
		clearValueCount: uint32(len(clears)),
	}
	if len(clears) > 0 {
		bi.pClearValues = &clears[0]
	}
	vkCmdBeginRenderPass(c, unsafe.Pointer(&bi), SubpassContentsInline)
	runtime.KeepAlive(&bi)
	runtime.KeepAlive(clears)
}

// EndRenderPass ends the current render pass.
func (c CommandBuffer) EndRenderPass() { vkCmdEndRenderPass(c) }

// BindPipeline binds a graphics pipeline.
func (c CommandBuffer) BindPipeline(p Pipeline) { vkCmdBindPipeline(c, BindPointGraphics, p) }

// SetViewport sets a single viewport.
func (c CommandBuffer) SetViewport(v Viewport) { vkCmdSetViewport(c, 0, 1, &v) }

// SetScissor sets a single scissor rectangle.
func (c CommandBuffer) SetScissor(r Rect2D) { vkCmdSetScissor(c, 0, 1, &r) }

// BindVertexBuffer binds one vertex buffer at binding 0 with the given offset.
func (c CommandBuffer) BindVertexBuffer(b Buffer, offset DeviceSize) {
	vkCmdBindVertexBuffers(c, 0, 1, &b, &offset)
}

// BindVertexBuffers binds buffers starting at firstBinding. The offsets slice
// must match buffers in length.
func (c CommandBuffer) BindVertexBuffers(firstBinding uint32, buffers []Buffer, offsets []DeviceSize) {
	vkCmdBindVertexBuffers(c, firstBinding, uint32(len(buffers)), &buffers[0], &offsets[0])
	runtime.KeepAlive(buffers)
	runtime.KeepAlive(offsets)
}

// BindIndexBuffer binds an index buffer.
func (c CommandBuffer) BindIndexBuffer(b Buffer, offset DeviceSize, indexType uint32) {
	vkCmdBindIndexBuffer(c, b, offset, indexType)
}

// BindDescriptorSet binds a single descriptor set at firstSet.
func (c CommandBuffer) BindDescriptorSet(layout PipelineLayout, firstSet uint32, set DescriptorSet) {
	vkCmdBindDescriptorSets(c, BindPointGraphics, layout, firstSet, 1, &set, 0, nil)
}

// PushConstants uploads push constant data.
func (c CommandBuffer) PushConstants(layout PipelineLayout, stage, offset uint32, data unsafe.Pointer, size uint32) {
	vkCmdPushConstants(c, layout, stage, offset, size, data)
}

// Draw issues a non-indexed draw.
func (c CommandBuffer) Draw(vertexCount, instanceCount, firstVertex, firstInstance uint32) {
	vkCmdDraw(c, vertexCount, instanceCount, firstVertex, firstInstance)
}

// DrawIndexed issues an indexed draw.
func (c CommandBuffer) DrawIndexed(indexCount, instanceCount, firstIndex uint32, vertexOffset int32, firstInstance uint32) {
	vkCmdDrawIndexed(c, indexCount, instanceCount, firstIndex, vertexOffset, firstInstance)
}
