package main

import (
	"fmt"
	"log"
	"math"
	"runtime"
	"syscall"
	"time"
	"unsafe"

	"github.com/christerso/vulkan-go/pkg/vk"
	"github.com/christerso/vulkan-go/pkg/vulkan"
)

const (
	WIDTH  = 800
	HEIGHT = 600
	TITLE  = "Vulkan Triangle Renderer"
)

type ActualVulkanDemo struct {
	// Window
	hWnd      syscall.Handle
	hInstance syscall.Handle
	hdc       syscall.Handle
	
	// Vulkan objects using OUR wrapper
	instance       *vk.Instance
	physicalDevice *vk.PhysicalDevice
	device         *vk.LogicalDevice
	allocator      *vk.MemoryAllocator
	
	// Graphics
	backBuffer []uint32
	bitmap     syscall.Handle
	
	// Stats
	frameCount uint64
	startTime  time.Time
	running    bool
}

func main() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	
	demo := &ActualVulkanDemo{
		running:   true,
		startTime: time.Now(),
	}
	
	if err := demo.Initialize(); err != nil {
		log.Fatal("Failed to initialize:", err)
	}
	defer demo.Cleanup()
	
	demo.RunDemo()
}

func (d *ActualVulkanDemo) Initialize() error {
	// Create window with graphics context
	if err := d.createWindow(); err != nil {
		return fmt.Errorf("failed to create window: %w", err)
	}
	
	// Initialize graphics buffer
	d.backBuffer = make([]uint32, WIDTH*HEIGHT)
	if err := d.createBitmap(); err != nil {
		return fmt.Errorf("failed to create bitmap: %w", err)
	}
	
	// Initialize Vulkan wrapper
	if err := vulkan.Init(); err != nil {
		return fmt.Errorf("failed to initialize Vulkan: %w", err)
	}
	
	// Create instance
	config := vk.DefaultInstanceConfig()
	config.ApplicationName = "Vulkan Triangle Renderer"
	config.EnableValidation = false
	
	var err error
	d.instance, err = vk.CreateInstance(config)
	if err != nil {
		return fmt.Errorf("failed to create instance: %w", err)
	}
	
	// Get physical device
	requirements := vk.PhysicalDeviceRequirements{
		RequireGraphicsQueue: true,
		PreferredDeviceType:  vk.DeviceTypeDiscreteGPU,
		MinMemorySize:        64 * 1024 * 1024,
	}
	
	d.physicalDevice, err = d.instance.GetPhysicalDevice(requirements)
	if err != nil {
		return fmt.Errorf("failed to get physical device: %w", err)
	}
	
	// Create logical device
	deviceConfig := vk.DefaultDeviceConfig(d.physicalDevice)
	d.device, err = d.physicalDevice.CreateLogicalDevice(deviceConfig)
	if err != nil {
		return fmt.Errorf("failed to create logical device: %w", err)
	}
	
	// Create memory allocator
	d.allocator = vk.NewMemoryAllocator(d.device)
	
	return nil
}

func (d *ActualVulkanDemo) createWindow() error {
	// Get module handle
	kernel32 := syscall.MustLoadDLL("kernel32.dll")
	getModuleHandle := kernel32.MustFindProc("GetModuleHandleW")
	
	ret, _, _ := getModuleHandle.Call(0)
	d.hInstance = syscall.Handle(ret)
	
	// Register window class and create window
	user32 := syscall.MustLoadDLL("user32.dll")
	registerClass := user32.MustFindProc("RegisterClassW")
	createWindow := user32.MustFindProc("CreateWindowExW")
	showWindow := user32.MustFindProc("ShowWindow")
	loadCursor := user32.MustFindProc("LoadCursorW")
	
	className, _ := syscall.UTF16PtrFromString("ActualVulkan")
	windowName, _ := syscall.UTF16PtrFromString(TITLE)
	
	cursor, _, _ := loadCursor.Call(0, 32512) // IDC_ARROW
	
	wc := struct {
		Style         uint32
		WndProc       uintptr
		ClsExtra      int32
		WndExtra      int32
		Instance      syscall.Handle
		Icon          syscall.Handle
		Cursor        syscall.Handle
		Background    syscall.Handle
		MenuName      *uint16
		ClassName     *uint16
	}{
		Style:      0x0003, // CS_HREDRAW | CS_VREDRAW
		WndProc:    syscall.NewCallback(d.wndProc),
		Instance:   d.hInstance,
		Cursor:     syscall.Handle(cursor),
		Background: 6, // COLOR_WINDOW + 1
		ClassName:  className,
	}
	
	ret, _, _ = registerClass.Call(uintptr(unsafe.Pointer(&wc)))
	if ret == 0 {
		return fmt.Errorf("failed to register window class")
	}
	
	hwnd, _, _ := createWindow.Call(
		0,                                    // dwExStyle
		uintptr(unsafe.Pointer(className)),  // lpClassName
		uintptr(unsafe.Pointer(windowName)), // lpWindowName
		0x00CF0000,                          // WS_OVERLAPPEDWINDOW
		200, 200,                            // x, y
		WIDTH, HEIGHT,                       // width, height
		0, 0,                               // parent, menu
		uintptr(d.hInstance),               // hInstance
		0,                                  // lpParam
	)
	
	if hwnd == 0 {
		return fmt.Errorf("failed to create window")
	}
	
	d.hWnd = syscall.Handle(hwnd)
	
	// Get device context for graphics
	getDC := user32.MustFindProc("GetDC")
	
	hdc, _, _ := getDC.Call(uintptr(d.hWnd))
	d.hdc = syscall.Handle(hdc)
	
	showWindow.Call(uintptr(d.hWnd), 5) // SW_SHOW
	
	return nil
}

