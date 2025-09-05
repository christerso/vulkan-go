package main

import (
	"fmt"
	"log"
	"math"
	"time"

	"github.com/christerso/vulkan-go/pkg/vk"
	"github.com/christerso/vulkan-go/pkg/vulkan"
)

const AppName = "Vulkan-Go Demo"

// Vertex represents a triangle vertex with position and color
type Vertex struct {
	Position [2]float32 // X, Y
	Color    [3]float32 // R, G, B
}

// Triangle vertices with animated colors
var triangleVertices = []Vertex{
	// Top vertex (red-ish)
	{Position: [2]float32{0.0, -0.6}, Color: [3]float32{1.0, 0.3, 0.3}},
	// Bottom right vertex (green-ish)
	{Position: [2]float32{0.6, 0.6}, Color: [3]float32{0.3, 1.0, 0.3}},
	// Bottom left vertex (blue-ish)
	{Position: [2]float32{-0.6, 0.6}, Color: [3]float32{0.3, 0.3, 1.0}},
}

// DemoRenderer simulates Vulkan rendering for demonstration
type DemoRenderer struct {
	startTime  time.Time
	frameCount uint64
	running    bool
}

func NewDemoRenderer() *DemoRenderer {
	return &DemoRenderer{
		startTime: time.Now(),
		running:   true,
	}
}

func main() {
	fmt.Printf("🚀 Starting %s\n", AppName)
	fmt.Printf("📦 This is a demonstration of the Vulkan-Go wrapper\n")
	fmt.Printf("⚠️  Note: This demo runs without requiring Vulkan drivers\n\n")

	renderer := NewDemoRenderer()
	
	if err := renderer.Initialize(); err != nil {
		log.Fatal("Failed to initialize renderer:", err)
	}
	defer renderer.Cleanup()
	
	fmt.Printf("🎮 Running fancy triangle animation...\n")
	fmt.Printf("🔄 Press Ctrl+C to stop\n\n")
	
	// Main demo loop
	for renderer.ShouldRun() {
		if err := renderer.DrawFrame(); err != nil {
			log.Printf("Draw frame error: %v", err)
			break
		}
		
		renderer.frameCount++
		
		// Print FPS every 120 frames (2 seconds at 60 FPS)
		if renderer.frameCount%120 == 0 {
			elapsed := time.Since(renderer.startTime)
			fps := float64(renderer.frameCount) / elapsed.Seconds()
			fmt.Printf("📊 FPS: %.1f | Frames: %d | Time: %.1fs\n", 
				fps, renderer.frameCount, elapsed.Seconds())
		}
	}
	
	fmt.Printf("\n✅ Demo completed successfully!\n")
}

func (dr *DemoRenderer) Initialize() error {
	fmt.Printf("🔧 Initializing Vulkan-Go wrapper...\n")
	
	// Test the low-level Vulkan bindings
	if err := vulkan.Init(); err != nil {
		return fmt.Errorf("failed to initialize Vulkan: %w", err)
	}
	
	// Test the high-level wrapper
	config := vk.DefaultInstanceConfig()
	config.ApplicationName = AppName
	config.EnableValidation = false // Disable for demo
	
	// In a real implementation, this would create a Vulkan instance
	fmt.Printf("✅ Instance configuration created: %s v%s\n", 
		config.ApplicationName, config.ApplicationVersion.String())
	
	// Simulate device selection
	fmt.Printf("🎯 Mock GPU selected: High-Performance Graphics Device\n")
	fmt.Printf("💾 Mock memory allocator initialized\n")
	
	// Simulate pipeline creation
	fmt.Printf("🔨 Graphics pipeline created for triangle rendering\n")
	fmt.Printf("🎨 Shaders loaded: vertex + fragment with psychedelic effects\n")
	
	return nil
}

func (dr *DemoRenderer) DrawFrame() error {
	elapsed := time.Since(dr.startTime).Seconds()
	
	// Animate triangle colors using sine waves
	animateTriangleColors(float32(elapsed))
	
	// Simulate rendering work
	dr.simulateRendering()
	
	// Simulate 60 FPS
	time.Sleep(16 * time.Millisecond)
	
	return nil
}

func animateTriangleColors(time float32) {
	// Create psychedelic color animation
	colorPhase := time * 2.0
	
	// Vertex 0: Red wave
	triangleVertices[0].Color = [3]float32{
		0.5 + 0.5*float32(math.Sin(float64(colorPhase))),
		0.5 + 0.5*float32(math.Sin(float64(colorPhase+2.0))),
		0.5 + 0.5*float32(math.Sin(float64(colorPhase+4.0))),
	}
	
	// Vertex 1: Green wave
	triangleVertices[1].Color = [3]float32{
		0.5 + 0.5*float32(math.Sin(float64(colorPhase+1.0))),
		0.5 + 0.5*float32(math.Sin(float64(colorPhase+3.0))),
		0.5 + 0.5*float32(math.Sin(float64(colorPhase+5.0))),
	}
	
	// Vertex 2: Blue wave
	triangleVertices[2].Color = [3]float32{
		0.5 + 0.5*float32(math.Sin(float64(colorPhase+2.0))),
		0.5 + 0.5*float32(math.Sin(float64(colorPhase+4.0))),
		0.5 + 0.5*float32(math.Sin(float64(colorPhase+6.0))),
	}
	
	// Rotate triangle
	angle := time * 0.5
	for i := range triangleVertices {
		x, y := triangleVertices[i].Position[0], triangleVertices[i].Position[1]
		cos, sin := float32(math.Cos(float64(angle))), float32(math.Sin(float64(angle)))
		
		// Apply rotation matrix
		triangleVertices[i].Position[0] = x*cos - y*sin
		triangleVertices[i].Position[1] = x*sin + y*cos
	}
}

func (dr *DemoRenderer) simulateRendering() {
	// Simulate GPU work with realistic timing
	
	// Command buffer recording (1ms)
	time.Sleep(1 * time.Millisecond)
	
	// GPU execution simulation (variable timing)
	elapsed := time.Since(dr.startTime).Seconds()
	gpuLoad := 0.5 + 0.3*math.Sin(elapsed*0.5) // Simulate varying GPU load
	gpuTime := time.Duration(gpuLoad * float64(time.Millisecond * 8))
	time.Sleep(gpuTime)
	
	// Present simulation (2ms)
	time.Sleep(2 * time.Millisecond)
}

func (dr *DemoRenderer) ShouldRun() bool {
	// Run for 10 seconds or until stopped
	elapsed := time.Since(dr.startTime)
	return dr.running && elapsed < 10*time.Second
}

func (dr *DemoRenderer) Cleanup() {
	dr.running = false
	
	fmt.Printf("\n🧹 Cleaning up resources...\n")
	
	// Test error handling
	result := vulkan.SUCCESS
	fmt.Printf("✅ Vulkan result: %s (is error: %v)\n", result.Error(), result.IsError())
	
	// Test memory management
	stats := vk.MemoryStats{
		TotalAllocated:  1024 * 1024, // 1MB
		AllocationCount: 42,
		PoolCount:       3,
	}
	fmt.Printf("💾 Mock memory stats: %.2f MB allocated, %d allocations, %d pools\n",
		float64(stats.TotalAllocated)/(1024*1024), stats.AllocationCount, stats.PoolCount)
	
	// Cleanup Vulkan
	vulkan.Destroy()
	
	fmt.Printf("✨ All resources cleaned up\n")
}