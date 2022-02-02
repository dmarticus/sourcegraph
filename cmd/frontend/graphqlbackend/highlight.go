package graphqlbackend

import (
	"context"
	"html/template"

	"github.com/gogo/protobuf/jsonpb"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/highlight"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
)

type highlightedRangeResolver struct {
	inner result.HighlightedRange
}

func (h highlightedRangeResolver) Line() int32      { return h.inner.Line }
func (h highlightedRangeResolver) Character() int32 { return h.inner.Character }
func (h highlightedRangeResolver) Length() int32    { return h.inner.Length }

type highlightedStringResolver struct {
	inner result.HighlightedString
}

func (s *highlightedStringResolver) Value() string { return s.inner.Value }
func (s *highlightedStringResolver) Highlights() []highlightedRangeResolver {
	res := make([]highlightedRangeResolver, len(s.inner.Highlights))
	for i, hl := range s.inner.Highlights {
		res[i] = highlightedRangeResolver{hl}
	}
	return res
}

type HighlightArgs struct {
	DisableTimeout     bool
	IsLightTheme       *bool
	HighlightLongLines bool
	TreeSitterEnabled  bool
}

type highlightedFileResolver struct {
	aborted bool
	html    template.HTML
	lsif    string
}

func (h *highlightedFileResolver) Aborted() bool { return h.aborted }
func (h *highlightedFileResolver) HTML() string  { return string(h.html) }
func (h *highlightedFileResolver) LSIF() string  { return string(h.lsif) }
func (h *highlightedFileResolver) LineRanges(args *struct{ Ranges []highlight.LineRange }) ([][]string, error) {
	return highlight.SplitLineRanges(h.html, args.Ranges)
}

func highlightContent(ctx context.Context, args *HighlightArgs, content, path string, metadata highlight.Metadata) (*highlightedFileResolver, error) {
	var (
		result          = &highlightedFileResolver{}
		err             error
		simulateTimeout = metadata.RepoName == "github.com/sourcegraph/AlwaysHighlightTimeoutTest"
	)

	html, document, aborted, err := highlight.Code(ctx, highlight.Params{
		Content:            []byte(content),
		Filepath:           path,
		DisableTimeout:     args.DisableTimeout,
		HighlightLongLines: args.HighlightLongLines,
		SimulateTimeout:    simulateTimeout,
		Metadata:           metadata,
		TreeSitterEnabled:  args.TreeSitterEnabled,
	})
	result.html = html
	result.aborted = aborted

	if document != nil {
		marshaller := &jsonpb.Marshaler{
			EnumsAsInts:  true,
			EmitDefaults: false,
		}

		result.lsif, err = marshaller.MarshalToString(document)
	}

	if err != nil {
		return nil, err
	}
	return result, nil
}
