package repro_lang

import (
	"fmt"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/inputs"
)

type sourceFile struct {
	relativePath string
	code         string
	lines        []string
	node         *sitter.Node
	definitions  []definitionStatement
	references   []referenceStatement
	localScope   *scope
	localCounter int
}

func newSourceFile(relativePath, code string, node *sitter.Node) *sourceFile {
	return &sourceFile{
		relativePath: relativePath,
		code:         code,
		lines:        strings.Split(code, "\n"),
		node:         node,
		definitions:  nil,
		references:   nil,
		localScope:   newScope(),
	}
}

func (d *sourceFile) enterNewLocalSymbol(name identifier) string {
	symbol := fmt.Sprintf("local %v", name.value[len("local"):])
	d.localScope.names[name.value] = symbol
	return symbol
}

func (d *sourceFile) slicePosition(n *sitter.Node) string {
	return d.code[n.StartByte():n.EndByte()]
}
func (d *sourceFile) newIdentifier(n *sitter.Node) *identifier {
	if n == nil {
		return nil
	}
	if n.Type() != "identifier" {
		panic("expected identifier, obtained " + n.Type())
	}
	return &identifier{
		value:    d.slicePosition(n),
		position: inputs.NewRangePositionFromNode(n),
	}
}
