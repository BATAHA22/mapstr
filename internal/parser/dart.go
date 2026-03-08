package parser

import (
	"regexp"
	"strings"
)

func init() {
	Register(&DartParser{})
}

// DartParser parses Dart and Flutter source files using regex-based extraction.
type DartParser struct{}

func (p *DartParser) Language() string     { return "Dart" }
func (p *DartParser) Extensions() []string { return []string{".dart"} }

var (
	dartImportRe = regexp.MustCompile(`(?m)^import\s+['"]([^'"]+)['"]`)
	dartExportRe = regexp.MustCompile(`(?m)^export\s+['"]([^'"]+)['"]`)

	dartFuncRe  = regexp.MustCompile(`(?m)^(?:\w+\s+)*(\w+)\s*\([^)]*\)\s*(?:async\s*)?[{=]`)
	dartClassRe = regexp.MustCompile(`(?m)^(?:abstract\s+)?class\s+(\w+)(?:\s+extends\s+(\w+))?(?:\s+(?:with|implements)\s+[^{]+)?`)
	dartEnumRe  = regexp.MustCompile(`(?m)^enum\s+(\w+)`)
	dartMixinRe = regexp.MustCompile(`(?m)^mixin\s+(\w+)`)

	// Flutter route patterns
	dartRouteRe = regexp.MustCompile(`(?m)['"](/[^'"]*)['"]\s*:\s*(?:\(.*?\)\s*=>)?\s*(\w+)`)
)

func (p *DartParser) Parse(path string, content []byte) (*FileNode, error) {
	src := string(content)
	lines := strings.Split(src, "\n")

	node := &FileNode{
		Path:     path,
		Language: "Dart",
	}

	// Imports
	for _, m := range dartImportRe.FindAllStringSubmatch(src, -1) {
		node.Imports = append(node.Imports, m[1])
	}
	node.Imports = dedupe(node.Imports)

	// Exports (re-exports)
	for _, m := range dartExportRe.FindAllStringSubmatch(src, -1) {
		node.Exports = append(node.Exports, m[1])
	}

	// Functions (top-level only — lines starting without indentation)
	for i, line := range lines {
		if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
			continue
		}
		if fm := dartFuncRe.FindStringSubmatch(line); fm != nil {
			name := fm[1]
			// Skip keywords that look like functions
			if name == "if" || name == "for" || name == "while" || name == "switch" || name == "catch" || name == "class" || name == "return" {
				continue
			}
			exported := !strings.HasPrefix(name, "_")
			node.Functions = append(node.Functions, FunctionDef{
				Name:     name,
				Line:     i + 1,
				Exported: exported,
			})
			if exported {
				node.Exports = append(node.Exports, name)
			}
		}
	}

	// Classes
	for _, m := range dartClassRe.FindAllStringSubmatchIndex(src, -1) {
		name := src[m[2]:m[3]]
		line := lineNumber(lines, m[0])
		exported := !strings.HasPrefix(name, "_")
		node.Types = append(node.Types, TypeDef{
			Name:     name,
			Kind:     "class",
			Line:     line,
			Exported: exported,
		})
		if exported {
			node.Exports = append(node.Exports, name)
		}
	}

	// Enums
	for _, m := range dartEnumRe.FindAllStringSubmatchIndex(src, -1) {
		name := src[m[2]:m[3]]
		line := lineNumber(lines, m[0])
		exported := !strings.HasPrefix(name, "_")
		node.Types = append(node.Types, TypeDef{
			Name:     name,
			Kind:     "enum",
			Line:     line,
			Exported: exported,
		})
		if exported {
			node.Exports = append(node.Exports, name)
		}
	}

	// Mixins
	for _, m := range dartMixinRe.FindAllStringSubmatchIndex(src, -1) {
		name := src[m[2]:m[3]]
		line := lineNumber(lines, m[0])
		exported := !strings.HasPrefix(name, "_")
		node.Types = append(node.Types, TypeDef{
			Name:     name,
			Kind:     "mixin",
			Line:     line,
			Exported: exported,
		})
		if exported {
			node.Exports = append(node.Exports, name)
		}
	}

	// Flutter routes (MaterialApp route maps)
	for _, m := range dartRouteRe.FindAllStringSubmatchIndex(src, -1) {
		routePath := src[m[2]:m[3]]
		handler := src[m[4]:m[5]]
		line := lineNumber(lines, m[0])
		node.APIRoutes = append(node.APIRoutes, APIRoute{
			Method:  "GET",
			Path:    routePath,
			Handler: handler,
			Line:    line,
		})
	}

	return node, nil
}
