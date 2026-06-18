package main

import (
	"fmt"
	"strings"
)

// ---- structs ----

func (b *Builder) emitStructs(sb *strings.Builder) {
	sb.WriteString("\nimport \"unsafe\"\n\n")
	for _, t := range b.neededOf("struct") {
		if t.Alias != "" {
			fmt.Fprintf(sb, "type %s = %s\n\n", typeName(t), t.Alias)
			continue
		}
		b.emitStruct(sb, t)
	}
	// silence unused import if no struct used unsafe.Pointer
	sb.WriteString("\nvar _ = unsafe.Pointer(nil)\n")
}

func (b *Builder) emitStruct(sb *strings.Builder, t *xmlType) {
	fmt.Fprintf(sb, "type %s struct {\n", typeName(t))
	members := b.parseMembers(t.Members)
	b.emitFields(sb, members)
	sb.WriteString("}\n\n")
}

// parseMembers parses members and merges consecutive bitfield members that pack
// into a single underlying integer into one field.
func (b *Builder) parseMembers(ms []xmlMember) []memberInfo {
	var out []memberInfo
	for _, m := range ms {
		if !apiIncludesVulkan(m.API) {
			continue
		}
		mi := b.parseMember(m.Raw, m.Type, m.Name)
		out = append(out, mi)
	}
	return mergeBitfields(out)
}

// mergeBitfields collapses runs of bitfield members of the same base type into a
// single non-bitfield field of that base type (one per machine word). This keeps
// struct size/alignment correct without per-bit fields.
func mergeBitfields(in []memberInfo) []memberInfo {
	var out []memberInfo
	for i := 0; i < len(in); i++ {
		if in[i].bitwidth == 0 {
			out = append(out, in[i])
			continue
		}
		// accumulate consecutive bitfields of the same base type
		base := in[i].cType
		bits := 0
		first := in[i]
		j := i
		for j < len(in) && in[j].bitwidth > 0 && in[j].cType == base {
			bits += in[j].bitwidth
			j++
		}
		wordBits := cTypeBits(base)
		if wordBits == 0 {
			wordBits = 32
		}
		words := (bits + wordBits - 1) / wordBits
		if words < 1 {
			words = 1
		}
		// emit `words` fields of the base type, named after the run start
		f := first
		f.bitwidth = 0
		f.pointer = 0
		f.arrayLen = ""
		if words > 1 {
			f.arrayLen = itoa(words)
		}
		out = append(out, f)
		i = j - 1
	}
	return out
}

func cTypeBits(name string) int {
	switch name {
	case "uint8_t", "int8_t", "char":
		return 8
	case "uint16_t", "int16_t":
		return 16
	case "uint32_t", "int32_t", "int", "float", "VkFlags", "VkBool32":
		return 32
	case "uint64_t", "int64_t", "double", "VkFlags64", "VkDeviceSize", "VkDeviceAddress":
		return 64
	}
	return 32
}

func (b *Builder) emitFields(sb *strings.Builder, members []memberInfo) {
	for _, mi := range members {
		ft := b.goFieldType(mi)
		fmt.Fprintf(sb, "\t%s %s\n", mi.goName, ft)
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

// ---- unions ----

func (b *Builder) emitUnions(sb *strings.Builder) {
	sb.WriteString("\nimport \"unsafe\"\n\n")
	for _, t := range b.neededOf("union") {
		if t.Alias != "" {
			fmt.Fprintf(sb, "type %s = %s\n\n", typeName(t), t.Alias)
			continue
		}
		b.emitUnion(sb, t)
	}
	sb.WriteString("\nvar _ = unsafe.Pointer(nil)\n")
}

func (b *Builder) emitUnion(sb *strings.Builder, t *xmlType) {
	name := typeName(t)
	members := b.parseMembers(t.Members)
	// size = max member size in bytes
	maxSize := 0
	for _, mi := range members {
		sz := b.sizeOf(mi)
		if sz > maxSize {
			maxSize = sz
		}
	}
	if maxSize == 0 {
		maxSize = 8
	}
	fmt.Fprintf(sb, "// %s is a C union stored as a byte array sized to its largest member.\n", name)
	fmt.Fprintf(sb, "type %s [%d]byte\n\n", name, maxSize)
	for _, mi := range members {
		ft := b.goFieldType(mi)
		fmt.Fprintf(sb, "func (u *%s) As%s() *%s { return (*%s)(unsafe.Pointer(u)) }\n",
			name, mi.goName, ft, ft)
	}
	sb.WriteString("\n")
}

// sizeOf estimates the byte size of a member for union sizing.
func (b *Builder) sizeOf(mi memberInfo) int {
	if mi.pointer > 0 {
		return 8
	}
	elem := b.scalarSize(mi.cType)
	if mi.arrayLen != "" {
		n := atoiSafe(b.resolveConst(mi.arrayLen))
		if n == 0 {
			n = 1
		}
		m := 1
		if mi.arrayLen2 != "" {
			m = atoiSafe(b.resolveConst(mi.arrayLen2))
			if m == 0 {
				m = 1
			}
		}
		return elem * n * m
	}
	return elem
}

func (b *Builder) resolveConst(s string) string {
	if v, ok := b.constInts[s]; ok {
		return v
	}
	return s
}

// scalarSize returns the size in bytes of a C type (best-effort, recursing into
// structs).
func (b *Builder) scalarSize(name string) int {
	if bits := cTypeBitsKnown(name); bits > 0 {
		return bits / 8
	}
	if _, ok := baseTypeGo(name); ok {
		// VkDeviceSize etc.
		if g, _ := baseTypeGo(name); g == "uint64" {
			return 8
		}
		return 4
	}
	t, ok := b.types[name]
	if !ok {
		return 8
	}
	switch t.Category {
	case "handle":
		return 8
	case "enum":
		return 4
	case "bitmask":
		if b.flagWidth[name] == "uint64" {
			return 8
		}
		return 4
	case "struct", "union":
		sz := 0
		for _, m := range t.Members {
			if !apiIncludesVulkan(m.API) {
				continue
			}
			mi := b.parseMember(m.Raw, m.Type, m.Name)
			s := b.sizeOf(mi)
			if t.Category == "union" {
				if s > sz {
					sz = s
				}
			} else {
				sz += s
			}
		}
		if sz == 0 {
			sz = 8
		}
		return sz
	}
	return 8
}

func cTypeBitsKnown(name string) int {
	switch name {
	case "uint8_t", "int8_t", "char":
		return 8
	case "uint16_t", "int16_t":
		return 16
	case "uint32_t", "int32_t", "int", "float":
		return 32
	case "uint64_t", "int64_t", "double":
		return 64
	}
	return 0
}
