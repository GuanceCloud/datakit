package man

import (
	"html/template"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gobuffalo/packr/v2"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	dkhttp "gitlab.jiagouyun.com/cloudcare-tools/datakit/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	ManualBox   = packr.New("manulas", "./manuals")
	tocTemplate *template.Template
	l           = logger.DefaultSLogger("man")
)

type Input struct {
	InputName    string
	InputSample  string
	Version      string
	ReleaseDate  string
	Measurements []*inputs.MeasurementInfo
}

func Get(name string) (string, error) {
	return ManualBox.FindString(name + ".md")
}

type tocPage struct {
	PageTitle  string
	InputNames []string
}

func InitManPage() error {
	// table of content template
	//data, err := ManualBox.FindString("toc.html")
	//if err != nil {
	//	return err
	//}

	l = logger.SLogger("man")

	tocTemplate = template.Must(template.ParseFiles("toc.html"))

	dkhttp.RegGinHandler("GET", "/man", func(c *gin.Context) {

		v := c.Query("input")
		if v == "" { // TOC page
			data := tocPage{
				PageTitle:  "DataKit Manuals",
				InputNames: []string{},
			}

			if err := tocTemplate.Execute(c.Writer, data); err != nil {
				l.Error(err)
			}
		} else { // input detail page
			data, err := Get(v)
			if err != nil {
				l.Warn(err)
				http.Redirect(c.Writer, c.Request, "/man", 302)
			}

			mdext := parser.CommonExtensions
			psr := parser.NewWithExtensions(mdext)

			htmlFlags := html.CommonFlags | html.HrefTargetBlank | html.TOC
			opts := html.RendererOptions{Flags: htmlFlags}
			renderer := html.NewRenderer(opts)

			out := markdown.ToHTML([]byte(data), psr, renderer)
			c.Data(http.StatusOK, "text/html; charset=UTF-8", out)
		}
	})

	return nil
}
