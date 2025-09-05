# Changelog

All notable changes to the Vulkan-Go project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial project structure and Go module setup
- Comprehensive Vulkan 1.4.326 API bindings with auto-generation
- High-level wrapper API with Go idioms and safety features
- Advanced memory management with pooling and allocation tracking
- Resource lifecycle management and automatic cleanup
- Extensive error handling with context and validation
- Cross-platform support (Windows, Linux, macOS, Android, iOS)
- Fancy triangle rendering example with animated effects
- Comprehensive README with usage examples and documentation
- Automatic update script for latest Vulkan specifications
- SPIR-V shader compilation tools and utilities

### Features
- **Always Up-to-Date**: Automatically generated from latest Vulkan XML registry
- **Dual-Layer Architecture**: Both low-level bindings and high-level abstractions
- **Production Ready**: Comprehensive error handling and resource management
- **Developer Friendly**: Rich examples, documentation, and type safety
- **Advanced Memory Management**: Smart allocation with pooling and tracking
- **Extension Support**: Full extension system with automatic detection
- **Debug Integration**: Validation layer support and debug utilities

### Examples
- Triangle rendering with psychedelic animated effects
- Vertex and fragment shaders with noise and color cycling
- Memory management demonstrations
- Error handling patterns
- Cross-platform window integration patterns

### Technical Highlights
- Vulkan 1.4.326 support (latest as of August 2025)
- Generated from official Khronos Vulkan-Headers repository
- Type-safe API with Go generics where applicable
- Thread-safe memory allocator with pooling
- Automatic C string management
- Comprehensive validation and error context
- Resource lifecycle tracking and cleanup
- Cross-platform dynamic library loading

## [1.0.0] - Initial Release

This represents the first complete implementation of the Vulkan-Go wrapper, providing:

### Core Features
- Complete Vulkan 1.4 API coverage
- High-level Go wrapper with safety guarantees
- Memory management and resource lifecycle
- Cross-platform support
- Rich examples and documentation
- Auto-updating build system

### Supported Platforms
- Windows (vulkan-1.dll)
- Linux (libvulkan.so.1)
- macOS (MoltenVK)
- Android (limited)
- iOS (limited)

### Requirements
- Go 1.19 or later
- Vulkan SDK installed
- Platform-specific Vulkan drivers

---

**Note**: This project emphasizes being the most up-to-date and comprehensive Vulkan wrapper for Go, with automatic updates from official specifications and production-ready features.