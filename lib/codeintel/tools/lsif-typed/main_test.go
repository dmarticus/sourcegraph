package main

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/lsif_typed"
)

type document struct {
	relativeFilename string
	code             string
}

func indexExampleLanguage(documents []document) *lsif_typed.Index {
	index := &lsif_typed.Index{
		Metadata: &lsif_typed.Metadata{
			Version: lsif_typed.ProtocolVersion_UnspecifiedProtocolVersion,
			ToolInfo: &lsif_typed.ToolInfo{
				Name:      "example-name",
				Version:   "example-version",
				Arguments: []string{"example-argument1", "example-argument2"},
			},
			ProjectRoot:          "file://example-root",
			TextDocumentEncoding: lsif_typed.TextEncoding_UTF8,
		},
		Documents:       nil,
		ExternalSymbols: nil,
	}
	for _, document := range documents {
		index.Documents = append(index.Documents, &lsif_typed.Document{
			RelativePath: document.relativeFilename,
			Occurrences:  nil,
			Symbols:      nil,
		})
	}

	return index
}

func TestLsifTyped(t *testing.T) {
	symbol1 := "scheme package-manager package-name package-version path/to/symbol1#"
	symbol2 := "scheme package-manager package-name package-version path/to/symbol2#"
	symbol3 := "scheme package-manager package-name package-version path/to/symbol3#"
	index := &lsif_typed.Index{
		Documents: []*lsif_typed.Document{
			{
				RelativePath: "document1",
				Occurrences: []*lsif_typed.Occurrence{
					{
						Range:                 nil,
						Symbol:                "",
						SymbolRoles:           0,
						OverrideDocumentation: nil,
						SyntaxKind:            0,
						Diagnostics:           nil,
					},
				},
				Symbols: nil,
			},
		},
		ExternalSymbols: []*lsif_typed.SymbolInformation{
			{
				Symbol:        symbol3,
				Documentation: []string{"example documentation"},
				Relationships: []*lsif_typed.Relationship{
					{
						Symbol:           symbol2,
						IsReference:      true,
						IsImplementation: true,
						IsTypeDefinition: true,
					},
				},
			},
		},
	}
	elements

}
