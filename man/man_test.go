package man

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/yuin/goldmark-highlighting"

	mathjax "github.com/litao91/goldmark-mathjax"

	//fhtml "github.com/alecthomas/chroma/formatters/html"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	gparser "github.com/yuin/goldmark/parser"
	ghtml "github.com/yuin/goldmark/renderer/html"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
)

func TestMarkdownToHTML(t *testing.T) {

	source, err := ioutil.ReadFile("./test.md")
	if err != nil {
		t.Fatal(err)
	}

	opt := &testutil.HTTPServerOptions{
		Bind: ":12345",
		Exit: make(chan interface{}),
		Routes: map[string]func(*gin.Context){
			"/goldmark": func(c *gin.Context) {

				var buf bytes.Buffer

				md := goldmark.New(
					goldmark.WithExtensions(highlighting.Highlighting),
					goldmark.WithExtensions(extension.GFM),
					goldmark.WithExtensions(mathjax.MathJax),
					goldmark.WithParserOptions(
						gparser.WithAutoHeadingID(),
					),
					goldmark.WithRendererOptions(
						ghtml.WithHardWraps(),
						ghtml.WithXHTML(),
					),
				)

				if err := md.Convert(source, &buf); err != nil {
					t.Error(err)
				}

				c.Data(http.StatusOK, "text/html; charset=UTF-8", buf.Bytes())
			},

			"/doc": func(c *gin.Context) {

				mdext := parser.CommonExtensions
				psr := parser.NewWithExtensions(mdext)

				htmlFlags := html.CommonFlags | html.HrefTargetBlank | html.TOC
				opts := html.RendererOptions{Flags: htmlFlags}
				renderer := html.NewRenderer(opts)

				out := markdown.ToHTML(source, psr, renderer)
				c.Data(http.StatusOK, "text/html; charset=UTF-8", out)
			},
		},
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		testutil.NewHTTPServer(t, opt)
	}()

	time.Sleep(time.Second)

	resp, err := http.Get("http://:12345/doc")
	if err != nil {
		t.Error(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Error(err)
	}

	t.Log(string(body))

	time.Sleep(600e9)
}

func TestGetDoc(t *testing.T) {

	md, err := Get("mandemo.md")

	if err != nil {
		t.Errorf("mandemo.md not found: %s", err.Error())
	}

	t.Log(md)
}