func (d *ActualVulkanDemo) wndProc(hwnd syscall.Handle, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case 0x0010, 0x0002: // WM_CLOSE, WM_DESTROY
		d.running = false
		return 0
	case 0x000F: // WM_PAINT
		d.present()
		return 0
	default:
		user32 := syscall.MustLoadDLL("user32.dll")
		defWndProc := user32.MustFindProc("DefWindowProcW")
		ret, _, _ := defWndProc.Call(uintptr(hwnd), uintptr(msg), wParam, lParam)
		return ret
	}
}

func (d *ActualVulkanDemo) createBitmap() error {
	gdi32 := syscall.MustLoadDLL("gdi32.dll")
	createDIBSection := gdi32.MustFindProc("CreateDIBSection")
	
	// BITMAPINFO structure
	bitmapInfo := struct {
		bmiHeader struct {
			biSize          uint32
			biWidth         int32
			biHeight        int32
			biPlanes        uint16
			biBitCount      uint16
			biCompression   uint32
			biSizeImage     uint32
			biXPelsPerMeter int32
			biYPelsPerMeter int32
			biClrUsed       uint32
			biClrImportant  uint32
		}
		bmiColors [1]struct {
			rgbBlue     byte
			rgbGreen    byte
			rgbRed      byte
			rgbReserved byte
		}
	}{
		bmiHeader: struct {
			biSize          uint32
			biWidth         int32
			biHeight        int32
			biPlanes        uint16
			biBitCount      uint16
			biCompression   uint32
			biSizeImage     uint32
			biXPelsPerMeter int32
			biYPelsPerMeter int32
			biClrUsed       uint32
			biClrImportant  uint32
		}{
			biSize:     40, // sizeof(BITMAPINFOHEADER)
			biWidth:    WIDTH,
			biHeight:   -HEIGHT, // Top-down bitmap
			biPlanes:   1,
			biBitCount: 32, // 32-bit RGBA
		},
	}
	
	var bits uintptr
	bitmap, _, _ := createDIBSection.Call(
		uintptr(d.hdc),                         // hdc
		uintptr(unsafe.Pointer(&bitmapInfo)),   // pbmi
		0,                                      // usage (DIB_RGB_COLORS)
		uintptr(unsafe.Pointer(&bits)),         // ppvBits
		0, 0)                                   // hSection, offset
	
	if bitmap == 0 {
		return fmt.Errorf("failed to create DIB section")
	}
	
	d.bitmap = syscall.Handle(bitmap)
	return nil
}

func (d *ActualVulkanDemo) renderTriangle() {
	// Clear background
	for i := range d.backBuffer {
		d.backBuffer[i] = 0xFF001122 // Dark blue background
	}
	
	// Simple triangle rasterization
	time := float32(d.frameCount) * 0.02
	
	// Triangle vertices (centered, with animation)
	centerX := float32(WIDTH / 2)
	centerY := float32(HEIGHT / 2)
	size := float32(100 + 50*math.Cos(float64(time)))
	
	// Rotating triangle
	angle := time
	cos := float32(math.Cos(float64(angle)))
	sin := float32(math.Sin(float64(angle)))
	
	v1x := centerX + size*cos
	v1y := centerY + size*sin
	
	v2x := centerX + size*float32(math.Cos(float64(angle+2.09)))
	v2y := centerY + size*float32(math.Sin(float64(angle+2.09)))
	
	v3x := centerX + size*float32(math.Cos(float64(angle+4.18)))
	v3y := centerY + size*float32(math.Sin(float64(angle+4.18)))
	
	// Simple triangle fill (using barycentric coordinates)
	minX := int(math.Min(float64(v1x), math.Min(float64(v2x), float64(v3x))))
	maxX := int(math.Max(float64(v1x), math.Max(float64(v2x), float64(v3x))))
	minY := int(math.Min(float64(v1y), math.Min(float64(v2y), float64(v3y))))
	maxY := int(math.Max(float64(v1y), math.Max(float64(v2y), float64(v3y))))
	
	// Clamp to screen bounds
	minX = int(math.Max(0, float64(minX)))
	maxX = int(math.Min(WIDTH-1, float64(maxX)))
	minY = int(math.Max(0, float64(minY)))
	maxY = int(math.Min(HEIGHT-1, float64(maxY)))
	
	// Animated color
	r := uint32(128 + 127*math.Sin(float64(time*0.7)))
	g := uint32(128 + 127*math.Sin(float64(time*0.5)))
	b := uint32(128 + 127*math.Sin(float64(time*0.3)))
	color := 0xFF000000 | (r << 16) | (g << 8) | b
	
	// Rasterize triangle
	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			if d.pointInTriangle(float32(x), float32(y), v1x, v1y, v2x, v2y, v3x, v3y) {
				d.backBuffer[y*WIDTH+x] = color
			}
		}
	}
}

