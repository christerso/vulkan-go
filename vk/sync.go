package vk

import (
	"runtime"
	"unsafe"
)

type semaphoreCreateInfo struct {
	sType uint32
	pNext unsafe.Pointer
	flags uint32
}

type fenceCreateInfo struct {
	sType uint32
	pNext unsafe.Pointer
	flags uint32
}

type submitInfo struct {
	sType                uint32
	pNext                unsafe.Pointer
	waitSemaphoreCount   uint32
	pWaitSemaphores      *Semaphore
	pWaitDstStageMask    *uint32
	commandBufferCount   uint32
	pCommandBuffers      *CommandBuffer
	signalSemaphoreCount uint32
	pSignalSemaphores    *Semaphore
}

var (
	vkCreateSemaphore  func(device Device, pInfo, pAllocator unsafe.Pointer, pSem *Semaphore) Result
	vkDestroySemaphore func(device Device, sem Semaphore, pAllocator unsafe.Pointer)
	vkCreateFence      func(device Device, pInfo, pAllocator unsafe.Pointer, pFence *Fence) Result
	vkDestroyFence     func(device Device, fence Fence, pAllocator unsafe.Pointer)
	vkWaitForFences    func(device Device, count uint32, pFences *Fence, waitAll uint32, timeout uint64) Result
	vkResetFences      func(device Device, count uint32, pFences *Fence) Result
	vkQueueSubmit      func(queue Queue, count uint32, pSubmits unsafe.Pointer, fence Fence) Result
	vkDeviceWaitIdle   func(device Device) Result
	vkQueueWaitIdle    func(queue Queue) Result
)

func loadSyncCommands(device Device) {
	h := uintptr(device)
	bindDeviceProc(&vkCreateSemaphore, h, "vkCreateSemaphore")
	bindDeviceProc(&vkDestroySemaphore, h, "vkDestroySemaphore")
	bindDeviceProc(&vkCreateFence, h, "vkCreateFence")
	bindDeviceProc(&vkDestroyFence, h, "vkDestroyFence")
	bindDeviceProc(&vkWaitForFences, h, "vkWaitForFences")
	bindDeviceProc(&vkResetFences, h, "vkResetFences")
	bindDeviceProc(&vkQueueSubmit, h, "vkQueueSubmit")
}

// CreateSemaphore creates a binary semaphore.
func (d Device) CreateSemaphore() (Semaphore, error) {
	ci := semaphoreCreateInfo{sType: stSemaphoreCreateInfo}
	var s Semaphore
	res := vkCreateSemaphore(d, unsafe.Pointer(&ci), nil, &s)
	runtime.KeepAlive(&ci)
	return s, res.asError("vkCreateSemaphore")
}

// DestroySemaphore destroys a semaphore.
func (d Device) DestroySemaphore(s Semaphore) {
	if s != 0 {
		vkDestroySemaphore(d, s, nil)
	}
}

// CreateFence creates a fence. If signaled is true it starts signaled.
func (d Device) CreateFence(signaled bool) (Fence, error) {
	ci := fenceCreateInfo{sType: stFenceCreateInfo}
	if signaled {
		ci.flags = FenceCreateSignaled
	}
	var f Fence
	res := vkCreateFence(d, unsafe.Pointer(&ci), nil, &f)
	runtime.KeepAlive(&ci)
	return f, res.asError("vkCreateFence")
}

// DestroyFence destroys a fence.
func (d Device) DestroyFence(f Fence) {
	if f != 0 {
		vkDestroyFence(d, f, nil)
	}
}

// WaitFence waits for a single fence with the given timeout in nanoseconds.
func (d Device) WaitFence(f Fence, timeout uint64) error {
	res := vkWaitForFences(d, 1, &f, 1, timeout)
	return res.asError("vkWaitForFences")
}

// ResetFence resets a single fence.
func (d Device) ResetFence(f Fence) error {
	res := vkResetFences(d, 1, &f)
	return res.asError("vkResetFences")
}

// WaitIdle blocks until the device is idle.
func (d Device) WaitIdle() error { return vkDeviceWaitIdle(d).asError("vkDeviceWaitIdle") }

// WaitIdle blocks until the queue is idle.
func (q Queue) WaitIdle() error { return vkQueueWaitIdle(q).asError("vkQueueWaitIdle") }

// SubmitConfig describes a single queue submission.
type SubmitConfig struct {
	Wait       Semaphore
	WaitStage  uint32
	Command    CommandBuffer
	Signal     Semaphore
	Fence      Fence
}

// Submit submits one command buffer with one optional wait and signal semaphore.
func (q Queue) Submit(cfg SubmitConfig) error {
	cmd := cfg.Command
	si := submitInfo{
		sType:              stSubmitInfo,
		commandBufferCount: 1,
		pCommandBuffers:    &cmd,
	}
	wait := cfg.Wait
	stage := cfg.WaitStage
	if wait != 0 {
		si.waitSemaphoreCount = 1
		si.pWaitSemaphores = &wait
		si.pWaitDstStageMask = &stage
	}
	signal := cfg.Signal
	if signal != 0 {
		si.signalSemaphoreCount = 1
		si.pSignalSemaphores = &signal
	}
	res := vkQueueSubmit(q, 1, unsafe.Pointer(&si), cfg.Fence)
	runtime.KeepAlive(&si)
	runtime.KeepAlive(&cmd)
	runtime.KeepAlive(&wait)
	runtime.KeepAlive(&stage)
	runtime.KeepAlive(&signal)
	return res.asError("vkQueueSubmit")
}
