package vk

import (
	"fmt"
	"github.com/christerso/vulkan-go/pkg/vulkan"
	"sync"
	"unsafe"
)

// MemoryAllocator provides high-level memory allocation and management
type MemoryAllocator struct {
	device         *LogicalDevice
	allocations    map[vulkan.DeviceSize]*MemoryAllocation
	pools          []*MemoryPool
	mutex          sync.RWMutex
	totalAllocated vulkan.DeviceSize
	maxAllocations uint32
}

// MemoryAllocation represents a single memory allocation
type MemoryAllocation struct {
	Memory     vulkan.DeviceSize // Placeholder for VkDeviceMemory handle
	Size       vulkan.DeviceSize
	TypeIndex  uint32
	Properties MemoryPropertyFlags
	Mapped     unsafe.Pointer
	RefCount   int32
	Pool       *MemoryPool
}

// MemoryPool manages a pool of memory allocations for efficiency
type MemoryPool struct {
	Memory        vulkan.DeviceSize // Placeholder for VkDeviceMemory handle
	Size          vulkan.DeviceSize
	TypeIndex     uint32
	Properties    MemoryPropertyFlags
	BlockSize     vulkan.DeviceSize
	FreeBlocks    []MemoryBlock
	UsedBlocks    []MemoryBlock
	mutex         sync.Mutex
}

// MemoryBlock represents a block within a memory pool
type MemoryBlock struct {
	Offset vulkan.DeviceSize
	Size   vulkan.DeviceSize
	InUse  bool
}

// AllocationCreateInfo specifies parameters for memory allocation
type AllocationCreateInfo struct {
	Usage             MemoryUsage
	RequiredFlags     MemoryPropertyFlags
	PreferredFlags    MemoryPropertyFlags
	Pool              *MemoryPool
	UserData          interface{}
}

// MemoryUsage defines how the memory will be used
type MemoryUsage uint32

const (
	MemoryUsageUnknown MemoryUsage = iota
	MemoryUsageGPUOnly              // Device local memory
	MemoryUsageCPUOnly              // Host visible, host coherent
	MemoryUsageCPUToGPU             // Host visible, device local preferred
	MemoryUsageGPUToCPU             // Host visible, host cached preferred
	MemoryUsageCPUCopy              // Host visible, host coherent, temporary
	MemoryUsageGPULazilyAllocated   // Device local, lazily allocated
)

// MemoryRequirements represents Vulkan memory requirements
type MemoryRequirements struct {
	Size           vulkan.DeviceSize
	Alignment      vulkan.DeviceSize
	MemoryTypeBits uint32
}

// NewMemoryAllocator creates a new memory allocator
func NewMemoryAllocator(device *LogicalDevice) *MemoryAllocator {
	return &MemoryAllocator{
		device:         device,
		allocations:    make(map[vulkan.DeviceSize]*MemoryAllocation),
		pools:          make([]*MemoryPool, 0),
		maxAllocations: 4096,
	}
}

// Destroy cleans up the memory allocator and all its allocations
func (ma *MemoryAllocator) Destroy() {
	ma.mutex.Lock()
	defer ma.mutex.Unlock()

	// Free all allocations
	for _, alloc := range ma.allocations {
		ma.freeAllocationUnsafe(alloc)
	}
	ma.allocations = nil

	// Destroy all pools
	for _, pool := range ma.pools {
		pool.Destroy()
	}
	ma.pools = nil
}