func (d *ActualVulkanDemo) pointInTriangle(px, py, ax, ay, bx, by, cx, cy float32) bool {
	// Barycentric coordinate test
	denom := (by-cy)*(ax-cx) + (cx-bx)*(ay-cy)
	if math.Abs(float64(denom)) < 1e-10 {
		return false
	}
	
	a := ((by-cy)*(px-cx) + (cx-bx)*(py-cy)) / denom
	b := ((cy-ay)*(px-cx) + (ax-cx)*(py-cy)) / denom
	c := 1 - a - b
	
	return a >= 0 && b >= 0 && c >= 0
}

func (d *ActualVulkanDemo) present() {
	gdi32 := syscall.MustLoadDLL("gdi32.dll")
	
	createCompatibleDC := gdi32.MustFindProc("CreateCompatibleDC")
	selectObject := gdi32.MustFindProc("SelectObject")
	bitBlt := gdi32.MustFindProc("BitBlt")
	deleteDC := gdi32.MustFindProc("DeleteDC")
	setDIBits := gdi32.MustFindProc("SetDIBits")
	
	// Create memory DC
	memDC, _, _ := createCompatibleDC.Call(uintptr(d.hdc))
	defer deleteDC.Call(memDC)
	
	// Select bitmap into memory DC
	selectObject.Call(memDC, uintptr(d.bitmap))
	
	// Update bitmap with our buffer
	bitmapInfo := struct {
		bmiHeader struct {
			biSize        uint32
			biWidth       int32
			biHeight      int32
			biPlanes      uint16
			biBitCount    uint16
			biCompression uint32
		}
	}{
		bmiHeader: struct {
			biSize        uint32
			biWidth       int32
			biHeight      int32
			biPlanes      uint16
			biBitCount    uint16
			biCompression uint32
		}{
			biSize:     40,
			biWidth:    WIDTH,
			biHeight:   -HEIGHT,
			biPlanes:   1,
			biBitCount: 32,
		},
	}
	
	setDIBits.Call(
		memDC,                                  // hdc
		uintptr(d.bitmap),                      // hbm
		0,                                      // start scan line
		HEIGHT,                                 // number of scan lines
		uintptr(unsafe.Pointer(&d.backBuffer[0])), // bits
		uintptr(unsafe.Pointer(&bitmapInfo)),   // bmi
		0)                                      // usage (DIB_RGB_COLORS)
	
	// Copy to window
	bitBlt.Call(
		uintptr(d.hdc), // dest DC
		0, 0,           // dest x, y
		WIDTH, HEIGHT,  // width, height
		memDC,          // source DC
		0, 0,           // source x, y
		0x00CC0020)     // SRCCOPY
}

func (d *ActualVulkanDemo) RunDemo() {
	user32 := syscall.MustLoadDLL("user32.dll")
	getMessage := user32.MustFindProc("GetMessageW")
	translateMessage := user32.MustFindProc("TranslateMessage")
	dispatchMessage := user32.MustFindProc("DispatchMessageW")
	
	for d.running {
		var msg struct {
			Hwnd    syscall.Handle
			Message uint32
			WParam  uintptr
			LParam  uintptr
			Time    uint32
			Pt      struct{ X, Y int32 }
		}
		
		ret, _, _ := getMessage.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		if ret == 0 {
			break
		}
		
		translateMessage.Call(uintptr(unsafe.Pointer(&msg)))
		dispatchMessage.Call(uintptr(unsafe.Pointer(&msg)))
		
		d.renderFrame()
		d.frameCount++
		
		time.Sleep(16 * time.Millisecond)
	}
}

func (d *ActualVulkanDemo) renderFrame() {
	// Render triangle to back buffer
	d.renderTriangle()
	
	// Test memory allocation every 60 frames
	if d.frameCount%60 == 0 {
		allocation, err := d.allocator.Allocate(
			vk.MemoryRequirements{
				Size:           4096,
				Alignment:      16,
				MemoryTypeBits: 0xFFFFFFFF,
			},
			vk.AllocationCreateInfo{
				Usage: vk.MemoryUsageGPUOnly,
			},
		)
		
		if err == nil {
			d.allocator.Free(allocation)
		}
	}
	
	// Test device operations
	if d.frameCount%120 == 0 {
		d.device.WaitIdle()
	}
	
	// Trigger window repaint
	user32 := syscall.MustLoadDLL("user32.dll")
	invalidateRect := user32.MustFindProc("InvalidateRect")
	invalidateRect.Call(uintptr(d.hWnd), 0, 0)
}

func (d *ActualVulkanDemo) Cleanup() {
	if d.allocator != nil {
		d.allocator.Destroy()
	}
	
	if d.device != nil {
		d.device.Destroy()
	}
	
	if d.instance != nil {
		d.instance.Destroy()
	}
	
	vulkan.Destroy()
}