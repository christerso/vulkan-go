// Command flythrough renders a large procedural terrain and flies a camera over
// it, measuring frame times and GC pauses to show that Go drives heavy Vulkan
// workloads without hitching. Run:
//
//	go run ./examples/flythrough            # windowed, runs until closed/Escape
//	go run ./examples/flythrough -frames 1800 -n 640
//
// It prints a frame-time and GC report on exit.
package main

import (
	_ "embed"
	"flag"
	"fmt"
	"math"
	"os"
	"runtime"
	"time"
	"unsafe"

	"github.com/christerso/vulkan-go/examples/internal/win"
	"github.com/christerso/vulkan-go/vk"
)

//go:embed shaders/terrain.vert.spv
var vertSPV []byte

//go:embed shaders/terrain.frag.spv
var fragSPV []byte

const framesInFlight = 2

type uniform struct {
	viewProj [16]float32
	lightDir [4]float32
	params   [4]float32
}

func main() {
	gridN := flag.Int("n", 512, "terrain grid resolution (NxN vertices)")
	winW := flag.Int("w", 1280, "window width")
	winH := flag.Int("h", 800, "window height")
	maxFrames := flag.Int("frames", 0, "stop after N frames (0 = until window closed)")
	gcLoad := flag.Bool("gcload", true, "run a background allocator to force frequent GC")
	novalidate := flag.Bool("novalidate", false, "disable validation layer")
	flag.Parse()

	if err := run(*gridN, *winW, *winH, *maxFrames, *gcLoad, !*novalidate); err != nil {
		fmt.Fprintln(os.Stderr, "flythrough:", err)
		os.Exit(1)
	}
}

