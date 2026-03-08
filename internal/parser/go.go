package parser

import (
	"go/ast"
	goparser "go/parser"
	"go/token"
	"strings"
	"unicode"
)

func init() {
	Register(&GoParser{})
}

// GoParser parses Go source files using the standard library's go/parser.
type GoParser struct{}

func (p *GoParser) Language() string    { return "Go" }
func (p *GoParser) Extensions() []string { return []string{".go"} }

func (p *GoParser) Parse(path string, content []byte) (*FileNode, error) {
	fset := token.NewFileSet()
	file, err := goparser.ParseFile(fset, path, content, goparser.ParseComments)
	if err != nil {
		return nil, err
	}

	node := &FileNode{
		Path:     path,
		Language: "Go",
	}

	// Extract imports
	for _, imp := range file.Imports {
		importPath := strings.Trim(imp.Path.Value, `"`)
		node.Imports = append(node.Imports, importPath)
	}

	// Extract functions and methods
	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			fn := FunctionDef{
				Name:     d.Name.Name,
				Line:     fset.Position(d.Pos()).Line,
				Exported: d.Name.IsExported(),
			}

			if d.Recv != nil && len(d.Recv.List) > 0 {
				fn.Receiver = typeString(d.Recv.List[0].Type)
			}

			if d.Type.Params != nil {
				for _, param := range d.Type.Params.List {
					fn.Parameters = append(fn.Parameters, typeString(param.Type))
				}
			}

			if d.Type.Results != nil {
				for _, result := range d.Type.Results.List {
					fn.Returns = append(fn.Returns, typeString(result.Type))
				}
			}

			node.Functions = append(node.Functions, fn)

			// Detect exported functions as exports
			if fn.Exported {
				name := fn.Name
				if fn.Receiver != "" {
					name = fn.Receiver + "." + name
				}
				node.Exports = append(node.Exports, name)
			}

		case *ast.GenDecl:
			if d.Tok == token.TYPE {
				for _, spec := range d.Specs {
					ts, ok := spec.(*ast.TypeSpec)
					if !ok {
						continue
					}

					td := TypeDef{
						Name:     ts.Name.Name,
						Line:     fset.Position(ts.Pos()).Line,
						Exported: ts.Name.IsExported(),
					}

					switch t := ts.Type.(type) {
					case *ast.StructType:
						td.Kind = "struct"
						if t.Fields != nil {
							for _, field := range t.Fields.List {
								for _, name := range field.Names {
									td.Fields = append(td.Fields, name.Name)
								}
							}
						}
					case *ast.InterfaceType:
						td.Kind = "interface"
						if t.Methods != nil {
							for _, method := range t.Methods.List {
								for _, name := range method.Names {
									td.Fields = append(td.Fields, name.Name)
								}
							}
						}
					default:
						td.Kind = "alias"
					}

					node.Types = append(node.Types, td)

					if td.Exported {
						node.Exports = append(node.Exports, td.Name)
					}
				}
			}
		}
	}

	// Detect API routes (common Go patterns)
	detectGoAPIRoutes(file, fset, node)

	return node, nil
}

func detectGoAPIRoutes(file *ast.File, fset *token.FileSet, node *FileNode) {
	ast.Inspect(file, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}

		method := strings.ToUpper(sel.Sel.Name)
		httpMethods := map[string]bool{
			"GET": true, "POST": true, "PUT": true, "DELETE": true,
			"PATCH": true, "HEAD": true, "OPTIONS": true,
			"HANDLE": true, "HANDLEFUNC": true,
		}

		if !httpMethods[method] {
			return true
		}

		if len(call.Args) >= 1 {
			if lit, ok := call.Args[0].(*ast.BasicLit); ok && lit.Kind == token.STRING {
				route := APIRoute{
					Method: method,
					Path:   strings.Trim(lit.Value, `"`),
					Line:   fset.Position(call.Pos()).Line,
				}
				if len(call.Args) >= 2 {
					route.Handler = exprName(call.Args[len(call.Args)-1])
				}
				node.APIRoutes = append(node.APIRoutes, route)
			}
		}

		return true
	})
}

func typeString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + typeString(t.X)
	case *ast.SelectorExpr:
		return typeString(t.X) + "." + t.Sel.Name
	case *ast.ArrayType:
		return "[]" + typeString(t.Elt)
	case *ast.MapType:
		return "map[" + typeString(t.Key) + "]" + typeString(t.Value)
	case *ast.InterfaceType:
		return "interface{}"
	case *ast.Ellipsis:
		return "..." + typeString(t.Elt)
	default:
		return "unknown"
	}
}

func exprName(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.SelectorExpr:
		return exprName(e.X) + "." + e.Sel.Name
	default:
		return ""
	}
}

// isExportedName checks if a Go name is exported (starts with uppercase).
func isExportedName(name string) bool {
	if len(name) == 0 {
		return false
	}
	return unicode.IsUpper(rune(name[0]))
}
