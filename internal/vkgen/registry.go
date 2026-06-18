// Command vkgen generates a complete purego-based Vulkan binding for Go from
// the Vulkan XML registry. This file holds the registry model and XML parsing.
package main

import (
	"encoding/xml"
	"os"
	"strings"
)

// ---- raw XML model ----

type xmlRegistry struct {
	XMLName    xml.Name       `xml:"registry"`
	Types      xmlTypes       `xml:"types"`
	Enums      []xmlEnums     `xml:"enums"`
	Commands   xmlCommands    `xml:"commands"`
	Features   []xmlFeature   `xml:"feature"`
	Extensions xmlExtensions  `xml:"extensions"`
}

type xmlTypes struct {
	Type []xmlType `xml:"type"`
}

type xmlType struct {
	Category   string      `xml:"category,attr"`
	API        string      `xml:"api,attr"`
	Name       string      `xml:"name,attr"`
	NameInner  string      `xml:"name"`     // <name> child element
	TypeInner  string      `xml:"type"`     // first <type> child (basetype underlying)
	Requires   string      `xml:"requires,attr"`
	Alias      string      `xml:"alias,attr"`
	BitValues  string      `xml:"bitvalues,attr"`
	Parent     string      `xml:"parent,attr"`
	Members    []xmlMember `xml:"member"`
	Proto      xmlProto    `xml:"proto"`  // funcpointer return
	Params     []xmlParam  `xml:"param"`  // funcpointer params
	Raw        string      `xml:",innerxml"`
}

type xmlMember struct {
	Type     string `xml:"type"`
	Name     string `xml:"name"`
	Enum     string `xml:"enum"` // array size constant
	Values   string `xml:"values,attr"`
	API      string `xml:"api,attr"`
	Raw      string `xml:",innerxml"`
}

type xmlProto struct {
	Type string `xml:"type"`
	Name string `xml:"name"`
}

type xmlParam struct {
	Type string `xml:"type"`
	Name string `xml:"name"`
	API  string `xml:"api,attr"`
	Raw  string `xml:",innerxml"`
}

type xmlEnums struct {
	Name     string    `xml:"name,attr"`
	Type     string    `xml:"type,attr"`
	BitWidth string    `xml:"bitwidth,attr"`
	Enum     []xmlEnum `xml:"enum"`
}

type xmlEnum struct {
	Name      string `xml:"name,attr"`
	Value     string `xml:"value,attr"`
	BitPos    string `xml:"bitpos,attr"`
	Offset    string `xml:"offset,attr"`
	ExtNumber string `xml:"extnumber,attr"`
	Extends   string `xml:"extends,attr"`
	Dir       string `xml:"dir,attr"`
	Alias     string `xml:"alias,attr"`
	Type      string `xml:"type,attr"`
	Comment   string `xml:"comment,attr"`
}

type xmlCommands struct {
	Command []xmlCommand `xml:"command"`
}

type xmlCommand struct {
	Proto  xmlProto   `xml:"proto"`
	Params []xmlParam `xml:"param"`
	Name   string     `xml:"name,attr"`  // only for alias form
	Alias  string     `xml:"alias,attr"`
}

type xmlFeature struct {
	API     string       `xml:"api,attr"`
	Name    string       `xml:"name,attr"`
	Number  string       `xml:"number,attr"`
	Require []xmlRequire `xml:"require"`
}

type xmlExtensions struct {
	Extension []xmlExtension `xml:"extension"`
}

type xmlExtension struct {
	Name      string       `xml:"name,attr"`
	Number    string       `xml:"number,attr"`
	Type      string       `xml:"type,attr"`
	Supported string       `xml:"supported,attr"`
	Platform  string       `xml:"platform,attr"`
	Require   []xmlRequire `xml:"require"`
}

type xmlRequire struct {
	Depends string       `xml:"depends,attr"`
	Type    []xmlRefName `xml:"type"`
	Enum    []xmlEnum    `xml:"enum"`
	Command []xmlRefName `xml:"command"`
}

type xmlRefName struct {
	Name string `xml:"name,attr"`
}

func parseRegistry(path string) (*xmlRegistry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var reg xmlRegistry
	if err := xml.Unmarshal(data, &reg); err != nil {
		return nil, err
	}
	return &reg, nil
}

// apiIncludesVulkan reports whether a comma-separated api attribute selects the
// default "vulkan" authoring profile.
func apiIncludesVulkan(api string) bool {
	if api == "" {
		return true
	}
	for _, p := range strings.Split(api, ",") {
		if p == "vulkan" {
			return true
		}
	}
	return false
}