// Allocate allocates memory with the specified requirements
func (ma *MemoryAllocator) Allocate(requirements MemoryRequirements, createInfo AllocationCreateInfo) (*MemoryAllocation, error) {
	ma.mutex.Lock()
	defer ma.mutex.Unlock()

	// Check allocation limits
	if uint32(len(ma.allocations)) >= ma.maxAllocations {
		return nil, fmt.Errorf("maximum number of allocations (%d) reached", ma.maxAllocations)
	}

	// If a specific pool is requested, try to allocate from it
	if createInfo.Pool != nil {
		alloc, err := createInfo.Pool.Allocate(requirements.Size)
		if err == nil {
			ma.allocations[vulkan.DeviceSize(uintptr(unsafe.Pointer(alloc)))] = alloc
			ma.totalAllocated += alloc.Size
			return alloc, nil
		}
	}

	// Find suitable memory type
	memoryType, err := ma.findMemoryType(requirements, createInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to find suitable memory type: %w", err)
	}

	// Try to allocate from existing pools first
	for _, pool := range ma.pools {
		if pool.TypeIndex == memoryType {
			alloc, err := pool.Allocate(requirements.Size)
			if err == nil {
				ma.allocations[vulkan.DeviceSize(uintptr(unsafe.Pointer(alloc)))] = alloc
				ma.totalAllocated += alloc.Size
				return alloc, nil
			}
		}
	}

	// Create new allocation directly (for large allocations)
	if requirements.Size > 64*1024*1024 { // 64MB threshold
		alloc, err := ma.allocateDirect(requirements.Size, memoryType)
		if err != nil {
			return nil, err
		}
		ma.allocations[vulkan.DeviceSize(uintptr(unsafe.Pointer(alloc)))] = alloc
		ma.totalAllocated += alloc.Size
		return alloc, nil
	}

	// Create new pool for smaller allocations
	pool, err := ma.createPool(memoryType, 256*1024*1024) // 256MB pool
	if err != nil {
		return nil, fmt.Errorf("failed to create memory pool: %w", err)
	}

	ma.pools = append(ma.pools, pool)

	alloc, err := pool.Allocate(requirements.Size)
	if err != nil {
		return nil, err
	}

	ma.allocations[vulkan.DeviceSize(uintptr(unsafe.Pointer(alloc)))] = alloc
	ma.totalAllocated += alloc.Size
	return alloc, nil
}

// Free releases a memory allocation
func (ma *MemoryAllocator) Free(allocation *MemoryAllocation) error {
	ma.mutex.Lock()
	defer ma.mutex.Unlock()

	return ma.freeAllocationUnsafe(allocation)
}

// Map maps memory allocation to CPU accessible pointer
func (ma *MemoryAllocator) Map(allocation *MemoryAllocation) (unsafe.Pointer, error) {
	if allocation.Mapped != nil {
		return allocation.Mapped, nil
	}

	// Check if memory type supports mapping
	memProps := ma.device.GetPhysicalDevice().GetMemoryProperties()
	memType := memProps.MemoryTypes[allocation.TypeIndex]
	
	if memType.PropertyFlags&MemoryPropertyHostVisibleBit == 0 {
		return nil, fmt.Errorf("memory type does not support host mapping")
	}

	// TODO: Call vkMapMemory
	// For now, return a placeholder
	allocation.Mapped = unsafe.Pointer(uintptr(1)) // Placeholder
	return allocation.Mapped, nil
}

// Unmap unmaps a previously mapped memory allocation
func (ma *MemoryAllocator) Unmap(allocation *MemoryAllocation) {
	if allocation.Mapped != nil {
		// TODO: Call vkUnmapMemory
		allocation.Mapped = nil
	}
}

// GetStats returns memory allocation statistics
func (ma *MemoryAllocator) GetStats() MemoryStats {
	ma.mutex.RLock()
	defer ma.mutex.RUnlock()

	stats := MemoryStats{
		TotalAllocated:   ma.totalAllocated,
		AllocationCount:  uint32(len(ma.allocations)),
		PoolCount:        uint32(len(ma.pools)),
		MaxAllocations:   ma.maxAllocations,
	}

	for _, pool := range ma.pools {
		stats.PoolStats = append(stats.PoolStats, pool.GetStats())
	}

	return stats
}

// MemoryStats provides statistics about memory usage
type MemoryStats struct {
	TotalAllocated  vulkan.DeviceSize
	AllocationCount uint32
	PoolCount       uint32
	MaxAllocations  uint32
	PoolStats       []PoolStats
}

// PoolStats provides statistics about a memory pool
type PoolStats struct {
	TotalSize   vulkan.DeviceSize
	UsedSize    vulkan.DeviceSize
	FreeSize    vulkan.DeviceSize
	BlockCount  uint32
	TypeIndex   uint32
	Properties  MemoryPropertyFlags
}

// Helper methods

