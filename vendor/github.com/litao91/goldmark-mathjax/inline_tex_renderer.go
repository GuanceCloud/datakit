package mathjax

import (
	"bytes"
	"encoding/base64"
	"fmt"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

type InlineTexMathRenderer struct {
	texRenderer *TexRenderer
}

func (r *InlineTexMathRenderer) renderInlineMath(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		var buf bytes.Buffer
		for c := n.FirstChild(); c != nil; c = c.NextSibling() {
			segment := c.(*ast.Text).Segment
			value := segment.Value(source)
			if bytes.HasSuffix(value, []byte("\n")) {
				buf.Write(value[:len(value)-1])
				if c != n.LastChild() {
					buf.Write([]byte(" "))
				}
			} else {
				buf.Write(value)
			}
		}
		fmt.Println(buf.String())
		svg := r.texRenderer.RunInline(buf.String())
		_, _ = w.WriteString(`<span class="latex-svg inline">`)
		if svg != nil {
			_, _ = w.WriteString(`<img alt="" src="data:image/svg+xml;base64, ` + base64.StdEncoding.EncodeToString(svg) + `">`)
		}
		return ast.WalkSkipChildren, nil
	}
	_, _ = w.WriteString(`</span>`)
	return ast.WalkContinue, nil
}

func (r *InlineTexMathRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindInlineMath, r.renderInlineMath)
}
