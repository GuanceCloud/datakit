package mathjax

import (
	gast "github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

type MathBlockRenderer struct {
	startDelim string
	endDelim   string
}

func NewMathBlockRenderer(start, end string) renderer.NodeRenderer {
	return &MathBlockRenderer{start, end}
}

func (r *MathBlockRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindMathBlock, r.renderMathBlock)
}

func (r *MathBlockRenderer) writeLines(w util.BufWriter, source []byte, n gast.Node) {
	l := n.Lines().Len()
	for i := 0; i < l; i++ {
		line := n.Lines().At(i)
		w.Write(line.Value(source))
	}
}

func (r *MathBlockRenderer) renderMathBlock(w util.BufWriter, source []byte, node gast.Node, entering bool) (gast.WalkStatus, error) {
	n := node.(*MathBlock)
	if entering {
		_, _ = w.WriteString(`<p><span class="math display">` + r.startDelim)
		r.writeLines(w, source, n)
	} else {
		_, _ = w.WriteString(r.endDelim + `</span></p>` + "\n")
	}
	return gast.WalkContinue, nil
}
