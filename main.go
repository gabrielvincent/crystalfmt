package main

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"os"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/crystal"
)

var INDENT_SIZE = 4

type Formatter struct {
	strBuilder         *strings.Builder
	source             []byte
	lineStartPositions []int
	indentSize         int
	err                error
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
		strBuilder:         &strings.Builder{},
		source:             source,
		lineStartPositions: buildLineStartPositions(source),
		indentSize:         INDENT_SIZE,
	}

	f.formatNode(tree.RootNode(), 0)

	formatted := f.strBuilder.String()

	var shouldWrite bool
	for _, arg := range os.Args {
		if arg == "--write" || arg == "-w" {
			shouldWrite = true
			break
		}
	}
	if f.err != nil {
		shouldWrite = false
		formatted = string(source)
	}

	if shouldWrite {
		err = os.WriteFile(filename, []byte(formatted), 0644)
		if err != nil {
			fmt.Printf("Failed to write file: %v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Println(formatted)
	}

}

func (f *Formatter) formatMethod(node *sitter.Node, indent int) {
	nameNode := node.ChildByFieldName("name")

	f.writeString("def ")
	f.formatNode(nameNode, indent)

	paramsNode := node.ChildByFieldName("params")
	if paramsNode != nil {
		f.writeByte('(')
		f.formatNode(paramsNode, indent)
		f.writeByte(')')
	}

	for ch := range eachChild(node) {
		switch ch.Type() {
		case "comment":
			f.writeLF()
			f.writeIndent(indent + f.indentSize)
			f.formatNode(ch, indent+f.indentSize)
		case "expressions":
			f.writeLF()
			f.formatNode(ch, indent+f.indentSize)
		case "(", ")":
			if paramsNode == nil {
				f.writeContent(ch)
			}
		}
	}

	f.writeLF()
	f.writeIndent(indent)
	f.writeString("end")

}

func (f *Formatter) formatClass(node *sitter.Node, indent int) {
	nameNode := node.ChildByFieldName("name")

	f.writeString("class")
	f.writeByte(' ')
	f.formatNode(nameNode, indent)

	if superclassNode := node.ChildByFieldName("superclass"); superclassNode != nil {
		f.writeString(" < ")
		f.formatNode(superclassNode, indent)
	}

	if bodyNode := node.ChildByFieldName("body"); bodyNode != nil {
		f.writeLF()
		f.formatNode(bodyNode, indent+f.indentSize)
	}

	f.writeLF()
	f.writeIndent(indent)
	f.writeString("end")
}

func (f *Formatter) formatRequire(node *sitter.Node) {
	for ch := range eachChild(node) {
		switch ch.Type() {
		case "require":
			f.strBuilder.WriteString("require")
			f.strBuilder.WriteByte(' ')
		case "string":
			f.formatString(ch)
		}
	}
}

func (f *Formatter) formatString(node *sitter.Node) {
	for ch := range eachChild(node) {
		switch ch.Type() {
		case `"`:
			f.writeByte('"')
		case "literal_content":
			f.formatLiteral(ch)
		case "interpolation":
			f.formatInterpolation(ch)
		}
	}
}

func (f *Formatter) formatInterpolation(node *sitter.Node) {
	for ch := range eachChild(node) {
		switch ch.Type() {
		case "#{", "}":
			f.writeContent(ch)
		default:
			f.formatNode(ch, 0)
		}
	}
}

func (f *Formatter) formatLiteral(node *sitter.Node) {
	f.writeContent(node)
}

func (f *Formatter) formatComment(node *sitter.Node) {
	cmtVal := f.getContent(node)

	// If the comment doesn't have a space after '#'
	if len(cmtVal) > 1 && cmtVal[0] == '#' && cmtVal[1] != ' ' {
		f.writeByte('#')
		f.writeByte(' ')
		if len(cmtVal) > 1 {
			f.writeString(cmtVal[1:])
		}
	} else {
		f.writeString(cmtVal)
	}
}

func (f *Formatter) formatBlockOneLiner(node *sitter.Node, indent int) {
	bodyNode := node.ChildByFieldName("body")
	for ch := range eachChild(node) {
		switch ch.Type() {
		case "do":
			f.writeByte(' ')
			f.writeContent(ch)
		case "|":
			next := ch.NextSibling()
			switch next.Type() {
			case "param_list":
				f.writeByte(' ')
				f.writeContent(ch)
			case "expressions":
				f.formatExpressions(ch, 0, false)
				f.writeContent(ch)
			}
		case "{":
			f.writeByte(' ')
			f.writeContent(ch)
		case "}":
			f.writeContent(ch)
		case "expressions":
			f.writeByte(' ')
			f.formatExpressions(ch, indent, false)
			f.writeByte(' ')
		case "end":
			if bodyNode == nil {
				f.writeByte(' ')
			}
			f.writeContent(ch)
		default:
			f.formatNode(ch, indent)
		}
	}
}

func (f *Formatter) formatBlock(node *sitter.Node, indent int) {
	var blockDelimiterStart int
	var blockDelimiterEnd int
	for ch := range eachChild(node) {
		switch ch.Type() {
		case "do", "{":
			blockDelimiterStart = getAbsPosition(ch.Range().EndPoint, f.lineStartPositions)
		case "end", "}":
			blockDelimiterEnd = getAbsPosition(ch.Range().StartPoint, f.lineStartPositions)
		}
	}

	var isMultiline bool
	idx := blockDelimiterStart
	for idx < blockDelimiterEnd {
		if f.source[idx] == '\n' {
			isMultiline = true
			break
		}
		idx++
	}

	if !isMultiline {
		f.formatBlockOneLiner(node, indent)
		return
	}

	for ch := range eachChild(node) {
		switch ch.Type() {
		case "do":
			f.writeByte(' ')
			f.writeContent(ch)
		case "|":
			next := ch.NextSibling()
			switch next.Type() {
			case "param_list":
				f.writeByte(' ')
				f.writeContent(ch)
			case "expressions":
				f.formatExpressions(ch, 0, false)
				f.writeContent(ch)
			}
		case "{":
			f.writeByte(' ')
			f.writeContent(ch)
		case "}":
			f.writeContent(ch)
		case "expressions":
			f.writeLF()
			f.formatExpressions(ch, indent+f.indentSize, true)
			f.writeLF()
		case "end":
			f.writeContent(ch)
		default:
			f.formatNode(ch, indent)
		}
	}
}

func (f *Formatter) formatIdentifier(node *sitter.Node) {
	f.writeContent(node)
}

func (f *Formatter) formatSymbol(node *sitter.Node) {
	f.writeContent(node)
}

func (f *Formatter) formatBlockParam(node *sitter.Node) {
	for ch := range eachChild(node) {
		switch ch.Type() {
		case "&", "identifier":
			f.writeContent(ch)
		case "proc_type":
			f.writeString(" : ")
			f.formatProcType(ch)
		}
	}
}

func (f *Formatter) formatParam(node *sitter.Node) {
	if nameNode := node.ChildByFieldName("name"); nameNode != nil {
		f.formatNode(nameNode, 0)
	}

	for typeNode := range eachChildByFieldName(node, "type") {
		switch typeNode.Type() {
		case ":":
			f.writeString(" : ")
		default:
			f.formatNode(typeNode, 0)
		}
	}

	if defaultAssignNode := node.ChildByFieldName("default"); defaultAssignNode != nil {
		f.writeString(" = ")
		if defaultRhsNode := defaultAssignNode.NextNamedSibling(); defaultRhsNode != nil {
			switch defaultRhsNode.Type() {
			case "expressions":
				f.formatExpressions(defaultRhsNode, 0, false)
			default:
				f.formatNode(defaultRhsNode, 0)
			}
		}
	}
}

func (f *Formatter) formatParams(node *sitter.Node) {
	for ch := range eachChild(node) {
		switch ch.Type() {
		case "param":
			f.formatParam(ch)
		case "block_param":
			f.formatBlockParam(ch)
		case ",":
			f.writeString(", ")
		}

	}
}

func (f *Formatter) formatAssign(node *sitter.Node, indent int) {
	left := node.ChildByFieldName("lhs")
	right := node.ChildByFieldName("rhs")

	f.writeContent(left)

	f.strBuilder.WriteString(" = ")

	// Format the right hand side if it exists
	if right != nil {
		f.formatNode(right, indent)
	}
}

func (f *Formatter) formatCall(node *sitter.Node, indent int) {
	for ch := range eachChild(node) {
		switch ch.Type() {
		case "expressions":
			f.formatExpressions(ch, indent, false)
		case "argument_list":
			var prevType string
			var firstChildType string

			if ch.PrevSibling() != nil {
				prevType = ch.PrevSibling().Type()
			}
			if ch.ChildCount() > 0 {
				firstChildType = ch.Child(0).Type()
			}

			// Check for a method call not using parentheses. If this is the case,
			// a space is added between the identifier and the next character
			if prevType == "identifier" && firstChildType != "(" {
				f.writeByte(' ')
			}
			fallthrough
		default:
			f.formatNode(ch, indent)
		}
	}
}

func (f *Formatter) formatNamedExpr(node *sitter.Node) {
	for ch := range eachChild(node) {
		switch ch.Type() {
		case ":":
			f.writeString(": ")
		default:
			f.formatNode(ch, 0)
		}

	}
}

func (f *Formatter) formatArguments(node *sitter.Node, indent int) {
	for ch := range eachChild(node) {
		switch ch.Type() {
		case ",":
			f.strBuilder.WriteString(", ")
		case "expressions":
			f.formatExpressions(ch, 0, false)
		case "named_expr":
			f.formatNamedExpr(ch)
		default:
			f.formatNode(ch, indent)
		}

	}
}

func (f *Formatter) formatExpressions(node *sitter.Node, indent int, multiline bool) {
	for ch := range eachChild(node) {
		if prev := ch.PrevSibling(); prev != nil {
			prevEnd := getAbsPosition(prev.EndPoint(), f.lineStartPositions)
			currStart := getAbsPosition(ch.StartPoint(), f.lineStartPositions)
			between := f.source[prevEnd:currStart]

			if ch.Type() == "comment" && countLF(between) == 0 {
				f.writeByte(' ')
			} else {
				switch prev.Type() {
				case "class_def", "method_def":
					f.writeLF()
					f.writeLF()
				default:
					f.writeLF()
					if hasTwoNewlines(between) {
						f.writeLF()
					}
				}
			}
		}

		f.writeIndent(indent)
		f.formatNode(ch, indent)
	}
}

func (f *Formatter) formatOperator(node *sitter.Node) {
	f.writeByte(' ')
	f.writeContent(node)
	f.writeByte(' ')
}

func (f *Formatter) formatIf(node *sitter.Node, indent int) {

	condNode := node.ChildByFieldName("cond")
	if condNode != nil {
		// Write condition
		switch node.Type() {
		case "if":
			f.writeString("if")
		case "elsif":
			f.writeString("elsif")
		}
		f.writeByte(' ')

		if condNode.Type() == "expressions" {
			f.formatExpressions(condNode, indent, false)
		} else {
			f.formatNode(condNode, indent)
		}

		for ch := range eachChild(node) {
			if ch.Type() == "comment" {
				f.writeLF()
				f.writeIndent(indent + f.indentSize)
				f.formatNode(ch, indent)
			}
		}

		// Write then
		if thenNode := node.ChildByFieldName("then"); thenNode != nil {
			for ch := range eachChild(thenNode) {
				f.writeLF()
				f.writeIndent(indent + f.indentSize)
				f.formatNode(ch, indent+f.indentSize)
			}
		}

		// Write else
		if elseNode := node.ChildByFieldName("else"); elseNode != nil {

			// Write elsif
			if elsifCondNode := elseNode.ChildByFieldName("cond"); elsifCondNode != nil {
				f.writeLF()
				f.formatIf(elseNode, indent)
				return
			}

			// Write else body
			f.writeLF()
			f.writeIndent(indent)
			for ch := range eachChild(elseNode) {
				switch ch.Type() {
				case "else":
					f.writeContent(ch)
					continue
				default:
					f.writeLF()
					f.formatNode(ch, indent+f.indentSize)
				}
			}
		}

		// Write end
		f.writeLF()
		f.writeIndent(indent)
		f.writeString("end")
	}
}

func (f *Formatter) formatConditional(node *sitter.Node, indent int) {
	for ch := range eachChild(node) {
		switch ch.Type() {
		case "expressions":
			f.formatExpressions(ch, indent, false)
		case "?", ":":
			f.writeByte(' ')
			f.writeContent(ch)
			f.writeByte(' ')
		default:
			f.formatNode(ch, indent)
		}
	}
}

func (f *Formatter) formatConstant(node *sitter.Node) {
	f.writeContent(node)
}

func (f *Formatter) formatYield(node *sitter.Node) {
	for ch := range eachChild(node) {
		switch ch.Type() {
		case "yield":
			// if previous sibling does not end in '\n', prepend a ' '
			if sib := ch.PrevSibling(); sib != nil {
				endPos := getAbsPosition(sib.Range().EndPoint, f.lineStartPositions)
				endByte := f.source[endPos]
				if endByte != '\n' {
					f.writeByte(' ')
				}
			}
			f.writeContent(ch)
		case "argument_list":
			if ch.Child(0).Type() != "(" {
				f.writeByte(' ')
			}
			f.formatNode(ch, 0)
		default:
			// f.writeByte(' ')
			f.formatNode(ch, 0)
		}
	}
}

func (f *Formatter) formatModifierIf(node *sitter.Node) {
	thenNode := node.ChildByFieldName("then")
	f.formatNode(thenNode, 0)

	condNode := node.ChildByFieldName("cond")
	f.writeString(" if ")
	f.formatNode(condNode, 0)
}

func (f *Formatter) formatArray(node *sitter.Node) {
	for ch := range eachChild(node) {
		switch ch.Type() {
		case ",":
			f.writeContent(ch)
			f.writeByte(' ')
		default:
			f.formatNode(ch, 0)
		}
	}
}

func (f *Formatter) formatProcType(node *sitter.Node) {
	for ch := range eachChild(node) {
		switch ch.Type() {
		case "->":
			f.writeByte(' ')
			f.writeContent(ch)
			f.writeByte(' ')
		default:
			f.formatNode(ch, 0)
		}
	}
}

func (f *Formatter) formatWith(node *sitter.Node) {
	f.writeContent(node)
}

func (f *Formatter) formatSelf(node *sitter.Node) {
	if sib := node.PrevSibling(); sib != nil {
		if sib.Type() == "with" {
			f.writeByte(' ')
		}
	}
	f.writeContent(node)
}

// Recursive function to format the syntax tree
func (f *Formatter) formatNode(node *sitter.Node, indent int) {

	// for ch, idx := range eachChild(node) {
	// 	field := node.FieldNameForChild(int(idx))
	// 	fmt.Println(strings.Repeat(" ", indent) + node.Type())
	// 	if field != "" {
	// 		fmt.Println(strings.Repeat(" ", indent) + " field: " + field)
	// 	}
	// 	f.formatNode(ch, indent+f.indentSize)
	// }
	// return

	// for _, idx := range eachChild(node) {
	// 	field := node.FieldNameForChild(int(idx))
	// 	if field != "" {
	// 		fmt.Println("--- field:", field)
	// 		fieldNode := node.ChildByFieldName(field)
	// 		for ch := range eachChild(fieldNode) {
	// 			fmt.Println("------ type:", ch.Type())
	// 		}
	// 	}
	// }

	switch node.Type() {
	case "class_def":
		f.formatClass(node, indent)

	case "method_def":
		f.formatMethod(node, indent)

	case "expressions":
		f.formatExpressions(node, indent, true)

	case "require":
		f.formatRequire(node)

	case "assign":
		f.formatAssign(node, indent)

	case "call":
		f.formatCall(node, indent)

	case "param_list":
		f.formatParams(node)

	case "argument_list":
		f.formatArguments(node, indent)

	case "block":
		f.formatBlock(node, indent)

	case "string":
		f.formatString(node)

	case "array":
		f.formatArray(node)

	case "integer":
		f.formatLiteral(node)

	case "identifier":
		f.formatIdentifier(node)

	case "comment":
		f.formatComment(node)

	case "if", "then", "else":
		f.formatIf(node, indent)

	case "conditional":
		f.formatConditional(node, indent)

	case "modifier_if":
		f.formatModifierIf(node)

	case "yield":
		f.formatYield(node)

	case "operator":
		f.formatOperator(node)

	case "constant":
		f.formatConstant(node)

	case "symbol":
		f.formatSymbol(node)

	case "proc_type":
		f.formatProcType(node)

	case "with":
		f.formatWith(node)

	case "self":
		f.formatSelf(node)

	case "ERROR":
		f.err = errors.New(node.Content(f.source))
		return

	default:
		fmt.Println("--- caught:", node.Type())
		// Fallback to just printing the raw source content for unknown types
		f.writeContent(node)
	}
}

func (f *Formatter) writeIndent(indent int) {
	for range indent {
		f.strBuilder.WriteByte(' ')
	}
}

func (f *Formatter) writeContent(node *sitter.Node) {
	f.strBuilder.WriteString(f.getContent(node))
}

func (f *Formatter) writeString(str string, a ...any) {
	fmt.Fprintf(f.strBuilder, str, a...)
}

func (f *Formatter) writeByte(b byte) {
	f.strBuilder.WriteByte(b)
}

func (f *Formatter) writeLF() {
	f.writeByte('\n')
}

func (f *Formatter) getContent(node *sitter.Node) string {
	return node.Content(f.source)
}

func eachChild(node *sitter.Node) iter.Seq2[*sitter.Node, uint32] {
	return func(yield func(*sitter.Node, uint32) bool) {
		for idx := range node.ChildCount() {
			ch := node.Child(int(idx))
			if !yield(ch, idx) {
				return
			}
		}
	}
}

func eachChildByFieldName(node *sitter.Node, field string) iter.Seq2[*sitter.Node, uint32] {
	return func(yield func(*sitter.Node, uint32) bool) {
		for ch, idx := range eachChild(node) {
			chField := node.FieldNameForChild(int(idx))
			if chField == field {
				if !yield(ch, idx) {
					return
				}
			}
		}
	}
}

// Helper function to check for whitespace
func isWhitespace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r'
}

