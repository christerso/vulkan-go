package main

import (
	"fmt"
	"go/format"
	"os"
)

const (
	defaultXML = "/usr/share/vulkan/registry/vk.xml"
	defaultOut = "vulkan"
)

func main() {
	xmlPath := defaultXML
	if v := os.Getenv("VK_XML"); v != "" {
		xmlPath = v
	}
	outDir := defaultOut
	if len(os.Args) > 1 {
		outDir = os.Args[1]
	}

	reg, err := parseRegistry(xmlPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "vkgen: parse:", err)
		os.Exit(1)
	}

	b := newBuilder(reg)
	b.collectScope()

	if err := os.MkdirAll(outDir, 0o755); err != nil {
		fmt.Fprintln(os.Stderr, "vkgen:", err)
		os.Exit(1)
	}
	if err := b.emitAll(outDir); err != nil {
		fmt.Fprintln(os.Stderr, "vkgen: emit:", err)
		os.Exit(1)
	}

	// stats
	fmt.Printf("vkgen: enums=%d bitmasks=%d structs=%d unions=%d handles=%d funcptrs=%d commands=%d\n",
		countCat(b, "enum"), countCat(b, "bitmask"), countCat(b, "struct"),
		countCat(b, "union"), countCat(b, "handle"), countCat(b, "funcpointer"),
		len(b.needCmd))
}

func countCat(b *Builder, cat string) int {
	seen := map[string]bool{}
	n := 0
	for i := range b.reg.Types.Type {
		t := &b.reg.Types.Type[i]
		name := typeName(t)
		if b.needType[name] && t.Category == cat && apiIncludesVulkan(t.API) && !seen[name] {
			seen[name] = true
			n++
		}
	}
	return n
}

// writeGoFile formats Go source and writes it. On format error it writes the raw
// source so the failure can be inspected.
func writeGoFile(path, src string) error {
	formatted, err := format.Source([]byte(src))
	if err != nil {
		os.WriteFile(path, []byte(src), 0o644)
		return fmt.Errorf("format %s: %w", path, err)
	}
	return os.WriteFile(path, formatted, 0o644)
}
