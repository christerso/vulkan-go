# Vulkan-Go Project Structure

## Architecture Overview

Our Vulkan wrapper will follow this modular structure:

```
vulkan-go/
├── cmd/                    # Example applications and tools
│   ├── triangle/          # Basic triangle rendering example
│   ├── compute/          # Compute shader example
│   └── generator/        # Code generation tool
├── pkg/
│   ├── vulkan/           # Core Vulkan bindings (auto-generated)
│   │   ├── api.go        # Core API functions
│   │   ├── enums.go      # Vulkan enumerations
│   │   ├── structs.go    # Vulkan structures
│   │   ├── extensions.go # Extension support
│   │   └── platform/     # Platform-specific code
│   ├── vk/              # High-level wrapper API
│   │   ├── instance.go   # Instance management
│   │   ├── device.go     # Device management
│   │   ├── buffer.go     # Buffer utilities
│   │   ├── image.go      # Image utilities
│   │   ├── pipeline.go   # Pipeline creation
│   │   ├── command.go    # Command buffer utilities
│   │   ├── memory.go     # Memory management
│   │   └── sync.go       # Synchronization primitives
│   └── internal/        # Internal utilities
│       ├── loader/      # Dynamic library loading
│       ├── validation/  # Validation layer helpers
│       └── platform/    # Platform detection
├── scripts/             # Build and generation scripts
│   ├── generate.go      # Vulkan API code generator
│   ├── update.sh        # Auto-update script
│   └── build.sh         # Build script
├── assets/             # Example assets (shaders, models)
├── docs/              # Documentation
└── tests/             # Test suite
```

## Design Principles

1. **Separation of Concerns**: Low-level bindings separate from high-level API
2. **Go Idioms**: Proper error handling, interfaces, and type safety  
3. **Auto-generation**: Core bindings generated from Vulkan specification
4. **Memory Safety**: Proper resource lifecycle management
5. **Cross-platform**: Support Windows, Linux, macOS, Android, iOS