package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode"

	"github.com/hexops/gotextdiff"
	"github.com/hexops/gotextdiff/myers"
	"github.com/hexops/gotextdiff/span"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/lsif/protocol/reader"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/lsif_typed"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type documentInput struct {
	RelativePath string
	Code         string
	lines        []string
}

func newDocumentInput(filename, code string) *documentInput {
	return &documentInput{
		RelativePath: filename,
		Code:         code,
		lines:        strings.Split(code, "\n"),
	}
}

func (d *documentInput) lineContent(position rangePosition) string {
	return d.lines[position.startLine]
}
func (d *documentInput) lineCaret(position rangePosition) string {
	carets := strings.Repeat("^", position.endCharacter-position.startCharacter)
	if position.startLine != position.endLine {
		carets = strings.Repeat("^", len(d.lines[position.startLine])-position.startCharacter)
	}
	return strings.Repeat(" ", position.startCharacter) + carets
}

func (d *documentInput) String() string {
	data, err := json.Marshal(&d)
	if err != nil {
		panic(err)
	}
	return string(data)
}

func (d *documentInput) SlicePosition(position rangePosition) string {
	result := strings.Builder{}
	for line := position.startLine; line < position.endLine; line++ {
		start := position.startCharacter
		if line > position.startLine {
			result.WriteString("\n")
			start = 0
		}
		end := position.endCharacter
		if line < position.endLine {
			end = len(d.lines[line])
		}
		result.WriteString(d.lines[line][start:end])
	}
	return result.String()
}

type exampleTokenKind string

const (
	identifierToken     exampleTokenKind = "identifier"
	definitionToken                      = "definition"
	referenceToken                       = "reference"
	implementationToken                  = "implements"
	referencesToken                      = "references"
	typeDefinitionToken                  = "type_definition"
	commentToken                         = "comment"
	newlineToken                         = "\n"
)

var exampleKeywords = []exampleTokenKind{
	definitionToken,
	referenceToken,
	implementationToken,
	referencesToken,
	typeDefinitionToken,
}

type rangePosition struct {
	startLine      int
	startCharacter int
	endLine        int
	endCharacter   int
}

type exampleToken struct {
	value    string
	kind     exampleTokenKind
	position rangePosition
}

type exampleRole string

const (
	definitionRole exampleRole = "definition"
	referenceRole  exampleRole = "reference"
)

type exampleDocument struct {
	globalScope  *exampleScope
	localScope   *exampleScope
	lsifDocument *lsif_typed.Document
	input        *documentInput
}

func newDocument(input *documentInput) *exampleDocument {
	return &exampleDocument{
		globalScope:  newScope(),
		localScope:   newScope(),
		lsifDocument: &lsif_typed.Document{RelativePath: input.RelativePath},
		input:        input,
	}
}

type exampleIdentifier struct {
	value    string
	position rangePosition
	symbol   string
}

func (i *exampleIdentifier) resolveSymbol(role exampleRole, document *exampleDocument) {
	isLocal := strings.HasPrefix(i.value, "local")
	scope := document.globalScope
	if isLocal {
		scope = document.localScope
	}
	symbol, ok := scope.namesToSymbols[i.value]
	switch role {
	case definitionRole:
		if ok {
			symbol = fmt.Sprintf("local ERROR_SYMBOL_ALREADY_DEFINED_%v", i.value)
		} else /* !ok */ {
			if isLocal {
				symbol = fmt.Sprintf("local %v", i.value[len("local"):])
			} else {
				symbol = fmt.Sprintf("%s/%s", document.lsifDocument.RelativePath, i.value)
			}
		}
		scope.namesToSymbols[i.value] = symbol
	case referenceRole:
		if !ok {
			symbol = fmt.Sprintf("local ERROR%v", i.value)
		}
	}
	i.symbol = symbol
}

type exampleStatement struct {
	Docstring              *exampleToken
	Role                   exampleRole
	Identifier             *exampleIdentifier
	ImplementationRelation *exampleIdentifier
	ReferenceRelation      *exampleIdentifier
	TypeDefinitionRelation *exampleIdentifier
}

