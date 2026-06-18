package main

import (
	"fmt"
	"sort"
	"strings"
)

type cmdLevel int

const (
	levelGlobal cmdLevel = iota
	levelInstance
	levelDevice
)

// resolvedCommand is a command ready to emit.
type resolvedCommand struct {
	name    string
	retGo   string // "VkResult", "" (void), or other
	params  []memberInfo
	level   cmdLevel
}

func (b *Builder) resolveCommands() []resolvedCommand {
	var names []string
	for n := range b.needCmd {
		names = append(names, n)
	}
	sort.Strings(names)

	var out []resolvedCommand
	for _, n := range names {
		// These two are declared and bound by hand in loader.go.
		if n == "vkGetInstanceProcAddr" || n == "vkGetDeviceProcAddr" {
			continue
		}
		cmd := b.commands[n]
		if cmd == nil {
			continue
		}
		rc := resolvedCommand{name: n}
		rc.retGo = b.commandReturn(cmd.Proto.Type)
		for _, p := range cmd.Params {
			if !apiIncludesVulkan(p.API) {
				continue
			}
			mi := b.parseMember(p.Raw, p.Type, p.Name)
			rc.params = append(rc.params, mi)
		}
		rc.level = b.classify(cmd)
		out = append(out, rc)
	}
	return out
}

func (b *Builder) commandReturn(cType string) string {
	if cType == "VkResult" {
		return "VkResult"
	}
	if cType == "void" || cType == "" {
		return ""
	}
	// other scalar/handle returns
	return b.goValueType(cType)
}

// classify determines the load level by the first parameter's handle type,
// resolving handle aliases.
func (b *Builder) classify(cmd *xmlCommand) cmdLevel {
	name := cmd.Proto.Name
	switch name {
	case "vkGetInstanceProcAddr", "vkCreateInstance",
		"vkEnumerateInstanceVersion", "vkEnumerateInstanceLayerProperties",
		"vkEnumerateInstanceExtensionProperties":
		return levelGlobal
	}
	if len(cmd.Params) == 0 {
		return levelGlobal
	}
	first := b.resolveHandle(cmd.Params[0].Type)
	switch first {
	case "VkInstance", "VkPhysicalDevice":
		return levelInstance
	case "VkDevice", "VkQueue", "VkCommandBuffer":
		return levelDevice
	}
	return levelGlobal
}

// resolveHandle follows handle aliases to the canonical name.
func (b *Builder) resolveHandle(name string) string {
	for {
		t, ok := b.types[name]
		if !ok || t.Category != "handle" || t.Alias == "" {
			return name
		}
		name = t.Alias
	}
}

func (b *Builder) emitCommands(sb *strings.Builder) {
	sb.WriteString("\nimport \"unsafe\"\n\n")
	cmds := b.resolveCommands()
	sb.WriteString("// Command function pointers. Nil until the matching Load* call binds them.\n")
	sb.WriteString("// Each variable is the exported Vk-cased name; it is bound to the real\n")
	sb.WriteString("// lowercase vk* entry point at load time.\nvar (\n")
	for _, c := range cmds {
		fmt.Fprintf(sb, "\t%s func(%s)%s\n", exportCmd(c.name), b.paramSig(c.params), retSuffix(c.retGo))
	}
	sb.WriteString(")\n")
	sb.WriteString("\nvar _ = unsafe.Pointer(nil)\n")
}

// exportCmd capitalizes the leading "vk" so the command variable is exported.
func exportCmd(name string) string {
	if len(name) >= 1 && name[0] == 'v' {
		return "V" + name[1:]
	}
	return name
}

func retSuffix(r string) string {
	if r == "" {
		return ""
	}
	return " " + r
}

// paramSig builds the Go parameter list for a command function variable.
func (b *Builder) paramSig(params []memberInfo) string {
	var parts []string
	for i, mi := range params {
		gt := b.goParamType(mi)
		pn := paramName(mi.goName, i)
		parts = append(parts, pn+" "+gt)
	}
	return strings.Join(parts, ", ")
}

// goParamType maps a parameter. Pointers -> unsafe.Pointer; handles/enums/scalars
// by value.
func (b *Builder) goParamType(mi memberInfo) string {
	if mi.pointer > 0 {
		return "unsafe.Pointer"
	}
	if mi.arrayLen != "" {
		// array params are passed as pointers in C; treat as unsafe.Pointer
		return "unsafe.Pointer"
	}
	return b.goValueType(mi.cType)
}