func run(gridN, winW, winH, maxFrames int, gcLoad, validate bool) error {
	runtime.LockOSThread() // present/acquire on a single thread

	window, err := win.New("vulkan-go flythrough", int32(winW), int32(winH))
	if err != nil {
		return err
	}
	defer window.Destroy()

	if err := vk.Load(); err != nil {
		return err
	}

	exts := append(window.InstanceExtensions(), vk.ExtDebugUtils)
	var layers []string
	if validate {
		layers = []string{vk.ValidationLayer}
	}
	instance, err := vk.CreateInstance(vk.InstanceConfig{
		ApplicationName: "flythrough",
		EngineName:      "vulkan-go",
		Extensions:      exts,
		Layers:          layers,
	})
	if err != nil {
		return err
	}
	defer instance.Destroy()

	var messenger vk.DebugMessenger
	if validate {
		if messenger, err = instance.CreateDebugMessenger(); err != nil {
			return err
		}
		defer messenger.Destroy()
	}

	surface, err := window.CreateSurface(uintptr(instance))
	if err != nil {
		return err
	}
	surf := vk.SurfaceKHR(surface)
	defer instance.DestroySurface(surf)

	devices, err := instance.EnumeratePhysicalDevices()
	if err != nil {
		return err
	}
	if len(devices) == 0 {
		return fmt.Errorf("no Vulkan devices")
	}
	pd := devices[0]
	info := pd.Info()
	gfx, err := pd.GraphicsFamily()
	if err != nil {
		return err
	}
	if !pd.SurfaceSupport(gfx, surf) {
		return fmt.Errorf("graphics queue cannot present")
	}
	device, queue, err := pd.CreateDevice(vk.DeviceConfig{
		GraphicsFamily: gfx,
		Extensions:     []string{"VK_KHR_swapchain"},
	})
	if err != nil {
		return err
	}
	defer device.Destroy()
	defer device.WaitIdle()

	// Surface format and present mode.
	format, colorSpace := chooseFormat(pd, surf)
	present := choosePresentMode(pd, surf)
	const depthFormat = vk.FormatD32Sfloat

	renderPass, err := device.CreateColorDepthRenderPass(format, depthFormat)
	if err != nil {
		return err
	}
	defer device.DestroyRenderPass(renderPass)

	// Command pool and per-frame command buffers.
	pool, err := device.CreateCommandPool(gfx)
	if err != nil {
		return err
	}
	defer device.DestroyCommandPool(pool)
	cmds, err := device.AllocateCommandBuffers(pool, framesInFlight)
	if err != nil {
		return err
	}

	// Mesh.
	terrain := GenerateTerrain(gridN, 400, 70)
	fmt.Printf("GPU: %s (%s)\n", info.Name, info.Type)
	fmt.Printf("terrain: %d vertices, %d triangles\n", len(terrain.Vertices), len(terrain.Indices)/3)

	vertBytes := unsafe.Slice((*byte)(unsafe.Pointer(&terrain.Vertices[0])), len(terrain.Vertices)*int(unsafe.Sizeof(Vertex{})))
	idxBytes := unsafe.Slice((*byte)(unsafe.Pointer(&terrain.Indices[0])), len(terrain.Indices)*4)
	vbuf, err := device.CreateDeviceLocalBuffer(pd, queue, pool, vertBytes, vk.BufferUsageVertexBuffer)
	if err != nil {
		return err
	}
	defer device.DestroyBuffer(vbuf)
	ibuf, err := device.CreateDeviceLocalBuffer(pd, queue, pool, idxBytes, vk.BufferUsageIndexBuffer)
	if err != nil {
		return err
	}
	defer device.DestroyBuffer(ibuf)
	indexCount := uint32(len(terrain.Indices))

	// Shaders and pipeline.
	vs, err := device.CreateShaderModule(vertSPV)
	if err != nil {
		return err
	}
	defer device.DestroyShaderModule(vs)
	fs, err := device.CreateShaderModule(fragSPV)
	if err != nil {
		return err
	}
	defer device.DestroyShaderModule(fs)

	setLayout, err := device.CreateDescriptorSetLayout([]vk.DescriptorBinding{
		{Binding: 0, Type: vk.DescriptorUniformBuffer, Count: 1, Stages: vk.ShaderStageVertex | vk.ShaderStageFragment},
	})
	if err != nil {
		return err
	}
	defer device.DestroyDescriptorSetLayout(setLayout)
	pipelineLayout, err := device.CreatePipelineLayout([]vk.DescriptorSetLayout{setLayout}, 0, 0)
	if err != nil {
		return err
	}
	defer device.DestroyPipelineLayout(pipelineLayout)

	pipeline, err := device.CreateGraphicsPipeline(vk.GraphicsPipelineConfig{
		Layout:       pipelineLayout,
		RenderPass:   renderPass,
		VertexShader: vs,
		FragShader:   fs,
		Bindings:     []vk.VertexInputBinding{{Binding: 0, Stride: uint32(unsafe.Sizeof(Vertex{})), InputRate: vk.VertexInputRateVertex}},
		Attributes: []vk.VertexInputAttribute{
			{Location: 0, Binding: 0, Format: vk.FormatR32G32B32Sfloat, Offset: 0},
			{Location: 1, Binding: 0, Format: vk.FormatR32G32B32Sfloat, Offset: 12},
		},
		Topology:    vk.TopologyTriangleList,
		PolygonMode: vk.PolygonFill,
		CullMode:    vk.CullBack,
		FrontFace:   vk.FrontFaceCounterClockwise,
		DepthTest:   true,
		DepthWrite:  true,
	})
	if err != nil {
		return err
	}
	defer device.DestroyPipeline(pipeline)

	// Per-frame uniform buffers and descriptor sets.
	descPool, err := device.CreateDescriptorPool(framesInFlight, map[vk.DescriptorType]uint32{
		vk.DescriptorUniformBuffer: framesInFlight,
	})
	if err != nil {
		return err
	}
	defer device.DestroyDescriptorPool(descPool)

	ubufs := make([]vk.AllocBuffer, framesInFlight)
	dsets := make([]vk.DescriptorSet, framesInFlight)
	uboSize := vk.DeviceSize(unsafe.Sizeof(uniform{}))
	for i := 0; i < framesInFlight; i++ {
		ubufs[i], err = device.CreateBuffer(pd, vk.BufferConfig{
			Size:       uboSize,
			Usage:      vk.BufferUsageUniformBuffer,
			Properties: vk.MemoryHostVisible | vk.MemoryHostCoherent,
			Map:        true,
		})
		if err != nil {
			return err
		}
		defer device.DestroyBuffer(ubufs[i])
		dsets[i], err = device.AllocateDescriptorSet(descPool, setLayout)
		if err != nil {
			return err
		}
		device.UpdateBufferDescriptor(dsets[i], 0, vk.DescriptorUniformBuffer, ubufs[i].Buffer, 0, uboSize)
	}

	// Sync objects.
	imageAvailable := make([]vk.Semaphore, framesInFlight)
	inFlight := make([]vk.Fence, framesInFlight)
	for i := 0; i < framesInFlight; i++ {
		if imageAvailable[i], err = device.CreateSemaphore(); err != nil {
			return err
		}
		defer device.DestroySemaphore(imageAvailable[i])
		if inFlight[i], err = device.CreateFence(true); err != nil {
			return err
		}
		defer device.DestroyFence(inFlight[i])
	}

	sc, err := newSwapchain(device, pd, surf, renderPass, format, colorSpace, present, depthFormat, window)
	if err != nil {
		return err
	}
	defer sc.destroy(device)

	if gcLoad {
		startGCLoad()
	}

	// Stats.
	frameTimes := make([]float64, 0, 1<<16)
	startMetrics := readGCStats()
	start := time.Now()
	last := start
	frame := 0
	count := 0

	for window.Poll() {
		fi := frame % framesInFlight
		_ = device.WaitFence(inFlight[fi], math.MaxUint64)

		imgIndex, res := device.AcquireNextImage(sc.handle, imageAvailable[fi], math.MaxUint64)
		if res == vk.ErrorOutOfDateKHR {
			device.WaitIdle()
			sc.destroy(device)
			if sc, err = newSwapchain(device, pd, surf, renderPass, format, colorSpace, present, depthFormat, window); err != nil {
				return err
			}
			continue
		} else if !res.Ok() && res != vk.SuboptimalKHR {
			return fmt.Errorf("acquire: %s", res)
		}
		_ = device.ResetFence(inFlight[fi])

		// Update uniform: fly the camera.
		t := float32(time.Since(start).Seconds())
		u := buildUniform(t, terrain.WorldSize, terrain.HeightScale, float32(sc.extent.Width)/float32(sc.extent.Height))
		vk.CopyToMapped(ubufs[fi].Mapped, unsafe.Slice((*byte)(unsafe.Pointer(&u)), int(uboSize)))

		cmd := cmds[fi]
		_ = cmd.Reset()
		_ = cmd.Begin(0)
		area := vk.Rect2D{Extent: sc.extent}
		cmd.BeginRenderPass(renderPass, sc.framebuffers[imgIndex], area, []vk.ClearValue{
			vk.ClearColor(0.52, 0.70, 0.92, 1.0),
			vk.ClearDepthStencil(1.0, 0),
		})
		cmd.BindPipeline(pipeline)
		cmd.SetViewport(vk.Viewport{Width: float32(sc.extent.Width), Height: float32(sc.extent.Height), MaxDepth: 1})
		cmd.SetScissor(area)
		cmd.BindDescriptorSet(pipelineLayout, 0, dsets[fi])
		cmd.BindVertexBuffer(vbuf.Buffer, 0)
		cmd.BindIndexBuffer(ibuf.Buffer, 0, vk.IndexTypeUint32)
		cmd.DrawIndexed(indexCount, 1, 0, 0, 0)
		cmd.EndRenderPass()
		_ = cmd.End()

		if err := queue.Submit(vk.SubmitConfig{
			Wait:      imageAvailable[fi],
			WaitStage: vk.StageColorAttachmentOutput,
			Command:   cmd,
			Signal:    sc.renderFinished[imgIndex],
			Fence:     inFlight[fi],
		}); err != nil {
			return err
		}
		pres := queue.Present(sc.handle, imgIndex, sc.renderFinished[imgIndex])
		if pres == vk.ErrorOutOfDateKHR || pres == vk.SuboptimalKHR {
			device.WaitIdle()
			sc.destroy(device)
			if sc, err = newSwapchain(device, pd, surf, renderPass, format, colorSpace, present, depthFormat, window); err != nil {
				return err
			}
		}

		now := time.Now()
		frameTimes = append(frameTimes, now.Sub(last).Seconds()*1000)
		last = now
		frame++
		count++
		if maxFrames > 0 && count >= maxFrames {
			break
		}
	}

	device.WaitIdle()
	report(frameTimes, time.Since(start), readGCStats().sub(startMetrics))
	return nil
}