func (e *exampleStatement) String() string {
	data, err := json.Marshal(&e)
	if err != nil {
		panic(err)
	}
	return string(data)
}

func (e *exampleStatement) RelationIdentifiers() []*exampleIdentifier {
	return []*exampleIdentifier{e.ImplementationRelation, e.ReferenceRelation, e.TypeDefinitionRelation}
}

func (e *exampleStatement) Relationships() []*lsif_typed.Relationship {
	relationships := map[string]*lsif_typed.Relationship{}
	for _, relation := range e.RelationIdentifiers() {
		if relation != nil {
			relationships[e.ImplementationRelation.symbol] = &lsif_typed.Relationship{Symbol: relation.symbol}
		}
	}
	if e.ImplementationRelation != nil {
		relationships[e.ImplementationRelation.symbol].IsImplementation = true
	}
	if e.ReferenceRelation != nil {
		relationships[e.ReferenceRelation.symbol].IsReference = true
	}
	if e.TypeDefinitionRelation != nil {
		relationships[e.TypeDefinitionRelation.symbol].IsTypeDefinition = true
	}
	var result []*lsif_typed.Relationship
	for _, value := range result {
		result = append(result, value)
	}
	return result
}

func indexExampleLanguage(projectName string, documents []*documentInput) (*lsif_typed.Index, error) {
	index := &lsif_typed.Index{
		Metadata: &lsif_typed.Metadata{
			Version: lsif_typed.ProtocolVersion_UnspecifiedProtocolVersion,
			ToolInfo: &lsif_typed.ToolInfo{
				Name:      "example-name",
				Version:   "example-version",
				Arguments: []string{"example-argument1", "example-argument2"},
			},
			ProjectRoot:          "file:/" + projectName,
			TextDocumentEncoding: lsif_typed.TextEncoding_UTF8,
		},
		Documents:       nil,
		ExternalSymbols: nil,
	}
	for _, input := range documents {
		statements, err := parseExampleLanguage(input)
		if err != nil {
			return nil, err
		}
		document := newDocument(input)
		scope := &exampleScope{namesToSymbols: map[string]string{}}
		for _, statement := range statements {

			// Resolve symbols
			statement.Identifier.resolveSymbol(statement.Role, document)
			for _, relation := range statement.RelationIdentifiers() {
				if relation != nil {
					relation.resolveSymbol("reference", document)
				}
			}

			role := int32(0)
			if statement.Role == definitionRole {
				role |= int32(lsif_typed.SymbolRole_Definition)
				info := &lsif_typed.SymbolInformation{
					Symbol:        statement.Identifier.symbol,
					Relationships: nil,
				}
				if statement.Docstring != nil {
					info.Documentation = append(info.Documentation, input.SlicePosition(statement.Docstring.position))
				}
				info.Relationships = statement.Relationships()
				document.lsifDocument.Symbols = append(document.lsifDocument.Symbols, info)
			}
			document.lsifDocument.Occurrences = append(document.lsifDocument.Occurrences, &lsif_typed.Occurrence{
				Range:       statement.Identifier.position.lsifRange(),
				Symbol:      statement.Identifier.symbol,
				SymbolRoles: role,
				SyntaxKind:  lsif_typed.SyntaxKind_Identifier,
			})
		}
		for identifier, symbol := range scope {
		}
		index.Documents = append(index.Documents, document)
	}

	return index, nil
}

type exampleScope struct {
	namesToSymbols map[string]string
}

func newScope() *exampleScope {
	return &exampleScope{namesToSymbols: map[string]string{}}
}

func (pos *rangePosition) lsifRange() []int32 {
	if pos.startLine == pos.endLine {
		return []int32{int32(pos.startLine), int32(pos.startCharacter), int32(pos.endCharacter)}
	}
	return []int32{int32(pos.startLine), int32(pos.startCharacter), int32(pos.endLine), int32(pos.endCharacter)}
}

