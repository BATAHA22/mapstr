package graph

import (
	"fmt"
	"strings"
)

// ToMermaid converts a DependencyGraph to a Mermaid diagram string.
func (g *DependencyGraph) ToMermaid() string {
	var b strings.Builder

	b.WriteString("graph TD\n")

	// Write nodes with styling based on type
	for _, node := range g.Nodes {
		switch node.Type {
		case "module":
			fmt.Fprintf(&b, "    %s[%s]\n", node.ID, escapeLabel(node.Label))
		case "type":
			fmt.Fprintf(&b, "    %s{{%s}}\n", node.ID, escapeLabel(node.Label))
		case "api":
			fmt.Fprintf(&b, "    %s([%s])\n", node.ID, escapeLabel(node.Label))
		case "function":
			fmt.Fprintf(&b, "    %s(%s)\n", node.ID, escapeLabel(node.Label))
		default:
			fmt.Fprintf(&b, "    %s[%s]\n", node.ID, escapeLabel(node.Label))
		}
	}

	b.WriteString("\n")

	// Write edges
	for _, edge := range g.Edges {
		switch edge.Type {
		case "imports":
			fmt.Fprintf(&b, "    %s --> %s\n", edge.From, edge.To)
		case "calls":
			fmt.Fprintf(&b, "    %s -.-> %s\n", edge.From, edge.To)
		case "extends":
			fmt.Fprintf(&b, "    %s ==> %s\n", edge.From, edge.To)
		case "implements":
			fmt.Fprintf(&b, "    %s -.->|implements| %s\n", edge.From, edge.To)
		default:
			fmt.Fprintf(&b, "    %s --> %s\n", edge.From, edge.To)
		}
	}

	// Add style classes
	b.WriteString("\n")
	b.WriteString("    classDef module fill:#e1f5fe,stroke:#01579b\n")
	b.WriteString("    classDef type fill:#f3e5f5,stroke:#4a148c\n")
	b.WriteString("    classDef api fill:#e8f5e9,stroke:#1b5e20\n")

	// Apply styles
	for _, node := range g.Nodes {
		switch node.Type {
		case "module":
			fmt.Fprintf(&b, "    class %s module\n", node.ID)
		case "type":
			fmt.Fprintf(&b, "    class %s type\n", node.ID)
		case "api":
			fmt.Fprintf(&b, "    class %s api\n", node.ID)
		}
	}

	return b.String()
}

func escapeLabel(s string) string {
	s = strings.ReplaceAll(s, `"`, `#quot;`)
	s = strings.ReplaceAll(s, "<", `#lt;`)
	s = strings.ReplaceAll(s, ">", `#gt;`)
	return `"` + s + `"`
}
