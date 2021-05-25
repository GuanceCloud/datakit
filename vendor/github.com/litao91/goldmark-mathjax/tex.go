package mathjax

import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

type tex struct {
	texRenderer *TexRenderer
}

var Tex = &tex{
	texRenderer: NewDefaultTexRenderer(),
}

func (e *tex) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithBlockParsers(
		util.Prioritized(NewMathJaxBlockParser(), 701),
	))
	m.Parser().AddOptions(parser.WithInlineParsers(
		util.Prioritized(NewInlineMathParser(), 501),
	))
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(&MathTexBlockRenderer{e.texRenderer}, 501),
		util.Prioritized(&InlineTexMathRenderer{
			texRenderer: e.texRenderer,
		}, 502),
	))
}