func (ma *MemoryAllocator) findMemoryType(requirements MemoryRequirements, createInfo AllocationCreateInfo) (uint32, error) {
	memProps := ma.device.GetPhysicalDevice().GetMemoryProperties()

	// Convert usage to property flags
	requiredFlags := createInfo.RequiredFlags
	preferredFlags := createInfo.PreferredFlags

	switch createInfo.Usage {
	case MemoryUsageGPUOnly:
		requiredFlags |= MemoryPropertyDeviceLocalBit
	case MemoryUsageCPUOnly:
		requiredFlags |= MemoryPropertyHostVisibleBit | MemoryPropertyHostCoherentBit
	case MemoryUsageCPUToGPU:
		requiredFlags |= MemoryPropertyHostVisibleBit
		preferredFlags |= MemoryPropertyDeviceLocalBit
	case MemoryUsageGPUToCPU:
		requiredFlags |= MemoryPropertyHostVisibleBit
		preferredFlags |= MemoryPropertyHostCachedBit
	case MemoryUsageCPUCopy:
		requiredFlags |= MemoryPropertyHostVisibleBit | MemoryPropertyHostCoherentBit
	case MemoryUsageGPULazilyAllocated:
		requiredFlags |= MemoryPropertyDeviceLocalBit
		preferredFlags |= MemoryPropertyLazilyAllocatedBit
	}

	// First pass: try to find memory type with all preferred flags
	for i := uint32(0); i < memProps.MemoryTypeCount; i++ {
		if (requirements.MemoryTypeBits&(1<<i)) != 0 {
			memType := memProps.MemoryTypes[i]
			if (memType.PropertyFlags&requiredFlags) == requiredFlags &&
			   (memType.PropertyFlags&preferredFlags) == preferredFlags {
				return i, nil
			}
		}
	}

	// Second pass: find memory type with just required flags
	for i := uint32(0); i < memProps.MemoryTypeCount; i++ {
		if (requirements.MemoryTypeBits&(1<<i)) != 0 {
			memType := memProps.MemoryTypes[i]
			if (memType.PropertyFlags&requiredFlags) == requiredFlags {
				return i, nil
			}
		}
	}

	return 0, fmt.Errorf("no suitable memory type found")
}

func (ma *MemoryAllocator) allocateDirect(size vulkan.DeviceSize, typeIndex uint32) (*MemoryAllocation, error) {
	// TODO: Implement actual VkDeviceMemory allocation
	allocation := &MemoryAllocation{
		Memory:     vulkan.DeviceSize(size), // Placeholder
		Size:       size,
		TypeIndex:  typeIndex,
		Properties: ma.device.GetPhysicalDevice().GetMemoryProperties().MemoryTypes[typeIndex].PropertyFlags,
		RefCount:   1,
	}

	return allocation, nil
}

func (ma *MemoryAllocator) createPool(typeIndex uint32, size vulkan.DeviceSize) (*MemoryPool, error) {
	// TODO: Implement actual VkDeviceMemory allocation for pool
	pool := &MemoryPool{
		Memory:     vulkan.DeviceSize(size), // Placeholder
		Size:       size,
		TypeIndex:  typeIndex,
		Properties: ma.device.GetPhysicalDevice().GetMemoryProperties().MemoryTypes[typeIndex].PropertyFlags,
		BlockSize:  64 * 1024, // 64KB blocks
		FreeBlocks: []MemoryBlock{{Offset: 0, Size: size, InUse: false}},
		UsedBlocks: []MemoryBlock{},
	}

	return pool, nil
}

func (ma *MemoryAllocator) freeAllocationUnsafe(allocation *MemoryAllocation) error {
	if allocation == nil {
		return fmt.Errorf("cannot free nil allocation")
	}

	// Unmap if mapped
	if allocation.Mapped != nil {
		ma.Unmap(allocation)
	}

	// Remove from tracking
	key := vulkan.DeviceSize(uintptr(unsafe.Pointer(allocation)))
	if _, exists := ma.allocations[key]; !exists {
		return fmt.Errorf("allocation not found in allocator")
	}

	delete(ma.allocations, key)
	ma.totalAllocated -= allocation.Size

	// Free from pool or direct allocation
	if allocation.Pool != nil {
		return allocation.Pool.Free(allocation)
	} else {
		// TODO: Call vkFreeMemory for direct allocation
		return nil
	}
}

