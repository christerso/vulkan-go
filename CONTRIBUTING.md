# Contributing to Vulkan-Go

First off, thank you for considering contributing to Vulkan-Go! It's people like you that make this project a great tool for the Go and graphics programming community.

## 📋 Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Process](#development-process)
- [Pull Request Process](#pull-request-process)
- [Issue Guidelines](#issue-guidelines)
- [Coding Standards](#coding-standards)
- [Testing Guidelines](#testing-guidelines)
- [Documentation](#documentation)

## 🤝 Code of Conduct

This project and everyone participating in it is governed by a simple principle: **Be respectful, be constructive, and help make this project better for everyone**.

Examples of behavior that contributes to a positive environment:
- Using welcoming and inclusive language
- Being respectful of differing viewpoints and experiences
- Gracefully accepting constructive criticism
- Focusing on what is best for the community
- Showing empathy towards other community members

## 🚀 Getting Started

### Prerequisites

- **Go 1.19+** - [Download Go](https://golang.org/dl/)
- **Vulkan SDK** - [Download from LunarG](https://vulkan.lunarg.com/)
- **Git** - For version control
- **Make** - For build automation (optional)

### Setting Up Your Development Environment

1. **Fork the repository** on GitHub
2. **Clone your fork** locally:
   ```bash
   git clone https://github.com/YOUR_USERNAME/vulkan-go.git
   cd vulkan-go
   ```

3. **Add the upstream remote**:
   ```bash
   git remote add upstream https://github.com/christerso/vulkan-go.git
   ```

4. **Install dependencies**:
   ```bash
   make install
   # or
   go mod download && go mod tidy
   ```

5. **Set up development tools**:
   ```bash
   make dev-setup
   ```

6. **Verify everything works**:
   ```bash
   make test
   ```

### Project Structure

Understanding the project layout will help you contribute more effectively:

```
vulkan-go/
├── pkg/
│   ├── vulkan/          # Low-level Vulkan bindings (auto-generated)
│   └── vk/              # High-level Go wrapper
├── cmd/                 # Example applications
├── scripts/             # Build and generation tools
├── assets/              # Shaders and other assets
├── tests/              # Test suites
└── docs/               # Documentation
```

## 🔄 Development Process

### Branching Strategy

- `main` - Stable releases and production code
- `develop` - Integration branch for features
- `feature/*` - Individual feature branches
- `fix/*` - Bug fix branches
- `docs/*` - Documentation updates

### Workflow

1. **Create a feature branch** from `develop`:
   ```bash
   git checkout develop
   git pull upstream develop
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes** following our coding standards

3. **Test your changes**:
   ```bash
   make dev  # Full development build with tests
   ```

4. **Commit your changes**:
   ```bash
   git add .
   git commit -m "Add feature: brief description"
   ```

5. **Push to your fork**:
   ```bash
   git push origin feature/your-feature-name
   ```

6. **Create a Pull Request** on GitHub

## 🔧 Pull Request Process

### Before Submitting

- [ ] Run `make dev` to ensure all checks pass
- [ ] Add tests for new functionality
- [ ] Update documentation if needed
- [ ] Ensure your PR has a clear description

### PR Description Template

```markdown
## Description
Brief description of what this PR does.

## Type of Change
- [ ] Bug fix (non-breaking change which fixes an issue)
- [ ] New feature (non-breaking change which adds functionality)
- [ ] Breaking change (fix or feature that would cause existing functionality to change)
- [ ] Documentation update

## Testing
- [ ] Tests pass locally with `make test`
- [ ] Added tests for new functionality
- [ ] Tested on multiple platforms (if applicable)

## Checklist
- [ ] My code follows the style guidelines
- [ ] I have performed a self-review of my code
- [ ] I have commented my code, particularly in hard-to-understand areas
- [ ] I have made corresponding changes to the documentation
- [ ] My changes generate no new warnings
```

### Review Process

1. **Automated checks** must pass (CI/CD, tests, linting)
2. **Code review** by at least one maintainer
3. **Testing** on multiple platforms (for significant changes)
4. **Documentation review** (if applicable)
5. **Final approval** and merge

## 🐛 Issue Guidelines

### Before Creating an Issue

- Search existing issues to avoid duplicates
- Check the [FAQ section](README.md#-faq) in the README
- Ensure you're using the latest version

### Issue Templates

#### Bug Report
```markdown
**Describe the bug**
A clear and concise description of what the bug is.

**To Reproduce**
Steps to reproduce the behavior:
1. Go to '...'
2. Click on '....'
3. Scroll down to '....'
4. See error

**Expected behavior**
A clear and concise description of what you expected to happen.

**Environment:**
- OS: [e.g. Windows 10, Ubuntu 20.04]
- Go version: [e.g. 1.19.5]
- Vulkan SDK version: [e.g. 1.3.268]
- Graphics card: [e.g. NVIDIA GTX 3080]

**Additional context**
Add any other context about the problem here.
```

#### Feature Request
```markdown
**Is your feature request related to a problem? Please describe.**
A clear and concise description of what the problem is.

**Describe the solution you'd like**
A clear and concise description of what you want to happen.

**Describe alternatives you've considered**
A clear and concise description of any alternative solutions or features you've considered.

**Additional context**
Add any other context or screenshots about the feature request here.
```

## 📝 Coding Standards

### Go Style Guide

We follow the standard Go conventions with some additions:

1. **Formatting**: Use `gofmt` and `goimports`
2. **Naming**: Follow Go naming conventions
3. **Comments**: Document all exported functions and types
4. **Error handling**: Always handle errors explicitly
5. **Testing**: Write tests for all new functionality

### Vulkan-Specific Guidelines

1. **Resource Management**: Always implement proper cleanup
2. **Error Handling**: Use our error wrapper types
3. **Memory Safety**: Prefer safe abstractions over raw pointers
4. **Documentation**: Include Vulkan specification references

### Code Example

```go
// CreateBuffer creates a Vulkan buffer with the specified parameters.
// It automatically handles memory allocation and binding.
// 
// See: https://registry.khronos.org/vulkan/specs/1.3/html/chap12.html#VkBuffer
func (d *Device) CreateBuffer(createInfo BufferCreateInfo) (*Buffer, error) {
    if err := validateBufferCreateInfo(createInfo); err != nil {
        return nil, fmt.Errorf("invalid buffer create info: %w", err)
    }
    
    buffer := &Buffer{
        device: d,
        size:   createInfo.Size,
        usage:  createInfo.Usage,
    }
    
    // Implementation...
    
    return buffer, nil
}
```

## 🧪 Testing Guidelines

### Test Structure

```go
func TestDevice_CreateBuffer(t *testing.T) {
    tests := []struct {
        name        string
        createInfo  BufferCreateInfo
        expectError bool
    }{
        {
            name: "valid buffer",
            createInfo: BufferCreateInfo{
                Size:  1024,
                Usage: BufferUsageVertexBuffer,
            },
            expectError: false,
        },
        {
            name: "zero size buffer",
            createInfo: BufferCreateInfo{
                Size:  0,
                Usage: BufferUsageVertexBuffer,
            },
            expectError: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation...
        })
    }
}
```

### Test Categories

- **Unit tests**: Test individual functions and methods
- **Integration tests**: Test interaction between components
- **Example tests**: Ensure examples compile and run
- **Platform tests**: Test platform-specific functionality

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run benchmarks
make bench

# Run specific test
go test -run TestDevice_CreateBuffer ./pkg/vk
```

## 📚 Documentation

### Code Documentation

- Document all exported functions, types, and constants
- Include examples for complex functionality
- Reference Vulkan specification where applicable
- Use proper Go doc formatting

### README Updates

When adding new features:
- Update the feature list
- Add usage examples
- Update the comparison table if needed
- Add FAQ entries for common questions

### Changelog

Keep the `CHANGELOG.md` updated with:
- New features
- Bug fixes
- Breaking changes
- Deprecations

## 🏷️ Versioning and Releases

We use [Semantic Versioning](https://semver.org/):
- **MAJOR**: Breaking changes
- **MINOR**: New features (backward compatible)
- **PATCH**: Bug fixes (backward compatible)

## 🆘 Getting Help

- **GitHub Discussions**: For questions and general discussion
- **GitHub Issues**: For bug reports and feature requests
- **Discord**: [Join our community](https://discord.gg/vulkan-go) (if available)

## 🙏 Recognition

Contributors are recognized in:
- The project README
- Release notes
- Special mentions for significant contributions

## 📞 Contact

- **Maintainer**: [@christerso](https://github.com/christerso)
- **Email**: [If you have a public email you'd like to share]

---

Thank you for contributing to Vulkan-Go! Your efforts help make graphics programming in Go more accessible and powerful for everyone. 🚀