// getAbsPosition converts a Tree-sitter Point (line, column) to an absolute byte position
// in the source code. It requires a pre-computed array of line start positions.
func getAbsPosition(sp sitter.Point, lineStartPositions []int) int {
	// If we have an invalid point or empty lineStartPositions, return 0
	if len(lineStartPositions) == 0 {
		return 0
	}

	// Make sure we don't go out of bounds
	if int(sp.Row) >= len(lineStartPositions) {
		// Return the last position if the row is beyond our line count
		return lineStartPositions[len(lineStartPositions)-1]
	}

	// Get the starting position of the line
	lineStart := lineStartPositions[sp.Row]

	// Add the column offset to get the absolute position
	return lineStart + int(sp.Column)
}

// Helper function to build the line start positions array
// This should be called once when loading the source code
func buildLineStartPositions(source []byte) []int {
	positions := []int{0} // First line always starts at position 0

	for i := range source {
		// If we find a newline character, the next line starts at i+1
		if source[i] == '\n' {
			positions = append(positions, i+1)
		}
	}

	return positions
}

func hasTwoNewlines(b []byte) bool {
	count := 0
	for _, c := range b {
		if c == '\n' {
			count++
			if count >= 2 {
				return true
			}
		}
	}
	return false
}

func countLF(b []byte) int {
	count := 0
	for _, c := range b {
		if c == '\n' {
			count++
		}
	}
	return count
}