type exampleParser struct {
	input                *documentInput
	tokens               []exampleToken
	tokenIndex           int
	previousCommentToken *exampleToken
}

func (e exampleParser) currentToken() *exampleToken {
	if e.tokenIndex < len(e.tokens) {
		return &e.tokens[e.tokenIndex]
	}
	return nil
}

func (e exampleParser) hasMoreTokens() bool {
	return e.tokenIndex+1 < len(e.tokens)
}

func (e exampleParser) newError(position rangePosition, formatString string, args ...interface{}) error {
	return errors.Newf(
		"%v:%v:%v %v\n%v\n%v",
		e.input.RelativePath,
		position.startLine,
		position.startCharacter,
		fmt.Sprintf(formatString, args),
		e.input.lineContent(position),
		e.input.lineCaret(position),
	)
}

func (e exampleParser) nextToken() *exampleToken {
	for e.hasMoreTokens() {
		e.tokenIndex++
		token := &e.tokens[e.tokenIndex]
		if token.kind == commentToken {
			e.previousCommentToken = token
			continue
		} else {
			e.previousCommentToken = nil
		}
		return token
	}
	return nil
}

func (e exampleParser) nextTokenAsIdentifier() (*exampleToken, error) {
	token := e.nextToken()
	if token == nil {
		return nil, nil
	}
	if token.kind != identifierToken {
		return nil, e.newError(token.position, "expected identifier token, obtained %v", token.kind)
	}
	return token, nil
}

func (e exampleParser) parseStatement() (*exampleStatement, error) {
	statementRole := e.nextToken()
	if statementRole == nil {
		return nil, nil
	}
	token, err := e.nextTokenAsIdentifier()
	if err != nil {
		return nil, err
	}
	identifier := &exampleIdentifier{
		value:    token.value,
		position: token.position,
		symbol:   "",
	}
	e.nextToken()
	switch statementRole.kind {
	case definitionToken:
		return e.parseDefinitionStatement(identifier)
	case referencesToken:
		return &exampleStatement{
			Role:       referenceRole,
			Identifier: identifier,
		}, nil
	default:
		return nil, errors.Errorf("expected token '%v' or '%v', obtained %v", definitionRole, referenceRole, statementRole.kind)
	}
}
func (e exampleParser) parseDefinitionStatement(identifier *exampleIdentifier) (*exampleStatement, error) {
	definition := &exampleStatement{
		Docstring:  e.previousCommentToken,
		Role:       definitionRole,
		Identifier: identifier,
	}
	for e.currentToken() != nil {
		switch e.currentToken().kind {
		case implementationToken, referencesToken, typeDefinitionToken:
			kind := e.currentToken().kind
			token, err := e.nextTokenAsIdentifier()
			if err != nil {
				return nil, err
			}
			relationIdentifier := &exampleIdentifier{value: token.value}
			switch kind {
			case implementationToken:
				definition.ImplementationRelation = relationIdentifier
			case referencesToken:
				definition.ReferenceRelation = relationIdentifier
			case typeDefinitionToken:
				definition.TypeDefinitionRelation = relationIdentifier
			}
			e.nextToken()
		default:
			return definition, nil
		}
	}
	return definition, nil
}

func parseExampleLanguage(input *documentInput) ([]*exampleStatement, error) {
	tokens, err := tokenizeExampleLanguage(input)
	if err != nil {
		return nil, err
	}
	parser := &exampleParser{
		tokens:     tokens,
		tokenIndex: 0,
	}
	var result []*exampleStatement
	for parser.hasMoreTokens() {
		statement, err := parser.parseStatement()
		if err != nil {
			return nil, err
		}
		if statement == nil {
			break
		}
		result = append(result, statement)
	}
	for i := 0; i < len(tokens); i++ {

	}
	return result, nil
}

