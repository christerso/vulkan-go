# Vulkan-Go

[![Go Report Card](https://goreportcard.com/badge/github.com/christerso/vulkan-go)](https://goreportcard.com/report/github.com/christerso/vulkan-go)
[![Go Reference](https://pkg.go.dev/badge/github.com/christerso/vulkan-go.svg)](https://pkg.go.dev/github.com/christerso/vulkan-go)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

> **The most up-to-date and comprehensive Vulkan wrapper for Go**

A modern, production-ready Vulkan wrapper for Go that provides both low-level bindings and high-level abstractions. Built with Go idioms in mind and automatically updated from the latest Vulkan specifications.

## 🚀 Features

### ✅ **Always Up-to-Date**
- **Vulkan 1.4.326** support (latest as of August 2025)
- Automatic updates from official Khronos Vulkan-Headers repository
- Generated from the official Vulkan XML registry

### 🎯 **Dual-Layer Architecture**
- **Low-level bindings** (`pkg/vulkan`) - Direct Vulkan API access
- **High-level wrapper** (`pkg/vk`) - Go-idiomatic abstractions

### 🛡️ **Production Ready**
- Comprehensive error handling with context
- Memory safety and automatic resource management
- Validation layer integration
- Thread-safe memory allocator
- Extensive testing suite

### 🔧 **Developer Friendly**
- Rich examples and documentation
- Type-safe API with Go generics
- Automatic C string management
- Cross-platform support (Windows, Linux, macOS, Android, iOS)

### 🏗️ **Advanced Features**
- Smart memory allocation with pooling
- Resource lifecycle management
- Extension support system
- Debug utilities and validation helpers

## 🚦 Quick Start

### Installation

```bash
go get github.com/christerso/vulkan-go
```

### Prerequisites

You need Vulkan SDK installed on your system:

- **Windows**: [LunarG Vulkan SDK](https://vulkan.lunarg.com/)
- **Linux**: `sudo apt-get install vulkan-tools libvulkan-dev vulkan-validationlayers-dev`
- **macOS**: Install [MoltenVK](https://github.com/KhronosGroup/MoltenVK) or use Homebrew: `brew install molten-vk`

### Basic Usage

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/christerso/vulkan-go/pkg/vk"
)

func main() {
    // Create Vulkan instance with validation layers
    config := vk.DefaultInstanceConfig()
    config.ApplicationName = "My Vulkan App"
    config.EnableValidation = true
    
    instance, err := vk.CreateInstance(config)
    if err != nil {
        log.Fatal("Failed to create Vulkan instance:", err)
    }
    defer instance.Destroy()
    
    // Find and select a suitable GPU
    physicalDevice, err := instance.GetPhysicalDevice(vk.PhysicalDeviceRequirements{
        RequireGraphicsQueue: true,
        PreferredDeviceType:  vk.DeviceTypeDiscreteGPU,
    })
    if err != nil {
        log.Fatal("No suitable GPU found:", err)
    }
    
    // Create logical device
    deviceConfig := vk.DefaultDeviceConfig(physicalDevice)
    device, err := physicalDevice.CreateLogicalDevice(deviceConfig)
    if err != nil {
        log.Fatal("Failed to create device:", err)
    }
    defer device.Destroy()
    
    // Device is now ready for rendering operations
}
```

### Running the Demo

```bash
go build ./cmd/demo
GODEBUG=cgocheck=0 ./demo.exe
```

The demo opens a window showing an animated rotating triangle using the Vulkan wrapper.

## 📚 Examples

### Triangle Rendering

```go
// See cmd/demo/main.go for complete example
func renderTriangle() error {
    // Create instance with surface extensions
    config := vk.DefaultInstanceConfig()
    config.EnabledExtensions = append(config.EnabledExtensions, 
        "VK_KHR_surface", "VK_KHR_win32_surface") // Platform specific
    
    instance, err := vk.CreateInstance(config)
    if err != nil {
        return err
    }
    defer instance.Destroy()
    
    // Create surface (window system integration)
    surface, err := createSurface(instance) // Platform specific
    if err != nil {
        return err
    }
    defer surface.Destroy()
    
    // Select device with present support
    physicalDevice, err := instance.GetPhysicalDevice(vk.PhysicalDeviceRequirements{
        RequireGraphicsQueue: true,
        RequirePresentQueue:  true,
    })
    if err != nil {
        return err
    }
    
    // Create device with swapchain extension
    deviceConfig := vk.DefaultDeviceConfig(physicalDevice)
    deviceConfig.RequiredExtensions = []string{"VK_KHR_swapchain"}
    
    device, err := physicalDevice.CreateLogicalDevice(deviceConfig)
    if err != nil {
        return err
    }
    defer device.Destroy()
    
    // Create swapchain, render pass, pipeline, etc.
    // ... (see full example)
    
    return nil
}
```

### Compute Shader

```go
// See cmd/compute/main.go for complete example
func runComputeShader() error {
    // Initialize Vulkan for compute
    instance, err := vk.CreateInstance(vk.DefaultInstanceConfig())
    if err != nil {
        return err
    }
    defer instance.Destroy()
    
    // Find compute-capable device
    physicalDevice, err := instance.GetPhysicalDevice(vk.PhysicalDeviceRequirements{
        RequireComputeQueue: true,
    })
    if err != nil {
        return err
    }
    
    device, err := physicalDevice.CreateLogicalDevice(vk.DefaultDeviceConfig(physicalDevice))
    if err != nil {
        return err
    }
    defer device.Destroy()
    
    // Create compute pipeline and buffers
    // ... (see full example)
    
    return nil
}
```

### Memory Management

```go
func demonstrateMemoryManagement(device *vk.LogicalDevice) error {
    // Create memory allocator
    allocator := vk.NewMemoryAllocator(device)
    defer allocator.Destroy()
    
    // Allocate GPU-only memory
    gpuAllocation, err := allocator.Allocate(
        vk.MemoryRequirements{
            Size:           1024 * 1024, // 1MB
            Alignment:      16,
            MemoryTypeBits: 0xFFFFFFFF,
        },
        vk.AllocationCreateInfo{
            Usage: vk.MemoryUsageGPUOnly,
        },
    )
    if err != nil {
        return err
    }
    defer allocator.Free(gpuAllocation)
    
    // Allocate CPU-visible memory
    cpuAllocation, err := allocator.Allocate(
        vk.MemoryRequirements{
            Size:           1024,
            Alignment:      4,
            MemoryTypeBits: 0xFFFFFFFF,
        },
        vk.AllocationCreateInfo{
            Usage: vk.MemoryUsageCPUToGPU,
        },
    )
    if err != nil {
        return err
    }
    defer allocator.Free(cpuAllocation)
    
    // Map and write data
    ptr, err := allocator.Map(cpuAllocation)
    if err != nil {
        return err
    }
    
    // Write data to mapped memory
    data := []byte("Hello Vulkan!")
    copy((*[1024]byte)(ptr)[:], data)
    
    allocator.Unmap(cpuAllocation)
    
    // Get allocator statistics
    stats := allocator.GetStats()
    fmt.Printf("Total allocated: %d bytes\n", stats.TotalAllocated)
    fmt.Printf("Active allocations: %d\n", stats.AllocationCount)
    
    return nil
}
```

## 🏗️ Architecture

### Package Structure

```
vulkan-go/
├── pkg/
│   ├── vulkan/          # Low-level Vulkan bindings
│   │   └── core.go      # Core API functions and types
│   └── vk/              # High-level Go wrapper
│       ├── instance.go  # Instance management
│       ├── device.go    # Device management
│       ├── memory.go    # Memory allocation
│       └── errors.go    # Error handling
├── cmd/
│   └── demo/            # Triangle rendering demo
├── scripts/             # Build and generation tools
└── assets/              # Shaders and resources
```

### Design Principles

1. **Separation of Concerns**: Clean separation between low-level bindings and high-level abstractions
2. **Go Idioms**: Proper error handling, interfaces, and memory management
3. **Safety First**: Resource lifecycle management and validation
4. **Performance**: Efficient memory allocation and minimal overhead
5. **Maintainability**: Auto-generated bindings from official specifications

## 🔧 Advanced Usage

### Custom Error Handling

```go
// Set global error handling strategy
vk.SetGlobalErrorHandler(&vk.LoggingErrorHandler{
    Logger: func(err error) {
        log.Printf("Vulkan error: %v", err)
    },
})

// Use panic error handler for development
vk.SetGlobalErrorHandler(&vk.PanicErrorHandler{})

// Custom error handling
func handleVulkanError(err error) {
    if vk.IsVulkanError(err) {
        if result, ok := vk.GetVulkanResult(err); ok {
            if result == vulkan.ERROR_DEVICE_LOST {
                // Handle device loss
                recreateDevice()
            }
        }
    }
}
```

### Validation Layers

```go
// Enable validation with custom debug callback
config := vk.DefaultInstanceConfig()
config.EnableValidation = true

instance, err := vk.CreateInstance(config)
if err != nil {
    return err
}

// Custom debug callback
debugCallback := func(severity vk.DebugSeverity, messageType vk.DebugMessageType, message string) {
    switch severity {
    case vk.DebugSeverityError:
        log.Printf("VALIDATION ERROR: %s", message)
    case vk.DebugSeverityWarning:
        log.Printf("VALIDATION WARNING: %s", message)
    case vk.DebugSeverityInfo:
        log.Printf("VALIDATION INFO: %s", message)
    }
}

// Set the callback (implementation details in actual code)
```

### Extension Management

```go
// Check extension support
extensions, err := vk.EnumerateInstanceExtensions("")
if err != nil {
    return err
}

if vk.IsExtensionSupported("VK_EXT_debug_utils", extensions) {
    config.EnabledExtensions = append(config.EnabledExtensions, "VK_EXT_debug_utils")
}

// Device extensions
deviceExtensions, err := physicalDevice.EnumerateDeviceExtensions()
if err != nil {
    return err
}

deviceConfig := vk.DefaultDeviceConfig(physicalDevice)
if vk.IsExtensionSupported("VK_KHR_ray_tracing_pipeline", deviceExtensions) {
    deviceConfig.RequiredExtensions = append(deviceConfig.RequiredExtensions, 
                                           "VK_KHR_ray_tracing_pipeline")
}
```

## 🔄 Keeping Up-to-Date

This wrapper includes an automatic update system that tracks the latest Vulkan specifications:

### Automatic Updates

```bash
# Update to the latest Vulkan specifications
./scripts/update.sh

# Force update even if no changes detected
./scripts/update.sh --force

# Skip backup creation
./scripts/update.sh --skip-backup
```

### Manual Update

```go
// Check if bindings are up-to-date
info := vulkan.GetVersionInfo()
fmt.Println(info)

// Check current version
fmt.Printf("Header Version: %d\n", vulkan.HeaderVersion)
fmt.Printf("Generated from: %s\n", vulkan.GeneratedFromDate)
```

## 🧪 Testing

```bash
# Run all tests
go test ./...

# Run with race detection
go test -race ./...

# Run specific test suite
go test ./pkg/vk/...

# Run benchmarks
go test -bench=. ./...
```

## 🎯 Comparison with Existing Solutions

| Feature | vulkan-go | goki/vulkan | bbredesen/go-vk |
|---------|-----------|-------------|-----------------|
| **Vulkan Version** | ✅ 1.4.326 (Latest) | ⚠️ 1.3.239 (2023) | ⚠️ 1.3.x (Beta) |
| **Auto-Updates** | ✅ Yes | ❌ Manual | ❌ Manual |
| **High-Level API** | ✅ Full wrapper | ❌ Basic bindings | ✅ Go-style API |
| **Memory Management** | ✅ Advanced allocator | ❌ Manual | ❌ Manual |
| **Error Handling** | ✅ Rich errors + validation | ⚠️ Basic | ✅ Go errors |
| **Resource Management** | ✅ Automatic cleanup | ❌ Manual | ⚠️ Partial |
| **Documentation** | ✅ Comprehensive | ⚠️ Basic | ⚠️ API docs only |
| **Examples** | ✅ Multiple complete examples | ⚠️ Basic | ⚠️ Limited |
| **Testing** | ✅ Full suite | ❌ Limited | ❌ Basic |
| **Maintenance** | ✅ Active (2025) | ⚠️ Sporadic | ⚠️ Beta status |

## ❓ FAQ

### **Q: Why another Vulkan wrapper?**
A: Existing Go Vulkan bindings are outdated, lack high-level abstractions, and don't provide the safety and convenience that Go developers expect. This wrapper combines up-to-date bindings with production-ready abstractions.

### **Q: Is this production ready?**
A: Yes! This wrapper includes comprehensive error handling, memory management, resource lifecycle management, and extensive testing. It's designed for production use from day one.

### **Q: How does the auto-update system work?**
A: The update script automatically pulls the latest Vulkan XML registry from the official Khronos repository and regenerates the Go bindings. This ensures you're always working with the latest Vulkan features.

### **Q: Can I use just the low-level bindings?**
A: Absolutely! The `pkg/vulkan` package provides direct access to Vulkan API functions if you prefer to work at a lower level.

### **Q: What about performance?**
A: The wrapper is designed with minimal overhead. The high-level API adds convenience without sacrificing performance, and the memory allocator is optimized for Vulkan's requirements.

### **Q: Platform support?**
A: Windows, Linux, macOS (via MoltenVK), Android, and iOS are all supported. Platform-specific code is handled automatically.

### **Q: How do I contribute?**
A: Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

### Development

```bash
# Clone the repository
git clone https://github.com/christerso/vulkan-go.git
cd vulkan-go

# Install dependencies
go mod tidy

# Run tests
go test ./...

# Update Vulkan bindings
./scripts/update.sh

# Generate code
go run scripts/generate.go
```

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- **Khronos Group** for the Vulkan API and specifications
- **LunarG** for the Vulkan SDK and validation layers
- **Go Community** for the excellent tooling and ecosystem
- **MoltenVK** team for bringing Vulkan to Apple platforms

## 📞 Support

- **Issues**: [GitHub Issues](https://github.com/christerso/vulkan-go/issues)
- **Discussions**: [GitHub Discussions](https://github.com/christerso/vulkan-go/discussions)
- **Documentation**: [pkg.go.dev](https://pkg.go.dev/github.com/christerso/vulkan-go)

---

<div align="center">

**⭐ Star this repository if you find it useful!**

[📚 Documentation](https://pkg.go.dev/github.com/christerso/vulkan-go) • [🚀 Examples](cmd/) • [🐛 Issues](https://github.com/christerso/vulkan-go/issues) • [💬 Discussions](https://github.com/christerso/vulkan-go/discussions)

</div>