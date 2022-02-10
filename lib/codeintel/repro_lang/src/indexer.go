package repro_lang

import (
	"context"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/inputs"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/lsif_typed"
)

func Index(projectRoot, packageName string, documents []*inputs.Document) (*lsif_typed.Index, error) {
	index := &lsif_typed.Index{
		Metadata: &lsif_typed.Metadata{
			Version: 0,
			ToolInfo: &lsif_typed.ToolInfo{
				Name:      "repro_lang",
				Version:   "1.0.0",
				Arguments: []string{"arg1", "arg2"},
			},
			ProjectRoot:          projectRoot,
			TextDocumentEncoding: lsif_typed.TextEncoding_UTF8,
		},
		Documents:       nil,
		ExternalSymbols: nil,
	}

	ctx := &globalContext{
		globalScope: newScope(),
		packageInformation: packageInformation{
			name:    packageName,
			version: "1.0.0",
		},
	}

	// Phase 1: parse sources
	var sourceFiles []*sourceFile
	for _, document := range documents {
		tree, err := sitter.ParseCtx(context.Background(), []byte(document.Code), GetLanguage())
		if err != nil {
			return nil, err
		}
		parsedDocument := newSourceFile(document.RelativePath, document.Code, tree)
		parsedDocument.parseStatements()
		sourceFiles = append(sourceFiles, parsedDocument)
	}

	// Phase 2: resolve names for definitions
	for _, file := range sourceFiles {
		file.resolveDefinitions(ctx)
	}

	// Phase 3: resolve names for references
	for _, file := range sourceFiles {
		file.resolveReferences(ctx)
	}

	// Phase 4: emit LSIF Typed
	for _, file := range sourceFiles {
		lsifDocument := &lsif_typed.Document{
			RelativePath: file.relativePath,
			Occurrences:  file.occurrences(),
			Symbols:      file.symbols(),
		}
		index.Documents = append(index.Documents, lsifDocument)
	}

	return index, nil
}

type globalContext struct {
	globalScope *scope
	packageInformation
}

type packageInformation struct {
	name    string
	version string
}

type scope struct {
	names map[string]string
}

func newScope() *scope {
	return &scope{names: map[string]string{}}
}