func tokenizeExampleLanguage(input *documentInput) ([]exampleToken, error) {
	var result []exampleToken
	for line, lineString := range strings.Split(input.Code, "\n") {
		runes := []rune(lineString)
		for character := 0; character < len(runes); character++ {
			ch := runes[character]
			if unicode.IsSpace(ch) {
				continue
			} else if ch == '#' {
				result = append(result, exampleToken{
					value: string(runes[character:]),
					kind:  commentToken,
					position: rangePosition{
						startLine:      line,
						startCharacter: character,
						endLine:        line,
						endCharacter:   len(runes),
					},
				})
				break
			} else if isIdentifierRune(ch) {
				startCharacter := character
				for character < len(runes) && isIdentifierRune(runes[character]) {
					character++
				}
				value := string(runes[startCharacter:character])
				position := rangePosition{
					startLine:      line,
					startCharacter: startCharacter,
					endLine:        line,
					endCharacter:   character,
				}
				result = append(result, exampleToken{value: value, kind: tokenKind(value), position: position})
				character--
			} else {
				return nil, errors.Newf("unexpected token %v", string(ch))
			}
		}
		result = append(result, exampleToken{value: "\n", kind: newlineToken, position: rangePosition{
			startLine:      line,
			startCharacter: len(runes),
			endLine:        line,
			endCharacter:   len(runes) + 1,
		}})
	}
	return result, nil
}

func tokenKind(value string) exampleTokenKind {
	for _, keyword := range exampleKeywords {
		if value == string(keyword) {
			return keyword
		}
	}
	return identifierToken
}

func isIdentifierRune(r rune) bool {
	return !unicode.IsSpace(r)
}

func isUpdateSnapshots() bool {
	for _, arg := range os.Args {
		if arg == "-update-snapshots" {
			return true
		}
	}
	return false
}

func TestLsifTyped(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	inputDirectory := filepath.Join(cwd, "snapshots-input")
	outputDirectory := filepath.Join(cwd, "snapshots-output")
	testCases, err := os.ReadDir(inputDirectory)
	if err != nil {
		t.Fatal(err)
	}

	for _, testCase := range testCases {
		if !testCase.IsDir() {
			t.Fatalf("not a directory: %v", testCase.Name())
		}
		t.Run(testCase.Name(), func(t *testing.T) {
			dir := filepath.Join(inputDirectory, testCase.Name())
			outputFile := filepath.Join(outputDirectory, testCase.Name(), "dump.lsif")
			var inputs []*documentInput
			entries, err := os.ReadDir(dir)
			if err != nil {
				t.Fatal(err)
			}
			for _, entry := range entries {
				absolutePath := filepath.Join(dir, entry.Name())
				relativePath := filepath.Join(testCase.Name(), entry.Name())
				data, err := os.ReadFile(absolutePath)
				if err != nil {
					t.Fatal(err)
				}
				inputs = append(inputs, newDocumentInput(relativePath, string(data)))
			}
			index, err := indexExampleLanguage(testCase.Name(), inputs)
			if err != nil {
				t.Fatal(err)
			}
			lsif, err := reader.ConvertTypedIndexToGraphIndex(index)
			if err != nil {
				t.Fatal(err)
			}
			var obtained bytes.Buffer
			err = reader.WriteNDJSON(reader.ElementsToEmptyInterfaces(lsif), &obtained)
			if err != nil {
				t.Fatal(err)
			}
			if isUpdateSnapshots() {
				err = os.MkdirAll(filepath.Dir(outputFile), 0644)
				if err != nil {
					t.Fatal(err)
				}
				err = os.WriteFile(outputFile, obtained.Bytes(), 0644)
				if err != nil {
					t.Fatal(err)
				}
			} else {
				expected, err := os.ReadFile(outputFile)
				if err != nil {
					expected = []byte{}
				}
				edits := myers.ComputeEdits(span.URIFromPath(outputFile), string(expected), obtained.String())
				if len(edits) > 0 {
					diff := fmt.Sprint(gotextdiff.ToUnified(
						outputFile+" (obtained)",
						outputFile+" (expected)",
						string(expected),
						edits,
					))
					t.Fatalf("\n" + diff)
				}
			}
		})
	}
}
