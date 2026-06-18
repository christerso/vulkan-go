package main

import (
	"regexp"
	"strings"
)

// scalarCType maps a C scalar type name to its Go equivalent. The bool reports
// whether name is a known C primitive (so it needs no Go type emission).
func scalarCType(name string) (string, bool) {
	switch name {
	case "void":
		return "", true // only meaningful as a pointer
	case "char":
		return "byte", true
	case "uint8_t":
		return "uint8", true
	case "uint16_t":
		return "uint16", true
	case "uint32_t":
		return "uint32", true
	case "uint64_t":
		return "uint64", true
	case "int8_t":
		return "int8", true
	case "int16_t":
		return "int16", true
	case "int32_t":
		return "int32", true
	case "int64_t":
		return "int64", true
	case "int":
		return "int32", true
	case "float":
		return "float32", true
	case "double":
		return "float64", true
	case "size_t":
		return "uintptr", true
	}
	return "", false
}

// baseTypeGo maps the fixed <basetype> names to Go types.
func baseTypeGo(name string) (string, bool) {
	switch name {
	case "VkBool32", "VkFlags", "VkSampleMask":
		return "uint32", true
	case "VkFlags64", "VkDeviceSize", "VkDeviceAddress":
		return "uint64", true
	}
	return "", false
}

// goTypeName converts a C type name to the Go type name used in the generated
// package (Vk prefix stripped, otherwise identical). Scalars map directly.
func (b *Builder) goTypeName(cname string) string {
	if g, ok := scalarCType(cname); ok {
		if g == "" {
			return "byte" // void-by-value shouldn't happen; placeholder
		}
		return g
	}
	// Use the registry name verbatim (keeps Vk prefix => unique, valid Go ident).
	return cname
}

// memberInfo is a parsed struct member or command param.
type memberInfo struct {
	cType     string // base C type name
	goName    string // exported Go field/param name
	pointer   int    // number of '*'
	arrayLen  string // resolved array length expression (Go int literal) or ""
	arrayLen2 string // second dimension for [N][M]
	bitwidth  int    // bitfield width, 0 if none
	isStruct  bool   // base type is a struct/union (value embed)
}

var (
	nameTagRe = regexp.MustCompile(`<name>([^<]*)</name>`)
	typeTagRe = regexp.MustCompile(`<type>([^<]*)</type>`)
	enumTagRe = regexp.MustCompile(`<enum>([^<]*)</enum>`)
	bitRe     = regexp.MustCompile(`:(\d+)\s*$`)
)

// parseMember parses one member's raw XML into a memberInfo. The raw form looks
// like: const <type>void</type>*  <name>pNext</name>  or
// <type>char</type> <name>deviceName</name>[<enum>VK_MAX...</enum>]
func (b *Builder) parseMember(raw, cType, cName string) memberInfo {
	mi := memberInfo{cType: cType, goName: exportName(cName)}

	// strip the <type> and <name> elements, keeping surrounding punctuation
	// to count pointers and detect arrays.
	work := raw
	// remove comments
	if i := strings.Index(work, "<comment>"); i >= 0 {
		j := strings.Index(work, "</comment>")
		if j > i {
			work = work[:i] + work[j+len("</comment>"):]
		}
	}

	// pointers: count '*' anywhere between type and name
	// Extract text between </type> and <name>
	between := ""
	if ti := strings.Index(work, "</type>"); ti >= 0 {
		rest := work[ti+len("</type>"):]
		if ni := strings.Index(rest, "<name>"); ni >= 0 {
			between = rest[:ni]
		}
	}
	mi.pointer = strings.Count(between, "*")
	// "const X* const*" => 2 pointers (both counted above)

	// array dimensions after </name>
	afterName := ""
	if ni := strings.Index(work, "</name>"); ni >= 0 {
		afterName = work[ni+len("</name>"):]
	}
	// resolve [<enum>NAME</enum>] or [N]
	dims := parseArrayDims(afterName, b.constInts)
	if len(dims) >= 1 {
		mi.arrayLen = dims[0]
	}
	if len(dims) >= 2 {
		mi.arrayLen2 = dims[1]
	}

	// bitfield ":N"
	if m := bitRe.FindStringSubmatch(afterName); m != nil {
		mi.bitwidth = atoiSafe(m[1])
	}

	// is base type a struct/union?
	if t, ok := b.types[cType]; ok {
		if t.Category == "struct" || t.Category == "union" {
			mi.isStruct = true
		}
	}
	return mi
}

var arrayDimRe = regexp.MustCompile(`\[\s*(?:<enum>([^<]+)</enum>|([0-9]+))\s*\]`)

func parseArrayDims(s string, consts map[string]string) []string {
	var dims []string
	for _, m := range arrayDimRe.FindAllStringSubmatch(s, -1) {
		if m[1] != "" {
			// constant name -> resolve to integer literal
			if v, ok := consts[m[1]]; ok {
				dims = append(dims, v)
			} else {
				dims = append(dims, m[1])
			}
		} else {
			dims = append(dims, m[2])
		}
	}
	return dims
}

// goFieldType returns the Go type string for a parsed member, applying the
// pointer/array rules.
func (b *Builder) goFieldType(mi memberInfo) string {
	// Pointers (one or more '*') become unsafe.Pointer per the spec rules,
	// EXCEPT char* which we keep as *byte for ergonomics? The task says any
	// pointer member -> unsafe.Pointer. Honor that uniformly.
	if mi.pointer > 0 {
		return "unsafe.Pointer"
	}
	base := b.goValueType(mi.cType)
	if mi.arrayLen != "" {
		if mi.arrayLen2 != "" {
			return "[" + mi.arrayLen + "][" + mi.arrayLen2 + "]" + base
		}
		return "[" + mi.arrayLen + "]" + base
	}
	return base
}

// goValueType returns the by-value Go type for a C type name.
func (b *Builder) goValueType(cType string) string {
	if g, ok := scalarCType(cType); ok {
		if g == "" {
			return "byte"
		}
		return g
	}
	return cType
}

// exportName converts a C identifier to an exported Go identifier. Vulkan member
// names are camelCase (pNext, sType); we keep them but upper-case the first rune
// so fields are exported.
func exportName(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = upper(r[0])
	out := string(r)
	if isGoKeyword(out) {
		out += "_"
	}
	return out
}

func upper(r rune) rune {
	if r >= 'a' && r <= 'z' {
		return r - 32
	}
	return r
}

func isGoKeyword(s string) bool {
	switch s {
	case "Type", "Range", "Func", "String", "Map":
		return false // exported, fine
	}
	switch s {
	case "type", "range", "func", "map", "string", "len", "cap", "select",
		"chan", "interface", "package", "import", "return", "var", "const":
		return true
	}
	return false
}

func atoiSafe(s string) int {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return n
		}
		n = n*10 + int(c-'0')
	}
	return n
}
