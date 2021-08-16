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
	highlighting "github.com/yuin/goldmark-highlighting"

	mathjax "github.com/litao91/goldmark-mathjax"

	//fhtml "github.com/alecthomas/chroma/formatters/html"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	gparser "github.com/yuin/goldmark/parser"
	ghtml "github.com/yuin/goldmark/renderer/html"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
)

func TestMarkdownToHTML(t *testing.T) {

	source := `
| col    | col2  | col3 | col4 |
| ---:   | :---- | ---  | ---- |
| row1 | xxx   | int  | none |
| row2 | yyy   | int  | none |
	`

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

				if err := md.Convert([]byte(source), &buf); err != nil {
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

				out := markdown.ToHTML([]byte(source), psr, renderer)
				//c.Data(http.StatusOK, "text/html; charset=UTF-8", out)
				c.Data(http.StatusOK, "", out)
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

	md, err := Get("demo.md")

	if err != nil {
		t.Errorf("demo.md not found: %s", err.Error())
	}

	t.Log(md)
}

// func TestMarkdownTemplate(t *testing.T) {
// 	md, err := Get("demo")
// 	if err != nil {
// 		t.Error(err)
// 	}

// 	temp, err := template.New("demo").Parse(md)
// 	if err != nil {
// 		panic(err)
// 	}

// 	i := &Input{
// 		InputName:   "demo",
// 		InputSample: "[[inputs.demo]]",
// 		Measurements: []*inputs.MeasurementInfo{
// 			&inputs.MeasurementInfo{
// 				Name: "measurement-1",
// 				Fields: map[string]*inputs.FieldInfo{
// 					"disk_usage": &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "disk used bytes"},
// 					"mem_usage":  &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent, Desc: "mem used percent"},
// 					"net_rx_tx":  &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "network send/receive bytes"},
// 				},
// 				Tags: map[string]*inputs.TagInfo{
// 					"host": &inputs.TagInfo{Desc: "host name"},
// 				},
// 			},
// 		},
// 	}

// 	var buf bytes.Buffer
// 	temp.Execute(&buf, i)

// 	result := termmarkdown.Render(buf.String(), 80, 6)
// 	t.Log(string(result))
// }
