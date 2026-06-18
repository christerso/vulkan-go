// Command vkinfo loads Vulkan through purego, creates an instance and a logical
// device, and prints the GPU. It verifies the cgo-free binding end to end.
package main

import (
	"fmt"
	"os"

	"github.com/christerso/vulkan-go/vk"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "vkinfo:", err)
		os.Exit(1)
	}
}

func run() error {
	if err := vk.Load(); err != nil {
		return err
	}
	instance, err := vk.CreateInstance(vk.InstanceConfig{
		ApplicationName: "vkinfo",
		EngineName:      "chime",
	})
	if err != nil {
		return err
	}
	defer instance.Destroy()

	devices, err := instance.EnumeratePhysicalDevices()
	if err != nil {
		return err
	}
	if len(devices) == 0 {
		return fmt.Errorf("no Vulkan physical devices")
	}

	for i, pd := range devices {
		info := pd.Info()
		fmt.Printf("GPU %d: %s (%s)\n", i, info.Name, info.Type)
		fmt.Printf("  API %d.%d.%d  vendor 0x%04x  device 0x%04x\n",
			vk.VersionMajor(info.APIVersion), vk.VersionMinor(info.APIVersion),
			vk.VersionPatch(info.APIVersion), info.VendorID, info.DeviceID)
		families := pd.QueueFamilies()
		fmt.Printf("  queue families: %d\n", len(families))
	}

	pd := devices[0]
	gfx, err := pd.GraphicsFamily()
	if err != nil {
		return err
	}
	device, queue, err := pd.CreateDevice(vk.DeviceConfig{GraphicsFamily: gfx})
	if err != nil {
		return err
	}
	defer device.Destroy()
	fmt.Printf("created logical device on graphics family %d, queue %#x\n", gfx, uintptr(queue))
	return nil
}
