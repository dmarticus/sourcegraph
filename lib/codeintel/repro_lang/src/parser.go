package repro_lang

func (d *sourceFile) parseStatements() {
	for i := uint32(0); i < d.node.ChildCount(); i++ {
		child := d.node.Child(int(i))
		name := child.ChildByFieldName("name")
		if name == nil {
			continue
		}
		switch child.Type() {
		case "definition_statement":
			docstring := ""
			docstringNode := child.ChildByFieldName("docstring")
			if docstringNode != nil {
				docstring = d.slicePosition(docstringNode)[len("# doctring:"):]
			}
			d.definitions = append(d.definitions, definitionStatement{
				docstring:           docstring,
				name:                d.newIdentifier(child.ChildByFieldName("name")),
				implementsRelation:  d.newIdentifier(child.ChildByFieldName("implementation_relation")),
				referencesRelation:  d.newIdentifier(child.ChildByFieldName("references_relation")),
				typeDefinesRelation: d.newIdentifier(child.ChildByFieldName("type_definition_relation")),
			})
		case "reference_statement":
			d.references = append(d.references, referenceStatement{
				name: d.newIdentifier(child.ChildByFieldName("name")),
			})
		}
	}
}
