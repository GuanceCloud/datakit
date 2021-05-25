package mathjax

import (
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/util"
)

type InlineMath struct {
	ast.BaseInline
}

func (n *InlineMath) Inline() {}

func (n *InlineMath) IsBlank(source []byte) bool {
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		text := c.(*ast.Text).Segment
		if !util.IsBlank(text.Value(source)) {
			return false
		}
	}
	return true
}

func (n *InlineMath) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, nil, nil)
}

var KindInlineMath = ast.NewNodeKind("InlineMath")

func (n *InlineMath) Kind() ast.NodeKind {
	return KindInlineMath
}

func NewInlineMath() *InlineMath {
	return &InlineMath{
		BaseInline: ast.BaseInline{},
	}
}

