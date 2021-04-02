package mathjax

import "github.com/yuin/goldmark/ast"

type MathBlock struct {
	ast.BaseBlock
}

var KindMathBlock = ast.NewNodeKind("MathBLock")

func NewMathBlock() *MathBlock {
	return &MathBlock{}
}

func (n *MathBlock) Dump(source []byte, level int) {
	m:= map[string]string{}
	ast.DumpHelper(n, source, level, m, nil)
}

func (n *MathBlock) Kind() ast.NodeKind {
	return KindMathBlock
}

func (n *MathBlock) IsRaw() bool {
	return true
}
