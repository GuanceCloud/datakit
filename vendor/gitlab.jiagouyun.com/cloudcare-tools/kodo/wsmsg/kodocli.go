package wsmsg

import (
	"bytes"
	"crypto/md5"
	"fmt"
	uhttp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/network/http"
	"gitlab.jiagouyun.com/cloudcare-tools/kodo/config"
	"io/ioutil"
	"net/http"
	"time"
)

var (
	authType    = "DATAFLUX"
	authHeaders = []string{`Content-MD5`, `Content-Type`, `Date`}
)

type KoDoCli struct {
	Cli http.Client
}

func NewKoDoCli() *KoDoCli {
	duration, err := time.ParseDuration(config.C.WsKoDo.TimeOut)
	if err != nil {
		l.Warnf("illegal timeout")
		duration = time.Second * 30
	}
	return &KoDoCli{
		Cli: http.Client{Timeout: duration},
	}
}

func (cli *KoDoCli) PostKodo(urlstr string, headers map[string]string, body []byte) error {

	req, err := http.NewRequest("POST", fmt.Sprintf("%s%s", config.C.WsKoDo.Host, urlstr), bytes.NewBuffer(body))
	if err != nil {
		l.Error(err)
		return err
	}

	bodymd5 := fmt.Sprintf("%x", md5.Sum(body))
	req.Header.Set("Date", time.Now().Format(http.TimeFormat))
	req.Header.Set("Content-MD5", bodymd5)

	so := uhttp.DefaultSignOption(authType, authHeaders)
	// use token as AK
	so.AK = headers["X-Token"]
	so.SK = bodymd5

	sign, err := so.SignReq(req)
	if err != nil {
		l.Error(err)
		return err
	}

	req.Header.Set("Authorization", fmt.Sprintf("DATAFLUX %s:%s", so.AK, sign))

	// add external headers
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	// do POST

	resp, err := cli.Cli.Do(req)
	if err != nil {
		l.Error(err)
		return err
	}

	defer resp.Body.Close()

	respbody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		l.Errorf("failed to read Kodo response body: %s", err.Error())
		return err
	}

	switch resp.StatusCode / 100 {
	case 2: // ok
		l.Debugf("req ok: %s, urlstr: %s, token: %s", req.URL.String(), urlstr, headers["X-Token"])

	case 4: // bad request
		l.Warnf("req failed: %s, urlstr, %s, token %s",
			string(respbody), urlstr, headers["X-Token"])
		return nil

	case 5: // kodo error
		l.Errorf("kodo error: %s, urlstr: %s, token: %s", string(respbody), urlstr, headers["X-Token"])
		return fmt.Errorf(string(respbody))
	}

	return nil
}