// Memory pool methods

// Allocate allocates memory from the pool
func (mp *MemoryPool) Allocate(size vulkan.DeviceSize) (*MemoryAllocation, error) {
	mp.mutex.Lock()
	defer mp.mutex.Unlock()

	// Align size to block boundaries
	alignedSize := (size + mp.BlockSize - 1) &^ (mp.BlockSize - 1)

	// Find a suitable free block
	for i, block := range mp.FreeBlocks {
		if block.Size >= alignedSize {
			// Split the block if necessary
			if block.Size > alignedSize {
				// Create new free block for remaining space
				newFreeBlock := MemoryBlock{
					Offset: block.Offset + alignedSize,
					Size:   block.Size - alignedSize,
					InUse:  false,
				}
				mp.FreeBlocks = append(mp.FreeBlocks, newFreeBlock)
			}

			// Create used block
			usedBlock := MemoryBlock{
				Offset: block.Offset,
				Size:   alignedSize,
				InUse:  true,
			}
			mp.UsedBlocks = append(mp.UsedBlocks, usedBlock)

			// Remove or modify the free block
			if i < len(mp.FreeBlocks)-1 {
				mp.FreeBlocks[i] = mp.FreeBlocks[len(mp.FreeBlocks)-1]
			}
			mp.FreeBlocks = mp.FreeBlocks[:len(mp.FreeBlocks)-1]

			// Create allocation
			allocation := &MemoryAllocation{
				Memory:     mp.Memory,
				Size:       alignedSize,
				TypeIndex:  mp.TypeIndex,
				Properties: mp.Properties,
				Pool:       mp,
				RefCount:   1,
			}

			return allocation, nil
		}
	}

	return nil, fmt.Errorf("insufficient space in pool")
}

// Free releases memory back to the pool
func (mp *MemoryPool) Free(allocation *MemoryAllocation) error {
	mp.mutex.Lock()
	defer mp.mutex.Unlock()

	// Find and remove the used block
	var freedBlock MemoryBlock
	found := false

	for i, block := range mp.UsedBlocks {
		if block.Offset == 0 { // TODO: Proper offset matching
			freedBlock = block
			mp.UsedBlocks = append(mp.UsedBlocks[:i], mp.UsedBlocks[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("allocation not found in pool")
	}

	// Add back to free blocks
	mp.FreeBlocks = append(mp.FreeBlocks, MemoryBlock{
		Offset: freedBlock.Offset,
		Size:   freedBlock.Size,
		InUse:  false,
	})

	// TODO: Coalesce adjacent free blocks for efficiency

	return nil
}

// GetStats returns statistics about the pool
func (mp *MemoryPool) GetStats() PoolStats {
	mp.mutex.Lock()
	defer mp.mutex.Unlock()

	var usedSize vulkan.DeviceSize
	for _, block := range mp.UsedBlocks {
		usedSize += block.Size
	}

	return PoolStats{
		TotalSize:  mp.Size,
		UsedSize:   usedSize,
		FreeSize:   mp.Size - usedSize,
		BlockCount: uint32(len(mp.UsedBlocks) + len(mp.FreeBlocks)),
		TypeIndex:  mp.TypeIndex,
		Properties: mp.Properties,
	}
}

// Destroy cleans up the memory pool
func (mp *MemoryPool) Destroy() {
	mp.mutex.Lock()
	defer mp.mutex.Unlock()

	// TODO: Call vkFreeMemory
	mp.FreeBlocks = nil
	mp.UsedBlocks = nil
}

// Utility functions

// AlignUp aligns a size up to the specified alignment
func AlignUp(size, alignment vulkan.DeviceSize) vulkan.DeviceSize {
	return (size + alignment - 1) &^ (alignment - 1)
}

// AlignDown aligns a size down to the specified alignment
func AlignDown(size, alignment vulkan.DeviceSize) vulkan.DeviceSize {
	return size &^ (alignment - 1)
}

// IsAligned checks if a size is aligned to the specified alignment
func IsAligned(size, alignment vulkan.DeviceSize) bool {
	return size&(alignment-1) == 0
}