package inputs

import (
	"encoding/json"
	"strings"
)

type Document struct {
	RelativePath string
	Code         string
	Lines        []string
}

func NewDocument(filename, code string) *Document {
	return &Document{
		RelativePath: filename,
		Code:         code,
		Lines:        strings.Split(code, "\n"),
	}
}

func (d *Document) lineContent(position RangePosition) string {
	return d.Lines[position.Start.Line]
}
func (d *Document) lineCaret(position RangePosition) string {
	carets := strings.Repeat("^", position.End.Character-position.Start.Character)
	if position.Start.Line != position.End.Line {
		carets = strings.Repeat("^", len(d.Lines[position.Start.Line])-position.Start.Character)
	}
	return strings.Repeat(" ", position.Start.Character) + carets
}

func (d *Document) String() string {
	data, err := json.Marshal(&d)
	if err != nil {
		panic(err)
	}
	return string(data)
}

func (d *Document) SlicePosition(position RangePosition) string {
	result := strings.Builder{}
	for line := position.Start.Line; line < position.End.Line; line++ {
		start := position.Start.Character
		if line > position.Start.Line {
			result.WriteString("\n")
			start = 0
		}
		end := position.End.Character
		if line < position.End.Line {
			end = len(d.Lines[line])
		}
		result.WriteString(d.Lines[line][start:end])
	}
	return result.String()
}
