package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/crystal"
)

var INDENT_SIZE = 2

type Formatter struct {
	b      *strings.Builder
	source []byte
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: crystalfmt <file.cr>")
		os.Exit(1)
	}

	filename := os.Args[1]
	source, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("Failed to read file: %v\n", err)
		os.Exit(1)
	}

	// Set up Tree-sitter parser
	parser := sitter.NewParser()
	parser.SetLanguage(crystal.GetLanguage())
	tree, err := parser.ParseCtx(context.Background(), nil, source)
	if err != nil {
		fmt.Println("--- Parsing failed:", err)
	}

	f := Formatter{
		b:      &strings.Builder{},
		source: source,
	}

	f.formatNode(tree.RootNode(), 0)

	formatted := f.b.String()

	// Overwrite original file with formatted output
	err = os.WriteFile(filename, []byte(formatted), 0644)
	if err != nil {
		fmt.Printf("Failed to write file: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(formatted))
}

func foreachChild(node *sitter.Node, fn func(ch *sitter.Node)) {
	for idx := range node.ChildCount() {
		ch := node.Child(int(idx))
		fn(ch)
	}
}

func (f *Formatter) formatMethod(node *sitter.Node, indent int) {
	f.b.WriteByte('\n')
	foreachChild(node, func(ch *sitter.Node) {
		switch ch.Type() {
		case "def":
			f.writeString("def")
			f.b.WriteByte(' ')
		case "identifier":
			f.writeContent(ch)
		case "(":
			f.b.WriteByte('(')
		case ")":
			f.writeString(")\n")
		case "expressions":
			f.b.WriteByte('\n')
			f.formatNode(ch, indent+INDENT_SIZE)
		case "param_list":
			f.formatParams(ch)
		case "end":
			f.b.WriteString("end")
		}
	})
}

func (f *Formatter) formatRequire(node *sitter.Node) {
	foreachChild(node, func(ch *sitter.Node) {
		switch ch.Type() {
		case "require":
			f.b.WriteString("require")
			f.b.WriteByte(' ')
		case "string":
			f.formatString(ch)
		}
	})

	if node.NextSibling().Type() != "require" {
		f.b.WriteString("\n")
	}
}

func (f *Formatter) formatString(node *sitter.Node) {
	foreachChild(node, func(ch *sitter.Node) {
		switch ch.Type() {
		case `"`:
			f.writeByte('"')
		case "literal_content":
			f.formatLiteral(ch)
		case "interpolation":
			f.formatInterpolation(ch)
		}
	})
}

func (f *Formatter) formatInterpolation(node *sitter.Node) {
	foreachChild(node, func(ch *sitter.Node) {
		switch ch.Type() {
		case "#{", "}":
			f.writeContent(ch)
		default:
			f.formatNode(ch, 0)
		}
	})
}

func (f *Formatter) formatLiteral(node *sitter.Node) {
	f.b.WriteString(node.Content(f.source))
}

func (f *Formatter) formatBlock(node *sitter.Node, indent int) {
	foreachChild(node, func(ch *sitter.Node) {
		switch ch.Type() {
		case "do":
			f.b.WriteString(" do ")
		case "param_list":
			f.formatParams(ch)
		case "expressions":
			f.b.WriteByte('\n')
			f.formatNode(ch, indent+INDENT_SIZE)
		default:
			f.writeContent(ch)
		}
	})

}

func (f *Formatter) formatParams(node *sitter.Node) {
	foreachChild(node, func(ch *sitter.Node) {
		switch ch.Type() {
		case "param":
			f.writeContent(ch)
		case ",":
			f.b.WriteString(", ")
		}
	})
}

func (f *Formatter) formatAssign(node *sitter.Node, indent int) {
	left := node.ChildByFieldName("lhs")
	right := node.ChildByFieldName("rhs")

	f.writeContent(left)

	f.b.WriteString(" = ")

	// Format the right hand side if it exists
	if right != nil {
		f.formatNode(right, indent)
	}
}

func (f *Formatter) formatCall(node *sitter.Node, indent int) {
	foreachChild(node, func(ch *sitter.Node) {
		switch ch.Type() {
		case "call":
			f.formatCall(ch, indent)
		case "argument_list":
			f.formatArguments(ch, indent)
		default:
			f.formatNode(ch, indent)
		}
	})
}

func (f *Formatter) formatArguments(node *sitter.Node, indent int) {
	if node.Child(0).Type() != "(" {
		f.b.WriteByte(' ')
	}

	foreachChild(node, func(ch *sitter.Node) {
		switch ch.Type() {
		case "(", ")":
			f.writeContent(ch)
		case ",":
			f.b.WriteString(", ")
		default:
			f.formatNode(ch, indent)
		}
	})
}

func (f *Formatter) formatExpressions(node *sitter.Node, indent int) {
	foreachChild(node, func(ch *sitter.Node) {
		f.writeIndent(indent)
		f.formatNode(ch, indent)
		f.b.WriteByte('\n')
	})
}

// Recursive function to format the syntax tree
func (f *Formatter) formatNode(node *sitter.Node, indent int) {

	// fmt.Println(strings.Repeat(" ", indent) + node.Type())
	// if node.ChildCount() == 0 {
	// 	return
	// }
	//
	// foreachChild(node, func(ch *sitter.Node) {
	// 	f.formatNode(ch, indent+INDENT_SIZE)
	// })
	//
	// return

	fmt.Println("--- ", node.Type())

	switch node.Type() {
	case "method_def":
		f.formatMethod(node, indent)

	case "expressions":
		f.formatExpressions(node, indent)

	case "require":
		f.formatRequire(node)

	case "assign":
		f.formatAssign(node, indent)

	case "call":
		f.formatCall(node, indent)

	case "block":
		f.formatBlock(node, indent)

	case "string":
		f.formatString(node)

	case "integer":
		f.formatLiteral(node)

	case "identifier":
		f.writeContent(node)

	default:
		// Fallback to just printing the raw source content for unknown types
		f.writeContent(node)
	}
}

func (f *Formatter) writeIndent(indent int) {
	for range indent {
		f.b.WriteByte(' ')
	}
}

func (f *Formatter) writeContent(node *sitter.Node) {
	f.b.WriteString(node.Content(f.source))
}

func (f *Formatter) writeString(str string, a ...any) {
	fmt.Fprintf(f.b, str, a...)
}

func (f *Formatter) writeByte(b byte) {
	f.b.WriteByte(b)
}
