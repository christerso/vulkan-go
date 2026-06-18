package main

import (
	"regexp"
	"strconv"
	"strings"
)

// Builder holds resolved registry data and the set of items in scope.
type Builder struct {
	reg *xmlRegistry

	// type definitions indexed by C name
	types map[string]*xmlType
	// enums group indexed by enum-type name
	enumGroups map[string]*xmlEnums
	// commands indexed by C name (resolved through aliases)
	commands map[string]*xmlCommand
	// command alias name -> real name
	cmdAlias map[string]string

	// extension number indexed by extension name
	extNumber map[string]int

	// platform-tainted type names (reference non-Linux typedefs)
	platformTypes map[string]bool

	// in-scope sets
	needType map[string]bool
	needCmd  map[string]bool

	// integer constants from API Constants (for array sizes), value as string
	constInts map[string]string

	// enum values: enumType -> ordered list of (name,value)
	enumValues map[string][]enumConst
	// already-emitted enum constant names (dedupe across extends/aliases)
	seenEnumConst map[string]bool

	// bitmask flag type name -> underlying go type ("uint32"/"uint64")
	flagWidth map[string]string
	// bitmask FlagBits enum name -> the Flags typedef name that should carry consts
	bitsToFlags map[string]string
}

type enumConst struct {
	name  string
	value int64
	// for bitmask 64-bit we keep uint64 separately when needed; value covers most
	isU64 bool
	uval  uint64
}

// platformTypedefs are external C typedefs we cannot represent on Linux without
// extra headers. Any type transitively referencing one is skipped.
var platformTypedefs = map[string]bool{
	"HINSTANCE": true, "HWND": true, "HMONITOR": true, "HANDLE": true,
	"SECURITY_ATTRIBUTES": true, "DWORD": true, "LPCWSTR": true,
	"Display": true, "VisualID": true, "Window": true, "RROutput": true,
	"xcb_connection_t": true, "xcb_visualid_t": true, "xcb_window_t": true,
	"wl_display": true, "wl_surface": true,
	"ANativeWindow": true, "AHardwareBuffer": true, "OHNativeWindow": true,
	"CAMetalLayer": true, "MTLDevice_id": true, "MTLCommandQueue_id": true,
	"MTLBuffer_id": true, "MTLTexture_id": true, "MTLSharedEvent_id": true,
	"IOSurfaceRef": true, "MTLDevice": true, "MTLCommandQueue": true,
	"MTLBuffer": true, "MTLTexture": true, "MTLSharedEvent": true,
	"GgpStreamDescriptor": true, "GgpFrameToken": true,
	"_screen_context": true, "_screen_window": true, "_screen_buffer": true,
	"NvSciSyncAttrList": true, "NvSciSyncObj": true, "NvSciSyncFence": true,
	"NvSciBufAttrList": true, "NvSciBufObj": true,
	"zx_handle_t": true, "StdVideoH264ProfileIdc": true,
}

// isExternalCodecType reports whether a type comes from the vk_video std headers
// (StdVideoH264*, StdVideoAV1*, etc.), which we do not bind.
func isExternalCodecType(name string) bool {
	return strings.HasPrefix(name, "StdVideo")
}

func newBuilder(reg *xmlRegistry) *Builder {
	b := &Builder{
		reg:           reg,
		types:         map[string]*xmlType{},
		enumGroups:    map[string]*xmlEnums{},
		commands:      map[string]*xmlCommand{},
		cmdAlias:      map[string]string{},
		extNumber:     map[string]int{},
		platformTypes: map[string]bool{},
		needType:      map[string]bool{},
		needCmd:       map[string]bool{},
		constInts:     map[string]string{},
		enumValues:    map[string][]enumConst{},
		seenEnumConst: map[string]bool{},
		flagWidth:     map[string]string{},
		bitsToFlags:   map[string]string{},
	}
	b.index()
	return b
}

func typeName(t *xmlType) string {
	if t.Name != "" {
		return t.Name
	}
	if t.Category == "funcpointer" && t.Proto.Name != "" {
		return t.Proto.Name
	}
	return t.NameInner
}

