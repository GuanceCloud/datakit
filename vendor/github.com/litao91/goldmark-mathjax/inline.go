package mathjax

import (
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

type inlineMathParser struct {
}

var defaultInlineMathParser = &inlineMathParser{}

func NewInlineMathParser() parser.InlineParser {
	return defaultInlineMathParser
}

func (s *inlineMathParser) Trigger() []byte {
	return []byte{'$'}
}

func (s *inlineMathParser) Parse(parent ast.Node, block text.Reader, pc parser.Context) ast.Node {
	line, startSegment := block.PeekLine()
	opener := 0
	for ; opener < len(line) && line[opener] == '$'; opener++ {
	}
	block.Advance(opener)
	l, pos := block.Position()
	node := NewInlineMath()
	for {
		line, segment := block.PeekLine()
		if line == nil {
			block.SetPosition(l, pos)
			return ast.NewTextSegment(startSegment.WithStop(startSegment.Start + opener))
		}
		for i := 0; i < len(line); i++ {
			c := line[i]
			if c == '$' {
				oldi := i
				for ; i < len(line) && line[i] == '$'; i++ {
				}
				closure := i - oldi
				if closure == opener && (i+1 >= len(line) || line[i+1] != '$') {
					segment := segment.WithStop(segment.Start + i - closure)
					if !segment.IsEmpty() {
						node.AppendChild(node, ast.NewRawTextSegment(segment))
					}
					block.Advance(i)
					goto end
				}
			}
		}
		if !util.IsBlank(line) {
			node.AppendChild(node, ast.NewRawTextSegment(segment))
		}
		block.AdvanceLine()
	}
end:

	if !node.IsBlank(block.Source()) {
		// trim first halfspace and last halfspace
		segment := node.FirstChild().(*ast.Text).Segment
		shouldTrimmed := true
		if !(!segment.IsEmpty() && block.Source()[segment.Start] == ' ') {
			shouldTrimmed = false
		}
		segment = node.LastChild().(*ast.Text).Segment
		if !(!segment.IsEmpty() && block.Source()[segment.Stop-1] == ' ') {
			shouldTrimmed = false
		}
		if shouldTrimmed {
			t := node.FirstChild().(*ast.Text)
			segment := t.Segment
			t.Segment = segment.WithStart(segment.Start + 1)
			t = node.LastChild().(*ast.Text)
			segment = node.LastChild().(*ast.Text).Segment
			t.Segment = segment.WithStop(segment.Stop - 1)
		}

	}
	return node
}

func NewInlineMathRenderer(start, end string) renderer.NodeRenderer {
	return &InlineMathRenderer{start, end}
}
