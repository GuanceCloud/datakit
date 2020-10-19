package dataclean

import (
	"io/ioutil"
	"net/http"
	"sync"
	"time"

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
    ## http server route path
    ## required
    path = "/dataclean"

    points_lua_files = []

    object_lua_files = []

    # [[inputs.dataclean.crontab_lua_list]]
    # lua_file = ""
    # schedule = ""
`
)

var l = logger.DefaultSLogger(inputName)

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &DataClean{}
	})
}

type DataClean struct {
	Path           string              `toml:"path"`
	PointsLuaFiles []string            `toml:"points_lua_files"`
	ObjectLuaFiles []string            `toml:"object_lua_files"`
	Global         []map[string]string `toml:"crontab_lua_list"`

	ls   *luascript.LuaScript
	cron *luascript.LuaCron

	enable bool
	mut    sync.Mutex
}

func (*DataClean) SampleConfig() string {
	return sampleCfg
}

func (*DataClean) Catalog() string {
	return inputName
}

func (d *DataClean) Run() {
	l = logger.SLogger(inputName)

	if d.initCfg() {
		return
	}

	l.Infof("dataclean input started...")

	for {
		select {
		case <-datakit.Exit.Wait():
			d.mut.Lock()
			d.enable = false
			d.mut.Unlock()

			luascript.Stop()
			return
		default:
		}
	}
}

func (d *DataClean) initCfg() bool {
	var err error
	d.mut = sync.Mutex{}
	d.cron = luascript.NewLuaCron()
	d.ls = luascript.NewLuaScript(2)

	for {
	lable:
		time.Sleep(time.Second)

		select {
		case <-datakit.Exit.Wait():
			return true
		default:
			for _, global := range d.Global {
				err = d.cron.AddLuaFromFile(global["lua_file"], global["schedule"])
				if err != nil {
					l.Error(err)
					goto lable
				}
			}

			err = d.ls.AddLuaCodesFromFile("points", d.PointsLuaFiles)
			if err != nil {
				l.Error(err)
				goto lable
			}

			err = d.ls.AddLuaCodesFromFile("object", d.ObjectLuaFiles)
			if err != nil {
				l.Error(err)
				goto lable
			}

			d.ls.Run()
			d.enable = true
			return false
		}
	}
}

func (d *DataClean) RegHttpHandler() {
	httpd.RegGinHandler("POST", d.Path, d.handle)
}

func (d *DataClean) handle(c *gin.Context) {
	if !d.enable {
		l.Warnf("worker does not exist")
		return
	}

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

		err = d.ls.SendData(p)
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

		err = d.ls.SendData(j)
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