func paramName(n string, i int) string {
	if n == "" {
		return fmt.Sprintf("a%d", i)
	}
	// lower first letter for params; ensure not keyword
	r := []rune(n)
	if r[0] >= 'A' && r[0] <= 'Z' {
		r[0] += 32
	}
	out := string(r)
	switch out {
	case "type", "range", "func", "map", "string", "len", "cap", "select",
		"chan", "interface", "package", "import", "return", "var", "const":
		out += "_"
	}
	return out
}

// ---- loader ----

func (b *Builder) emitLoader(sb *strings.Builder) {
	cmds := b.resolveCommands()
	var global, instance, device []string
	for _, c := range cmds {
		switch c.level {
		case levelGlobal:
			global = append(global, c.name)
		case levelInstance:
			instance = append(instance, c.name)
		case levelDevice:
			device = append(device, c.name)
		}
	}

	sb.WriteString(`
import (
	"fmt"

	"github.com/ebitengine/purego"
)

var (
	libVulkan             uintptr
	vkGetInstanceProcAddr func(instance uintptr, name string) uintptr
	vkGetDeviceProcAddr   func(device uintptr, name string) uintptr
)

// Load opens the Vulkan loader and binds global commands. It is idempotent.
func Load() error {
	if libVulkan != 0 {
		return nil
	}
	var h uintptr
	var err error
	for _, name := range []string{"libvulkan.so.1", "libvulkan.so"} {
		h, err = purego.Dlopen(name, purego.RTLD_NOW|purego.RTLD_GLOBAL)
		if err == nil && h != 0 {
			break
		}
	}
	if h == 0 {
		return fmt.Errorf("vulkan: load loader: %w", err)
	}
	libVulkan = h
	purego.RegisterLibFunc(&vkGetInstanceProcAddr, h, "vkGetInstanceProcAddr")
	loadGlobalCommands()
	return nil
}

// bindInstance binds a command resolved through vkGetInstanceProcAddr. A zero
// address means the command is unavailable; the variable is left nil.
func bindInstance(fptr any, instance uintptr, name string) {
	addr := vkGetInstanceProcAddr(instance, name)
	if addr == 0 {
		return
	}
	purego.RegisterFunc(fptr, addr)
}

// bindDevice binds a command resolved through vkGetDeviceProcAddr.
func bindDevice(fptr any, device uintptr, name string) {
	addr := vkGetDeviceProcAddr(device, name)
	if addr == 0 {
		return
	}
	purego.RegisterFunc(fptr, addr)
}

`)

	emitBindFn(sb, "loadGlobalCommands", "", global, "bindInstance", "0")
	emitBindFn(sb, "LoadInstance", "instance uintptr", instance, "bindInstance", "instance")

	// LoadDevice must also bind vkGetDeviceProcAddr from the instance level first;
	// it's bound during LoadInstance. Here we use vkGetDeviceProcAddr directly.
	sb.WriteString("// LoadInstance binds all instance-level and physical-device-level commands.\n")
	sb.WriteString("// (generated above)\n\n")

	sb.WriteString("// LoadDevice binds all device-level commands for the given device.\n")
	sb.WriteString("func LoadDevice(device uintptr) {\n")
	for _, n := range device {
		fmt.Fprintf(sb, "\tbindDevice(&%s, device, %q)\n", exportCmd(n), n)
	}
	sb.WriteString("}\n")
}

func emitBindFn(sb *strings.Builder, fn, arg string, names []string, bind, handle string) {
	if fn == "loadGlobalCommands" {
		fmt.Fprintf(sb, "func %s() {\n", fn)
	} else {
		fmt.Fprintf(sb, "// %s binds all instance-level (and physical-device-level) commands.\nfunc %s(%s) {\n", fn, fn, arg)
		// also bind vkGetDeviceProcAddr so LoadDevice works
	}
	for _, n := range names {
		fmt.Fprintf(sb, "\t%s(&%s, %s, %q)\n", bind, exportCmd(n), handle, n)
	}
	if fn != "loadGlobalCommands" {
		sb.WriteString("\tbindInstance(&vkGetDeviceProcAddr, instance, \"vkGetDeviceProcAddr\")\n")
	}
	sb.WriteString("}\n\n")
}
