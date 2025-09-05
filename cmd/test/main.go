package main

import (
	"fmt"
	"log"
	"unsafe"

	"github.com/christerso/vulkan-go/pkg/vulkan"
)

func main() {
	fmt.Println("🚀 Starting Real Vulkan Test")
	fmt.Println("💻 Testing actual Vulkan API calls with your GPU...")

	// Initialize Vulkan
	if err := vulkan.Init(); err != nil {
		log.Fatal("Failed to initialize Vulkan:", err)
	}
	defer vulkan.Destroy()

	// Print Vulkan version
	version := vulkan.GetVersion()
	major := version >> 22
	minor := (version >> 12) & 0x3FF
	patch := version & 0xFFF
	fmt.Printf("📋 Vulkan API Version: %d.%d.%d\n", major, minor, patch)

	// Create application info
	appName := vulkan.CString("Real Vulkan Go Test")
	engineName := vulkan.CString("Vulkan-Go Engine")
	defer vulkan.FreeCString(appName)
	defer vulkan.FreeCString(engineName)

	appInfo := vulkan.ApplicationInfo{
		PApplicationName:   appName,
		ApplicationVersion: 1<<22 | 0<<12 | 0, // Version 1.0.0
		PEngineName:       engineName,
		EngineVersion:     1<<22 | 0<<12 | 0, // Version 1.0.0
		ApiVersion:        uint32(version),
	}

	// Create instance
	createInfo := vulkan.InstanceCreateInfo{
		PApplicationInfo:        &appInfo,
		EnabledLayerCount:       0,
		PpEnabledLayerNames:     nil,
		EnabledExtensionCount:   0,
		PpEnabledExtensionNames: nil,
	}

	var instance vulkan.Instance
	result := vulkan.CreateInstance(&createInfo, nil, &instance)
	if result != vulkan.SUCCESS {
		log.Fatalf("❌ Failed to create Vulkan instance: %v", result)
	}
	defer vulkan.DestroyInstance(instance, nil)

	fmt.Printf("✅ Vulkan instance created successfully!\n")

	// Enumerate physical devices
	var deviceCount uint32
	result = vulkan.EnumeratePhysicalDevices(instance, &deviceCount, nil)
	if result != vulkan.SUCCESS {
		log.Fatalf("❌ Failed to enumerate physical devices: %v", result)
	}

	fmt.Printf("🎮 Found %d physical device(s)\n", deviceCount)

	if deviceCount == 0 {
		log.Fatal("❌ No Vulkan-capable devices found!")
	}

	// Get device handles
	devices := make([]vulkan.PhysicalDevice, deviceCount)
	result = vulkan.EnumeratePhysicalDevices(instance, &deviceCount, &devices[0])
	if result != vulkan.SUCCESS {
		log.Fatalf("❌ Failed to get physical devices: %v", result)
	}

	fmt.Printf("🔥 Successfully enumerated %d Vulkan device(s):\n", len(devices))
	for i, device := range devices {
		fmt.Printf("   Device %d: %p\n", i, unsafe.Pointer(device))
	}

	fmt.Println("\n🎉 Real Vulkan API test completed successfully!")
	fmt.Println("✅ Your system has working Vulkan drivers and hardware")
	fmt.Println("🚀 This proves the Vulkan-Go wrapper is working with actual Vulkan!")
}