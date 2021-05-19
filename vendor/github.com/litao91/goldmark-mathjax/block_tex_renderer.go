package mathjax

import (
	"encoding/base64"
	"bytes"

	gast "github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

type MathTexBlockRenderer struct {
	renderer *TexRenderer
}

func (r *MathTexBlockRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindMathBlock, r.renderMathBlock)
}

func (r *MathTexBlockRenderer) writeLines(w *bytes.Buffer, source []byte, n gast.Node) {
	l := n.Lines().Len()
	for i := 0; i < l; i++ {
		line := n.Lines().At(i)
		w.Write(line.Value(source))
	}
}

func (r *MathTexBlockRenderer) renderMathBlock(w util.BufWriter, source []byte, node gast.Node, entering bool) (gast.WalkStatus, error) {
	n := node.(*MathBlock)
	if entering {
		_, _ = w.WriteString(`<div class="latex-svg display">`)
		var buf bytes.Buffer
		r.writeLines(&buf, source, n)
		str := buf.String()
		svg := r.renderer.Run(str)
		w.WriteString(`<embed alt="" src="data:image/svg+xml;base64, ` + base64.StdEncoding.EncodeToString(svg) + `"></embed>`)
	} else {
		_, _ = w.WriteString(`</div>` + "\n")
	}
	return gast.WalkContinue, nil
}
