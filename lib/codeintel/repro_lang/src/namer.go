package repro_lang

import "fmt"

func (d *sourceFile) resolveReferences(context *globalContext) {
	for _, def := range d.definitions {
		for _, ident := range def.relationIdentifiers() {
			if ident == nil {
				continue
			}
			ident.resolveSymbol(d.localScope, context)
		}
	}
	for _, ref := range d.references {
		ref.name.resolveSymbol(d.localScope, context)
	}
}

func (d *sourceFile) resolveDefinitions(context *globalContext) {
	for _, def := range d.definitions {
		scope := context.globalScope
		if def.name.isLocalSymbol() {
			scope = d.localScope
		}
		symbol, ok := scope.names[def.name.value]
		if ok {
			symbol = "ERROR_DUPLICATE_DEFINITION"
		} else {
			symbol = fmt.Sprintf(
				"repro_lang %v %v %v/%v.",
				context.packageInformation.name,
				context.packageInformation.version,
				d.relativePath,
				def.name.value,
			)
		}
		def.name.symbol = symbol
	}

}
