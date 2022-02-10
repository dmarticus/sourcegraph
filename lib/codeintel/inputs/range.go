package inputs

import sitter "github.com/smacker/go-tree-sitter"

type RangePosition struct {
	Start *Position
	End   *Position
}

func (r RangePosition) LsifRange() []int32 {
	if r.Start.Line == r.End.Line {
		return []int32{int32(r.Start.Line), int32(r.Start.Character), int32(r.End.Character)}
	}
	return []int32{int32(r.Start.Line), int32(r.Start.Character), int32(r.End.Line), int32(r.End.Character)}
}

func NewRangePositionFromNode(node *sitter.Node) *RangePosition {
	return &RangePosition{
		Start: &Position{
			Line:      int(node.StartPoint().Row),
			Character: int(node.StartPoint().Column),
		},
		End: &Position{
			Line:      int(node.EndPoint().Row),
			Character: int(node.EndPoint().Column),
		},
	}
}
