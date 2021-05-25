package mathjax

import (
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

type mathJaxBlockParser struct {
}

var defaultMathJaxBlockParser = &mathJaxBlockParser{}

type mathBlockData struct {
	indent int
}

var mathBlockInfoKey = parser.NewContextKey()

func NewMathJaxBlockParser() parser.BlockParser {
	return defaultMathJaxBlockParser
}

func (b *mathJaxBlockParser) Open(parent ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
	line, _ := reader.PeekLine()
	pos := pc.BlockOffset()
	if pos == -1 {
		return nil, parser.NoChildren
	}
	if line[pos] != '$' {
		return nil, parser.NoChildren
	}
	i := pos
	for ; i < len(line) && line[i] == '$'; i++ {
	}
	if i-pos < 2 {
		return nil, parser.NoChildren
	}
	pc.Set(mathBlockInfoKey, &mathBlockData{indent: pos})
	node := NewMathBlock()
	return node, parser.NoChildren
}

func (b *mathJaxBlockParser) Continue(node ast.Node, reader text.Reader, pc parser.Context) parser.State {
	line, segment := reader.PeekLine()
	data := pc.Get(mathBlockInfoKey).(*mathBlockData)
	w, pos := util.IndentWidth(line, 0)
	if w < 4 {
		i := pos
		for ; i < len(line) && line[i] == '$'; i++ {
		}
		length := i - pos
		if length >= 2 && util.IsBlank(line[i:]) {
			reader.Advance(segment.Stop - segment.Start - segment.Padding)
			return parser.Close
		}
	}

	pos, padding := util.DedentPosition(line, 0, data.indent)
	seg := text.NewSegmentPadding(segment.Start+pos, segment.Stop, padding)
	node.Lines().Append(seg)
	reader.AdvanceAndSetPadding(segment.Stop-segment.Start-pos-1, padding)
	return parser.Continue | parser.NoChildren
}

func (b *mathJaxBlockParser) Close(node ast.Node, reader text.Reader, pc parser.Context) {
	pc.Set(mathBlockInfoKey, nil)
}

func (b *mathJaxBlockParser) CanInterruptParagraph() bool {
	return true
}

func (b *mathJaxBlockParser) CanAcceptIndentedLine() bool {
	return false
}

func (b *mathJaxBlockParser) Trigger() []byte {
	return nil
}
