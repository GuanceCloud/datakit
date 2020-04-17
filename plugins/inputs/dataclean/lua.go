package dataclean

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	influxdb "github.com/influxdata/influxdb1-client/v2"
	uhttp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/network/http"
	"gitlab.jiagouyun.com/cloudcare-tools/ftagent/filter"
	"gitlab.jiagouyun.com/cloudcare-tools/ftagent/lscript"
	"gitlab.jiagouyun.com/cloudcare-tools/ftagent/utils"
)

func initLua() (err error) {
	return lscript.Start()
}

func (d *DataClean) LuaClean(contentType string, body []byte, route string, tid string) ([]*influxdb.Point, error) {

	var err error
	var pts []*influxdb.Point

	switch contentType {
	case `application/x-protobuf`:
		pts, err = filter.ParsePromToInflux(body, route)
		if err != nil {
			d.logger.Errorf("[%s] %s", tid, err.Error())
			err = utils.ErrParsePromPointFailed
		}
	case `application/json`:
		pts, err = filter.ParseJsonInflux(body, route)
		if err != nil {
			d.logger.Errorf("[%s] %s", tid, err.Error())
		}
	default:
		pts, err = filter.ParseInflux(body, "n", route)
		if err != nil {
			d.logger.Errorf("[%s] %s", tid, err.Error())
			err = utils.ErrParseInfluxPointFailed
		}
	}

	if err != nil {
		return nil, err
	}

	if len(pts) == 0 {
		d.logger.Errorf("has no valid points")
		err = utils.ErrEmptyBody
		return nil, err
	}

	d.logger.Debugf("send %d points to lua...", len(pts))
	pts, err = lscript.Send(pts, route)
	if err != nil {
		d.logger.Errorf("error from lua, %s", err.Error())
		return nil, err
	}

	d.logger.Debugf("recv %d points from lua", len(pts))

	return pts, nil
}

func (d *DataClean) saveLua(code []byte, force bool, fpath string) error {
	dwPath := filepath.Join(DWLuaPath, fpath)

	finfo, err := os.Stat(dwPath)
	if err == nil { // fpath exists
		if !force {
			return utils.ErrLuaFileExists
		}

		if finfo.IsDir() { // can not override exist dir
			return utils.ErrLuaUploadPathIsDir
		}
	}

	if err := os.MkdirAll(filepath.Dir(dwPath), 0640); err != nil {
		d.logger.Errorf("%s", err.Error())
		return err
	}

	if err := ioutil.WriteFile(dwPath, code, 0640); err != nil {
		d.logger.Errorf("%s", err.Error())
		return err
	}

	return nil
}

func (d *DataClean) apiUploadLua(c *gin.Context) {

	tid := c.Request.Header.Get("X-TraceId")

	if err := d.checkConfigAPIReq(c); err != nil {
		d.logger.Errorf("[%s] %s", tid, err.Error())
		uhttp.HttpErr(c, err)
		return
	}

	force := c.Query("force")
	fpath := c.Query("fpath")
	if path.IsAbs(fpath) {
		d.logger.Errorf("[%s] abs path disabled for lua upload", tid)
		uhttp.HttpErr(c, utils.ErrLuaAbsPathNotAllowed)
		return
	}

	luaCode, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		d.logger.Errorf("[%s] read http content failed, %s", tid, err.Error())
		uhttp.HttpErr(c, utils.ErrHTTPReadError)
		return
	}

	if err := lscript.TryLoad(string(luaCode)); err != nil {
		uhttp.HttpErr(c, uhttp.NewErr(
			errors.New(fmt.Sprintf(`load lua failed: %s`, err.Error())),
			http.StatusBadRequest, utils.ErrNamespace))
		return
	}

	// save file to lua
	if err := d.saveLua(luaCode, force != "", fpath); err != nil {
		d.logger.Errorf("E! [%s] %s", tid, err.Error())
		uhttp.HttpErr(c, err)
		return
	}

	utils.ErrOK.HttpResp(c)
}

type luaFileInfo struct {
	Path     string `json:"path"`
	ABSPath  string `json:"abs_path"`
	HostPath string `json:"host_path"`

	Size          int64    `json:"size"`
	ModTime       string   `json:"modtime_utc"`
	WithinDocker  bool     `json:"within_docker"`
	EnabledRoutes []string `json:"enabled_routes"`
}

func (d *DataClean) listLua() ([]*luaFileInfo, error) {
	lfis := []*luaFileInfo{}

	if _, err := os.Stat(DWLuaPath); err != nil {
		return lfis, nil
	}

	filepath.Walk(DWLuaPath, func(fpath string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("E! %s", err.Error())
			return err
		}

		if info.IsDir() {
			return nil // ignore dir
		}

		if err != nil {
			log.Printf("E! %s", err.Error())
			return err
		}

		hostPath := fpath
		// if config.DKConfig.MainCfg.WithinDocker {
		// 	hostPath = path.Join(fmt.Sprintf("/dataway-data/agnt_%s", cfg.Cfg.UUID),
		// 		`lua`,
		// 		strings.TrimPrefix(fpath, config.DWLuaPath+"/"))
		// }

		lfi := &luaFileInfo{
			Path:     strings.TrimPrefix(fpath, DWLuaPath), // remove dw workdir part
			ABSPath:  fpath,
			HostPath: hostPath,
			//WithinDocker: config.DKConfig.MainCfg.WithinDocker,
			Size:    info.Size(),
			ModTime: info.ModTime().Format(`2006-01-02 15:04:05`),
		}

		// how many routes used that lua?
		for _, r := range d.Routes {
			if r.DisableLua {
				continue
			}

			for _, l := range r.Lua {
				if lfi.Path == l.Path {
					lfi.EnabledRoutes = append(lfi.EnabledRoutes, r.Name)
					break
					// if the lua used under the route mutilple time? ignore anyway
				}
			}
		}

		lfis = append(lfis, lfi)

		return nil
	})

	return lfis, nil
}

func (d *DataClean) apiListLuas(c *gin.Context) {

	tid := c.Request.Header.Get("X-TraceId")
	if err := d.checkConfigAPIReq(c); err != nil {
		d.logger.Errorf("[%s] %s", tid, err.Error())
		uhttp.HttpErr(c, err)
		return
	}

	luaFiles, err := d.listLua()
	if err != nil {
		d.logger.Errorf("[%s] %s", tid, err.Error())
		uhttp.HttpErr(c, err)
		return
	}

	utils.ErrOK.HttpBody(c, luaFiles)
	return
}