func (b *Builder) index() {
	for i := range b.reg.Types.Type {
		t := &b.reg.Types.Type[i]
		n := typeName(t)
		if n == "" {
			continue
		}
		if !apiIncludesVulkan(t.API) {
			continue // vulkansc-only variant; ignore
		}
		// keep first definition; ignore later duplicate aliases overwriting
		if _, ok := b.types[n]; !ok {
			b.types[n] = t
		}
	}
	for i := range b.reg.Enums {
		e := &b.reg.Enums[i]
		if e.Name == "API Constants" {
			for _, c := range e.Enum {
				if c.Alias != "" {
					continue
				}
				b.constInts[c.Name] = c.Value
			}
			continue
		}
		b.enumGroups[e.Name] = e
	}
	// commands + aliases
	for i := range b.reg.Commands.Command {
		c := &b.reg.Commands.Command[i]
		if c.Alias != "" {
			b.cmdAlias[c.Name] = c.Alias
			continue
		}
		name := c.Proto.Name
		if name != "" {
			b.commands[name] = c
		}
	}
	// resolve flag-bit -> flags mapping and width
	for i := range b.reg.Types.Type {
		t := &b.reg.Types.Type[i]
		if t.Category != "bitmask" {
			continue
		}
		flagsName := typeName(t)
		width := "uint32"
		if strings.Contains(t.TypeInner, "Flags64") {
			width = "uint64"
		}
		b.flagWidth[flagsName] = width
		// 'requires' or 'bitvalues' attr names the FlagBits enum
		bits := t.Requires
		if bits == "" {
			bits = t.BitValues
		}
		if bits != "" {
			b.bitsToFlags[bits] = flagsName
		}
	}
	for _, ext := range b.reg.Extensions.Extension {
		if n, err := strconv.Atoi(ext.Number); err == nil {
			b.extNumber[ext.Name] = n
		}
	}
}

// ---- scope selection ----

// isCoreVersionFeature matches the core version profile feature names.
func isCoreVersionFeature(name string) bool {
	for _, p := range []string{"VK_VERSION_", "VK_BASE_VERSION_", "VK_GRAPHICS_VERSION_", "VK_COMPUTE_VERSION_"} {
		if strings.HasPrefix(name, p) {
			return true
		}
	}
	return false
}

// extensionInScope reports whether we generate the extension.
func extensionInScope(ext *xmlExtension) bool {
	if !apiIncludesVulkan(ext.Supported) {
		return false
	}
	if ext.Platform != "" {
		return false // needs non-Linux platform headers
	}
	if !strings.HasPrefix(ext.Name, "VK_KHR_") && !strings.HasPrefix(ext.Name, "VK_EXT_") {
		return false
	}
	return true
}

// collectScope walks features and in-scope extensions, marking required types
// and commands and computing extension-added enum values.
func (b *Builder) collectScope() {
	for i := range b.reg.Features {
		f := &b.reg.Features[i]
		if !apiIncludesVulkan(f.API) {
			continue
		}
		// The registry splits core API content across version "profile" features:
		// VK_BASE_VERSION_x_y, VK_GRAPHICS_VERSION_x_y, VK_COMPUTE_VERSION_x_y and
		// the umbrella VK_VERSION_x_y. Process all of them.
		if !isCoreVersionFeature(f.Name) {
			continue
		}
		extNum := 0 // core features: extends use their own extnumber attr
		for _, r := range f.Require {
			b.applyRequire(r, extNum)
		}
	}
	for i := range b.reg.Extensions.Extension {
		ext := &b.reg.Extensions.Extension[i]
		if !extensionInScope(ext) {
			continue
		}
		extNum := b.extNumber[ext.Name]
		for _, r := range ext.Require {
			b.applyRequire(r, extNum)
		}
	}
}

func (b *Builder) applyRequire(r xmlRequire, extNum int) {
	for _, t := range r.Type {
		b.markType(t.Name)
	}
	for _, c := range r.Command {
		b.markCommand(c.Name)
	}
	for _, e := range r.Enum {
		b.addEnumValue(e, extNum)
	}
}

func (b *Builder) markCommand(name string) {
	if real, ok := b.cmdAlias[name]; ok {
		name = real
	}
	if b.needCmd[name] {
		return
	}
	cmd, ok := b.commands[name]
	if !ok {
		return
	}
	// if any param/return type is platform-tainted, skip the command
	if b.commandTainted(cmd) {
		return
	}
	b.needCmd[name] = true
	// pull in referenced types
	b.markType(cmd.Proto.Type)
	for _, p := range cmd.Params {
		b.markType(p.Type)
	}
}

func (b *Builder) commandTainted(cmd *xmlCommand) bool {
	if b.typeTainted(cmd.Proto.Type, map[string]bool{}) {
		return true
	}
	for _, p := range cmd.Params {
		if b.typeTainted(p.Type, map[string]bool{}) {
			return true
		}
	}
	return false
}

// markType marks a C type name (and its dependencies) as needed. Platform-tainted
// types are skipped silently.
func (b *Builder) markType(name string) {
	if name == "" || b.needType[name] {
		return
	}
	if _, base := scalarCType(name); base {
		return // primitive C type, no emission needed
	}
	t, ok := b.types[name]
	if !ok {
		return
	}
	if b.typeTainted(name, map[string]bool{}) {
		return
	}
	b.needType[name] = true

	switch t.Category {
	case "struct", "union":
		for _, m := range t.Members {
			b.markType(m.Type)
		}
	case "funcpointer":
		b.markType(t.Proto.Type)
		for _, p := range t.Params {
			b.markType(p.Type)
		}
	case "bitmask":
		// underlying VkFlags + the FlagBits enum
		b.markType(t.TypeInner)
		bits := t.Requires
		if bits == "" {
			bits = t.BitValues
		}
		if bits != "" {
			b.markType(bits)
		}
	case "handle", "enum", "basetype":
		// nothing further; aliases handled via alias attr
	}
	if t.Alias != "" {
		b.markType(t.Alias)
	}
}

