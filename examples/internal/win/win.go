// Package win is a minimal SDL3 window backend loaded through purego, used by
// the examples to obtain a Vulkan surface and poll for quit. It is not part of
// the binding's public API.
package win

import (
	"fmt"
	"runtime"
	"unsafe"

	"github.com/ebitengine/purego"
)

const (
	initVideo       uint32 = 0x00000020
	windowVulkan    uint64 = 0x0000000010000000
	windowResizable uint64 = 0x0000000000000020

	eventQuit    uint32 = 0x100
	eventKeyDown uint32 = 0x300
	keyEscape    uint32 = 0x1B
)

var (
	sdlInit                      func(flags uint32) bool
	sdlQuit                      func()
	sdlGetError                  func() uintptr
	sdlCreateWindow              func(title *byte, w, h int32, flags uint64) uintptr
	sdlDestroyWindow             func(window uintptr)
	sdlVulkanGetInstanceExtensions func(count *uint32) uintptr
	sdlVulkanCreateSurface       func(window, instance, allocator uintptr, surface *uint64) bool
	sdlPollEvent                 func(event unsafe.Pointer) bool
	sdlGetWindowSizeInPixels     func(window uintptr, w, h *int32) bool
)

var loaded bool

func load() error {
	if loaded {
		return nil
	}
	h, err := purego.Dlopen("libSDL3.so.0", purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		return fmt.Errorf("win: load SDL3: %w", err)
	}
	purego.RegisterLibFunc(&sdlInit, h, "SDL_Init")
	purego.RegisterLibFunc(&sdlQuit, h, "SDL_Quit")
	purego.RegisterLibFunc(&sdlGetError, h, "SDL_GetError")
	purego.RegisterLibFunc(&sdlCreateWindow, h, "SDL_CreateWindow")
	purego.RegisterLibFunc(&sdlDestroyWindow, h, "SDL_DestroyWindow")
	purego.RegisterLibFunc(&sdlVulkanGetInstanceExtensions, h, "SDL_Vulkan_GetInstanceExtensions")
	purego.RegisterLibFunc(&sdlVulkanCreateSurface, h, "SDL_Vulkan_CreateSurface")
	purego.RegisterLibFunc(&sdlPollEvent, h, "SDL_PollEvent")
	purego.RegisterLibFunc(&sdlGetWindowSizeInPixels, h, "SDL_GetWindowSizeInPixels")
	loaded = true
	return nil
}

func sdlError() string { return cstr(sdlGetError()) }

// Window is an SDL3 window.
type Window struct {
	handle uintptr
}

// New initializes SDL video and creates a resizable Vulkan window.
func New(title string, width, height int32) (*Window, error) {
	if err := load(); err != nil {
		return nil, err
	}
	if !sdlInit(initVideo) {
		return nil, fmt.Errorf("win: SDL_Init: %s", sdlError())
	}
	h := sdlCreateWindow(cbytes(title), width, height, windowVulkan|windowResizable)
	if h == 0 {
		sdlQuit()
		return nil, fmt.Errorf("win: SDL_CreateWindow: %s", sdlError())
	}
	return &Window{handle: h}, nil
}

// InstanceExtensions returns the Vulkan instance extensions SDL needs.
func (w *Window) InstanceExtensions() []string {
	var count uint32
	arr := sdlVulkanGetInstanceExtensions(&count)
	if arr == 0 {
		return nil
	}
	out := make([]string, count)
	for i := uint32(0); i < count; i++ {
		p := *(*uintptr)(unsafe.Pointer(arr + uintptr(i)*unsafe.Sizeof(uintptr(0))))
		out[i] = cstr(p)
	}
	return out
}

// CreateSurface creates a VkSurfaceKHR for the given VkInstance handle. instance
// is the uintptr value of a vk.Instance; the returned uint64 is a vk.SurfaceKHR.
func (w *Window) CreateSurface(instance uintptr) (uint64, error) {
	var surface uint64
	if !sdlVulkanCreateSurface(w.handle, instance, 0, &surface) {
		return 0, fmt.Errorf("win: SDL_Vulkan_CreateSurface: %s", sdlError())
	}
	return surface, nil
}

// PixelSize returns the drawable size in pixels.
func (w *Window) PixelSize() (uint32, uint32) {
	var wi, he int32
	sdlGetWindowSizeInPixels(w.handle, &wi, &he)
	return uint32(wi), uint32(he)
}

// Poll processes pending events and returns false when the window should close
// (quit event or Escape key).
func (w *Window) Poll() bool {
	var ev [128]byte
	for sdlPollEvent(unsafe.Pointer(&ev[0])) {
		typ := *(*uint32)(unsafe.Pointer(&ev[0]))
		switch typ {
		case eventQuit:
			return false
		case eventKeyDown:
			key := *(*uint32)(unsafe.Pointer(&ev[28]))
			if key == keyEscape {
				return false
			}
		}
	}
	return true
}

// Destroy closes the window and shuts down SDL.
func (w *Window) Destroy() {
	if w.handle != 0 {
		sdlDestroyWindow(w.handle)
		w.handle = 0
	}
	sdlQuit()
}

// cbytes returns a NUL-terminated copy of s as *byte, kept alive by the caller.
func cbytes(s string) *byte {
	b := make([]byte, len(s)+1)
	copy(b, s)
	runtime.KeepAlive(b)
	return &b[0]
}

// cstr reads a NUL-terminated C string at p.
func cstr(p uintptr) string {
	if p == 0 {
		return ""
	}
	var n int
	for *(*byte)(unsafe.Pointer(p + uintptr(n))) != 0 {
		n++
	}
	return string(unsafe.Slice((*byte)(unsafe.Pointer(p)), n))
}
