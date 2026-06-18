package vk

import (
	"runtime"
	"unsafe"

	vulkan "github.com/christerso/vulkan-go/vulkan"
)

// CreateSemaphore creates a binary semaphore.
func (d Device) CreateSemaphore() (Semaphore, error) {
	ci := vulkan.VkSemaphoreCreateInfo{SType: vulkan.VkStructureType(stSemaphoreCreateInfo)}
	var s vulkan.VkSemaphore
	res := Result(vulkan.VkCreateSemaphore(vulkan.VkDevice(d), unsafe.Pointer(&ci), nil, unsafe.Pointer(&s)))
	runtime.KeepAlive(&ci)
	return Semaphore(s), res.asError("vkCreateSemaphore")
}

// DestroySemaphore destroys a semaphore.
func (d Device) DestroySemaphore(s Semaphore) {
	if s != 0 {
		vulkan.VkDestroySemaphore(vulkan.VkDevice(d), vulkan.VkSemaphore(s), nil)
	}
}

// CreateFence creates a fence. If signaled is true it starts signaled.
func (d Device) CreateFence(signaled bool) (Fence, error) {
	ci := vulkan.VkFenceCreateInfo{SType: vulkan.VkStructureType(stFenceCreateInfo)}
	if signaled {
		ci.Flags = FenceCreateSignaled
	}
	var f vulkan.VkFence
	res := Result(vulkan.VkCreateFence(vulkan.VkDevice(d), unsafe.Pointer(&ci), nil, unsafe.Pointer(&f)))
	runtime.KeepAlive(&ci)
	return Fence(f), res.asError("vkCreateFence")
}

// DestroyFence destroys a fence.
func (d Device) DestroyFence(f Fence) {
	if f != 0 {
		vulkan.VkDestroyFence(vulkan.VkDevice(d), vulkan.VkFence(f), nil)
	}
}

// WaitFence waits for a single fence with the given timeout in nanoseconds.
func (d Device) WaitFence(f Fence, timeout uint64) error {
	vf := vulkan.VkFence(f)
	res := Result(vulkan.VkWaitForFences(vulkan.VkDevice(d), 1, unsafe.Pointer(&vf), 1, timeout))
	runtime.KeepAlive(&vf)
	return res.asError("vkWaitForFences")
}

// ResetFence resets a single fence.
func (d Device) ResetFence(f Fence) error {
	vf := vulkan.VkFence(f)
	res := Result(vulkan.VkResetFences(vulkan.VkDevice(d), 1, unsafe.Pointer(&vf)))
	runtime.KeepAlive(&vf)
	return res.asError("vkResetFences")
}

// WaitIdle blocks until the device is idle.
func (d Device) WaitIdle() error {
	return Result(vulkan.VkDeviceWaitIdle(vulkan.VkDevice(d))).asError("vkDeviceWaitIdle")
}

// WaitIdle blocks until the queue is idle.
func (q Queue) WaitIdle() error {
	return Result(vulkan.VkQueueWaitIdle(vulkan.VkQueue(q))).asError("vkQueueWaitIdle")
}

// SubmitConfig describes a single queue submission.
type SubmitConfig struct {
	Wait      Semaphore
	WaitStage uint32
	Command   CommandBuffer
	Signal    Semaphore
	Fence     Fence
}

// Submit submits one command buffer with one optional wait and signal semaphore.
func (q Queue) Submit(cfg SubmitConfig) error {
	cmd := vulkan.VkCommandBuffer(cfg.Command)
	si := vulkan.VkSubmitInfo{
		SType:              vulkan.VkStructureType(stSubmitInfo),
		CommandBufferCount: 1,
		PCommandBuffers:    unsafe.Pointer(&cmd),
	}
	wait := vulkan.VkSemaphore(cfg.Wait)
	stage := cfg.WaitStage
	if cfg.Wait != 0 {
		si.WaitSemaphoreCount = 1
		si.PWaitSemaphores = unsafe.Pointer(&wait)
		si.PWaitDstStageMask = unsafe.Pointer(&stage)
	}
	signal := vulkan.VkSemaphore(cfg.Signal)
	if cfg.Signal != 0 {
		si.SignalSemaphoreCount = 1
		si.PSignalSemaphores = unsafe.Pointer(&signal)
	}
	res := Result(vulkan.VkQueueSubmit(vulkan.VkQueue(q), 1, unsafe.Pointer(&si), vulkan.VkFence(cfg.Fence)))
	runtime.KeepAlive(&si)
	runtime.KeepAlive(&cmd)
	runtime.KeepAlive(&wait)
	runtime.KeepAlive(&stage)
	runtime.KeepAlive(&signal)
	return res.asError("vkQueueSubmit")
}
