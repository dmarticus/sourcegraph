package repro_lang

import (
	"strings"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/inputs"
)

type definitionStatement struct {
	docstring           string
	name                *identifier
	implementsRelation  *identifier
	referencesRelation  *identifier
	typeDefinesRelation *identifier
}

func (s *definitionStatement) relationIdentifiers() []*identifier {
	return []*identifier{s.implementsRelation, s.referencesRelation, s.typeDefinesRelation}
}

type referenceStatement struct {
	name *identifier
}

type identifier struct {
	value    string
	symbol   string
	position *inputs.RangePosition
}

func (i *identifier) resolveSymbol(localScope *scope, context *globalContext) {
	scope := context.globalScope
	if i.isLocalSymbol() {
		scope = localScope
	}
	symbol, ok := scope.names[i.value]
	if !ok {
		symbol = "ERROR_UNRESOLVED_SYMBOL"
	}
	i.symbol = symbol
}

func (i *identifier) isLocalSymbol() bool {
	return strings.HasPrefix(i.value, "local")
}
