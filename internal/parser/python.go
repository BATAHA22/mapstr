package parser

import (
	"regexp"
	"strings"
)

func init() {
	Register(&PythonParser{})
}

// PythonParser parses Python source files using regex-based extraction.
type PythonParser struct{}

func (p *PythonParser) Language() string    { return "Python" }
func (p *PythonParser) Extensions() []string { return []string{".py"} }

var (
	pyImportRe     = regexp.MustCompile(`(?m)^import\s+(\S+)`)
	pyFromImportRe = regexp.MustCompile(`(?m)^from\s+(\S+)\s+import`)

	pyFuncRe  = regexp.MustCompile(`(?m)^(?:async\s+)?def\s+(\w+)\s*\(([^)]*)\)`)
	pyClassRe = regexp.MustCompile(`(?m)^class\s+(\w+)(?:\(([^)]*)\))?:`)

	// Flask/FastAPI route decorators
	pyRouteRe = regexp.MustCompile(`(?m)^@\w+\.(get|post|put|delete|patch|route)\s*\(\s*['"]([^'"]+)['"]`)

	// __all__ exports
	pyAllRe = regexp.MustCompile(`(?m)^__all__\s*=\s*\[([^\]]+)\]`)
)

func (p *PythonParser) Parse(path string, content []byte) (*FileNode, error) {
	src := string(content)
	lines := strings.Split(src, "\n")

	node := &FileNode{
		Path:     path,
		Language: "Python",
	}

	// Imports
	for _, m := range pyImportRe.FindAllStringSubmatch(src, -1) {
		node.Imports = append(node.Imports, m[1])
	}
	for _, m := range pyFromImportRe.FindAllStringSubmatch(src, -1) {
		node.Imports = append(node.Imports, m[1])
	}
	node.Imports = dedupe(node.Imports)

	// Functions
	pendingRoute := (*APIRoute)(nil)
	for i, line := range lines {
		// Check for route decorators
		if rm := pyRouteRe.FindStringSubmatch(line); rm != nil {
			pendingRoute = &APIRoute{
				Method: strings.ToUpper(rm[1]),
				Path:   rm[2],
				Line:   i + 1,
			}
			continue
		}

		if fm := pyFuncRe.FindStringSubmatch(line); fm != nil {
			name := fm[1]
			exported := !strings.HasPrefix(name, "_")
			node.Functions = append(node.Functions, FunctionDef{
				Name:     name,
				Line:     i + 1,
				Exported: exported,
			})
			if exported {
				node.Exports = append(node.Exports, name)
			}

			if pendingRoute != nil {
				pendingRoute.Handler = name
				node.APIRoutes = append(node.APIRoutes, *pendingRoute)
				pendingRoute = nil
			}
		} else {
			pendingRoute = nil
		}
	}

	// Classes
	for _, m := range pyClassRe.FindAllStringSubmatchIndex(src, -1) {
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

	// __all__ overrides exports
	if allMatch := pyAllRe.FindStringSubmatch(src); allMatch != nil {
		node.Exports = nil
		items := strings.Split(allMatch[1], ",")
		for _, item := range items {
			item = strings.TrimSpace(item)
			item = strings.Trim(item, `'"`)
			if item != "" {
				node.Exports = append(node.Exports, item)
			}
		}
	}

	return node, nil
}
