package parser

import (
	"regexp"
	"strings"
)

func init() {
	Register(&JSParser{})
}

// JSParser parses JavaScript and TypeScript files using regex-based extraction.
// This provides reasonable coverage without requiring Tree-sitter CGo bindings,
// keeping the binary simple and cross-compilable.
type JSParser struct{}

func (p *JSParser) Language() string    { return "JavaScript" }
func (p *JSParser) Extensions() []string { return []string{".js", ".jsx", ".ts", ".tsx", ".mjs", ".cjs"} }

var (
	// Imports
	jsImportRe     = regexp.MustCompile(`(?m)^import\s+.*?from\s+['"]([^'"]+)['"]`)
	jsRequireRe    = regexp.MustCompile(`(?m)require\(['"]([^'"]+)['"]\)`)
	jsDynImportRe  = regexp.MustCompile(`import\(['"]([^'"]+)['"]\)`)

	// Exports
	jsExportNamedRe   = regexp.MustCompile(`(?m)^export\s+(?:const|let|var|function|class|async\s+function)\s+(\w+)`)
	jsExportDefaultRe = regexp.MustCompile(`(?m)^export\s+default\s+(?:function|class)?\s*(\w*)`)
	jsModuleExportsRe = regexp.MustCompile(`(?m)module\.exports\s*=`)

	// Functions
	jsFuncDeclRe     = regexp.MustCompile(`(?m)^(?:export\s+)?(?:async\s+)?function\s+(\w+)\s*\(([^)]*)\)`)
	jsArrowFuncRe    = regexp.MustCompile(`(?m)^(?:export\s+)?(?:const|let|var)\s+(\w+)\s*=\s*(?:async\s+)?\([^)]*\)\s*=>`)

	// Classes and types
	jsClassRe        = regexp.MustCompile(`(?m)^(?:export\s+)?(?:abstract\s+)?class\s+(\w+)`)
	tsInterfaceRe    = regexp.MustCompile(`(?m)^(?:export\s+)?interface\s+(\w+)`)
	tsTypeRe         = regexp.MustCompile(`(?m)^(?:export\s+)?type\s+(\w+)`)

	// API routes (Express-style)
	jsRouteRe = regexp.MustCompile(`(?m)(?:app|router)\.(get|post|put|delete|patch|head|options)\s*\(\s*['"]([^'"]+)['"]`)
)

func (p *JSParser) Parse(path string, content []byte) (*FileNode, error) {
	src := string(content)
	lines := strings.Split(src, "\n")

	node := &FileNode{
		Path:     path,
		Language: p.detectLanguage(path),
	}

	// Parse imports
	for _, m := range jsImportRe.FindAllStringSubmatch(src, -1) {
		node.Imports = append(node.Imports, m[1])
	}
	for _, m := range jsRequireRe.FindAllStringSubmatch(src, -1) {
		node.Imports = append(node.Imports, m[1])
	}
	for _, m := range jsDynImportRe.FindAllStringSubmatch(src, -1) {
		node.Imports = append(node.Imports, m[1])
	}
	node.Imports = dedupe(node.Imports)

	// Parse exports
	for _, m := range jsExportNamedRe.FindAllStringSubmatch(src, -1) {
		node.Exports = append(node.Exports, m[1])
	}
	for _, m := range jsExportDefaultRe.FindAllStringSubmatch(src, -1) {
		name := m[1]
		if name == "" {
			name = "default"
		}
		node.Exports = append(node.Exports, name)
	}
	if jsModuleExportsRe.MatchString(src) {
		node.Exports = append(node.Exports, "module.exports")
	}

	// Parse functions
	for _, m := range jsFuncDeclRe.FindAllStringSubmatchIndex(src, -1) {
		name := src[m[2]:m[3]]
		line := lineNumber(lines, m[0])
		node.Functions = append(node.Functions, FunctionDef{
			Name:     name,
			Line:     line,
			Exported: isJSExported(name, src),
		})
	}
	for _, m := range jsArrowFuncRe.FindAllStringSubmatchIndex(src, -1) {
		name := src[m[2]:m[3]]
		line := lineNumber(lines, m[0])
		node.Functions = append(node.Functions, FunctionDef{
			Name:     name,
			Line:     line,
			Exported: isJSExported(name, src),
		})
	}

	// Parse classes
	for _, m := range jsClassRe.FindAllStringSubmatchIndex(src, -1) {
		name := src[m[2]:m[3]]
		line := lineNumber(lines, m[0])
		node.Types = append(node.Types, TypeDef{
			Name:     name,
			Kind:     "class",
			Line:     line,
			Exported: isJSExported(name, src),
		})
	}

	// Parse TypeScript interfaces and types
	for _, m := range tsInterfaceRe.FindAllStringSubmatchIndex(src, -1) {
		name := src[m[2]:m[3]]
		line := lineNumber(lines, m[0])
		node.Types = append(node.Types, TypeDef{
			Name:     name,
			Kind:     "interface",
			Line:     line,
			Exported: isJSExported(name, src),
		})
	}
	for _, m := range tsTypeRe.FindAllStringSubmatchIndex(src, -1) {
		name := src[m[2]:m[3]]
		line := lineNumber(lines, m[0])
		node.Types = append(node.Types, TypeDef{
			Name:     name,
			Kind:     "type",
			Line:     line,
			Exported: isJSExported(name, src),
		})
	}

	// Parse API routes
	for _, m := range jsRouteRe.FindAllStringSubmatchIndex(src, -1) {
		method := strings.ToUpper(src[m[2]:m[3]])
		routePath := src[m[4]:m[5]]
		line := lineNumber(lines, m[0])
		node.APIRoutes = append(node.APIRoutes, APIRoute{
			Method: method,
			Path:   routePath,
			Line:   line,
		})
	}

	return node, nil
}

func (p *JSParser) detectLanguage(path string) string {
	ext := strings.ToLower(path)
	if strings.HasSuffix(ext, ".ts") || strings.HasSuffix(ext, ".tsx") {
		return "TypeScript"
	}
	return "JavaScript"
}

func isJSExported(name, src string) bool {
	exportPattern := regexp.MustCompile(`(?m)^export\s+.*\b` + regexp.QuoteMeta(name) + `\b`)
	return exportPattern.MatchString(src)
}

func lineNumber(lines []string, byteOffset int) int {
	count := 0
	for i, line := range lines {
		count += len(line) + 1 // +1 for newline
		if count > byteOffset {
			return i + 1
		}
	}
	return len(lines)
}

func dedupe(ss []string) []string {
	seen := map[string]bool{}
	var result []string
	for _, s := range ss {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}
	return result
}
