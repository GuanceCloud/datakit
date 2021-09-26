package http

import (
	"html/template"
	"net/http"
	"sort"

	"github.com/gin-gonic/gin"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/man"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	manualTOCTemplate = `
<style>
div {
  border: 1px solid gray;
  /* padding: 8px; */
}

h1 {
  text-align: center;
  text-transform: uppercase;
  color: #4CAF50;
}

p {
  /* text-indent: 50px; */
  text-align: justify;
  /* letter-spacing: 3px; */
}

a {
  text-decoration: none;
  /* color: #008CBA; */
}

ul.a {
  list-style-type: square;
}
</style>

<h1>{{.PageTitle}}</h1>
采集器文档列表
<ul class="a">
	{{ range $name := .InputNames}}
	<li>
	<p><a href="/man/{{$name}}">
			{{$name}} </a> </p> </li>
	{{end}}
</ul>

其它文档集合

<ul class="a">
	{{ range $name := .OtherDocs}}
	<li>
	<p><a href="/man/{{$name}}">
			{{$name}} </a> </p> </li>
	{{end}}
</ul>
`

	headerScript = []byte(`<link rel="stylesheet"
      href="//cdnjs.cloudflare.com/ajax/libs/highlight.js/10.7.2/styles/default.min.css">
<script src="//cdnjs.cloudflare.com/ajax/libs/highlight.js/10.7.2/highlight.min.js"></script>
<script>
document.addEventListener('DOMContentLoaded', (event) => {
    hljs.highlightAll();
});
</script>`)
)

type manualTOC struct {
	PageTitle  string
	InputNames []string
	OtherDocs  []string
}

// request manual table of conotents
func apiManualTOC(c *gin.Context) {
	toc := &manualTOC{
		PageTitle: "DataKit文档列表",
	}

	for k, v := range inputs.Inputs {
		if _, ok := v().(inputs.InputV2); ok {
			// test if doc available
			if _, err := man.BuildMarkdownManual(k, &man.Option{WithCSS: true}); err != nil {
				l.Warn(err)
			} else {
				toc.InputNames = append(toc.InputNames, k)
			}
		}
	}
	sort.Strings(toc.InputNames)

	for k := range man.OtherDocs {
		// test if doc available
		if _, err := man.BuildMarkdownManual(k, &man.Option{WithCSS: true}); err != nil {
			l.Warn(err)
		} else {
			toc.OtherDocs = append(toc.OtherDocs, k)
		}
	}
	sort.Strings(toc.OtherDocs)

	t := template.New("man-toc")

	tmpl, err := t.Parse(manualTOCTemplate)
	if err != nil {
		l.Error(err)
		c.Data(http.StatusInternalServerError, "", []byte(err.Error()))
		return
	}

	if err := tmpl.Execute(c.Writer, toc); err != nil {
		l.Error(err)
		c.Data(http.StatusInternalServerError, "", []byte(err.Error()))
		return
	}
}

func apiManual(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.Redirect(200, "/man")
		return
	}

	mdtxt, err := man.BuildMarkdownManual(name, &man.Option{WithCSS: true})
	if err != nil {
		c.Redirect(200, "/man")
		return
	}

	// render markdown as HTML
	mdext := parser.CommonExtensions
	psr := parser.NewWithExtensions(mdext)

	htmlFlags := html.CommonFlags | html.HrefTargetBlank | html.TOC | html.CompletePage
	opts := html.RendererOptions{Flags: htmlFlags, Head: headerScript}
	renderer := html.NewRenderer(opts)

	out := markdown.ToHTML(mdtxt, psr, renderer)
	c.Data(http.StatusOK, "text/html; charset=UTF-8", out)
}