// typeTainted reports whether a C type transitively references a platform typedef.
func (b *Builder) typeTainted(name string, seen map[string]bool) bool {
	if name == "" || seen[name] {
		return false
	}
	if platformTypedefs[name] || isExternalCodecType(name) {
		return true
	}
	seen[name] = true
	t, ok := b.types[name]
	if !ok {
		return false
	}
	if t.Requires != "" && platformTypedefs[t.Requires] {
		return true
	}
	switch t.Category {
	case "struct", "union":
		for _, m := range t.Members {
			if b.typeTainted(m.Type, seen) {
				return true
			}
		}
	case "funcpointer":
		if b.typeTainted(t.Proto.Type, seen) {
			return true
		}
		for _, p := range t.Params {
			if b.typeTainted(p.Type, seen) {
				return true
			}
		}
	case "bitmask", "basetype":
		if b.typeTainted(t.TypeInner, seen) {
			return true
		}
	}
	if t.Alias != "" && b.typeTainted(t.Alias, seen) {
		return true
	}
	return false
}

// ---- enum value computation ----

func (b *Builder) addEnumValue(e xmlEnum, extNum int) {
	if e.Extends == "" {
		// constant defined inside a require (e.g. SPEC_VERSION); ignore unless API const
		return
	}
	if !apiIncludesVulkan("") { // extends always apply
	}
	target := e.Extends
	// dedupe by constant name
	if b.seenEnumConst[e.Name] {
		return
	}

	var val int64
	var uval uint64
	isU64 := b.flagWidth[b.flagsForBits(target)] == "uint64"

	switch {
	case e.Alias != "":
		// resolve alias to its already-computed value if present
		av, ok := b.lookupEnumValue(target, e.Alias)
		if !ok {
			return
		}
		val = av
		uval = uint64(av)
	case e.Value != "":
		v, err := parseIntLiteral(e.Value)
		if err != nil {
			return
		}
		val = v
		uval = uint64(v)
	case e.BitPos != "":
		bp, err := strconv.Atoi(e.BitPos)
		if err != nil {
			return
		}
		uval = uint64(1) << uint(bp)
		val = int64(uval)
		isU64 = isU64 || bp >= 32
	case e.Offset != "":
		off, err := strconv.Atoi(e.Offset)
		if err != nil {
			return
		}
		en := extNum
		if e.ExtNumber != "" {
			if n, err := strconv.Atoi(e.ExtNumber); err == nil {
				en = n
			}
		}
		val = 1000000000 + int64(en-1)*1000 + int64(off)
		if e.Dir == "-" {
			val = -val
		}
		uval = uint64(val)
	default:
		return
	}

	b.seenEnumConst[e.Name] = true
	b.enumValues[target] = append(b.enumValues[target], enumConst{
		name: e.Name, value: val, isU64: isU64, uval: uval,
	})
}

func (b *Builder) flagsForBits(bits string) string {
	if f, ok := b.bitsToFlags[bits]; ok {
		return f
	}
	return bits
}

func (b *Builder) lookupEnumValue(enumType, name string) (int64, bool) {
	for _, c := range b.enumValues[enumType] {
		if c.name == name {
			return c.value, true
		}
	}
	// search the enum group's base definition
	if g, ok := b.enumGroups[enumType]; ok {
		for _, c := range g.Enum {
			if c.Name == name {
				if v, err := parseIntLiteral(c.Value); err == nil {
					return v, true
				}
				if c.BitPos != "" {
					if bp, err := strconv.Atoi(c.BitPos); err == nil {
						return int64(1) << uint(bp), true
					}
				}
			}
		}
	}
	return 0, false
}

var hexRe = regexp.MustCompile(`^0[xX][0-9a-fA-F]+`)

func parseIntLiteral(s string) (int64, error) {
	s = strings.TrimSpace(s)
	s = strings.TrimSuffix(s, "ULL")
	s = strings.TrimSuffix(s, "LL")
	s = strings.TrimSuffix(s, "U")
	s = strings.TrimSuffix(s, "L")
	s = strings.TrimSuffix(s, "F")
	s = strings.TrimSuffix(s, "f")
	neg := false
	if strings.HasPrefix(s, "~") {
		// ~0 style -> all ones; handled by caller width, approximate as -1
		s = strings.TrimPrefix(s, "~")
		if s == "0" {
			return -1, nil
		}
	}
	if strings.HasPrefix(s, "(") && strings.HasSuffix(s, ")") {
		s = s[1 : len(s)-1]
	}
	if strings.HasPrefix(s, "-") {
		neg = true
		s = s[1:]
	}
	var v int64
	var err error
	if hexRe.MatchString(s) {
		var u uint64
		u, err = strconv.ParseUint(s[2:], 16, 64)
		v = int64(u)
	} else {
		v, err = strconv.ParseInt(s, 10, 64)
	}
	if err != nil {
		return 0, err
	}
	if neg {
		v = -v
	}
	return v, nil
}
