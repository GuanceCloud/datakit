package dataclean

import (
	"io/ioutil"
	"log"
	"net/http"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/luascript"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	httpd "gitlab.jiagouyun.com/cloudcare-tools/datakit/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"

	"github.com/gin-gonic/gin"
)

const (
	inputName = "dataclean"

	defaultMeasurement = "dataclean"

	sampleCfg = `
[[inputs.dataclean]]
    # http server route path
    # required
    path = "/dataclean"

    points_lua_files = []

    object_lua_files = []
`
)

var l = logger.DefaultSLogger(inputName)

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &DataClean{}
	})
}

type DataClean struct {
	Path           string   `toml:"path"`
	PointsLuaFiles []string `toml:"points_lua_files"`
	ObjectLuaFiles []string `toml:"object_lua_files"`
}

func (*DataClean) SampleConfig() string {
	return sampleCfg
}

func (*DataClean) Catalog() string {
	return inputName
}

func (d *DataClean) Run() {
	var err error
	l = logger.SLogger(inputName)

	err = luascript.AddLuaCodesFromFile("points", d.PointsLuaFiles)
	if err != nil {
		log.Println(err)
	}

	err = luascript.AddLuaCodesFromFile("object", d.ObjectLuaFiles)
	if err != nil {
		log.Println(err)
	}

	l.Infof("dataclean input started...")

	luascript.Run()

	for {
		select {
		case <-datakit.Exit.Wait():
			luascript.Stop()
		default:
		}
	}
}

func (d *DataClean) RegHttpHandler() {
	httpd.RegGinHandler("POST", d.Path, handle)
}

func handle(c *gin.Context) {
	category := c.Query("category")

	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		l.Errorf("read body, %s", err.Error())
		goto end
	}
	defer c.Request.Body.Close()

	switch category {
	case io.Metric, io.Logging, io.KeyEvent:
		pts, err := ParsePoints(body, "ns")
		if err != nil {
			l.Errorf("parse points, %s", err.Error())
			goto end
		}

		p, err := NewPointsData("points", category, pts)
		if err != nil {
			l.Errorf("new points data, %s", err.Error())
			goto end
		}

		err = luascript.SendData(p)
		if err != nil {
			l.Error(err)
			goto end
		}

	case io.Object:
		j, err := NewObjectData("object", category, body)
		if err != nil {
			l.Error(err)
			goto end
		}

		err = luascript.SendData(j)
		if err != nil {
			l.Error(err)
			goto end
		}

	default:
		l.Errorf("invalid category")
	}

end:
	c.Writer.WriteHeader(http.StatusOK)
}