// buildUniform animates a camera orbiting and looking over the terrain.
func buildUniform(t, worldSize, heightScale, aspect float32) uniform {
	radius := worldSize * 0.42
	angle := t * 0.12
	eye := vec3{
		x: float32(math.Cos(float64(angle))) * radius,
		y: heightScale*0.9 + 40,
		z: float32(math.Sin(float64(angle))) * radius,
	}
	ahead := angle + 0.6
	center := vec3{
		x: float32(math.Cos(float64(ahead))) * radius * 0.25,
		y: 0,
		z: float32(math.Sin(float64(ahead))) * radius * 0.25,
	}
	view := lookAt(eye, center, vec3{0, 1, 0})
	proj := perspective(float32(math.Pi)/3, aspect, 0.5, 4000)
	vp := mul(proj, view)
	light := normalize(vec3{0.4, 0.85, 0.3})
	return uniform{
		viewProj: vp,
		lightDir: [4]float32{light.x, light.y, light.z, 0},
		params:   [4]float32{heightScale, 0, 0, 0},
	}
}

func chooseFormat(pd vk.PhysicalDevice, surf vk.SurfaceKHR) (vk.Format, uint32) {
	formats, _ := pd.SurfaceFormats(surf)
	for _, f := range formats {
		if f.Format == vk.FormatB8G8R8A8Unorm && f.ColorSpace == vk.ColorSpaceSRGBNonlinear {
			return f.Format, f.ColorSpace
		}
	}
	if len(formats) > 0 {
		return formats[0].Format, formats[0].ColorSpace
	}
	return vk.FormatB8G8R8A8Unorm, vk.ColorSpaceSRGBNonlinear
}

func choosePresentMode(pd vk.PhysicalDevice, surf vk.SurfaceKHR) vk.PresentMode {
	modes, _ := pd.SurfacePresentModes(surf)
	for _, m := range modes {
		if m == vk.PresentModeMailbox {
			return m
		}
	}
	return vk.PresentModeFIFO
}
