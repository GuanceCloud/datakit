package dataclean

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"hash"
	"io"
	"net/http"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/ftagent/cfg"
	"gitlab.jiagouyun.com/cloudcare-tools/ftagent/utils"
)

const (
	_ACCESS_KEY_HEAD = "DWAY "

	_TIMEOUT int64 = 15 * 60 * 1e9
)

func abs(n int64) int64 {
	y := n >> 63
	return (n ^ y) - y
}

// sign  HTTP authorization 加密步骤
// signature = base64(hmac-sha1(AccessKeySecret,
//	       method + "\n"
//	       + base64(md5(body)) + "\n"
//	       + Content-Type + "\n"
//	       + Date))
// Date is HTTP head, GMT's RFC1123
func sign(method, ctype, date string, body []byte) string {
	bm := md5.Sum(body)

	var b bytes.Buffer
	b.WriteString(method)
	b.WriteString("\n")
	b.WriteString(base64.StdEncoding.EncodeToString(bm[:]))
	b.WriteString("\n")
	b.WriteString(ctype)
	b.WriteString("\n")
	b.WriteString(date)

	s := b.String()

	hm := hmac.New(func() hash.Hash { return sha1.New() }, []byte(cfg.Cfg.Sk))
	io.WriteString(hm, s)

	sg := base64.StdEncoding.EncodeToString(hm.Sum(nil))

	return _ACCESS_KEY_HEAD + cfg.Cfg.Ak + ":" + sg
}

func verify(route string, r *http.Request, body []byte) (bool, error) {

	for _, rt := range cfg.Cfg.Routes {
		if route == rt.Name && rt.AkOpen {
			goto __goon
		}
	}

	return false, nil

__goon:

	akopen := true

	auth := r.Header.Get("Authorization")
	if auth == "" {
		return akopen, utils.ErrInvalidArgument
	}

	date := r.Header.Get("Date")
	if date == "" {
		date = r.Header.Get(`X-Date`) // sometimes, `Date` can not be set manually
		if date == "" {
			return akopen, utils.ErrAkDenied
		}
	}

	dateTime, err := time.Parse(time.RFC1123, date)
	if err != nil {
		return akopen, utils.ErrAkDenied
	}

	if abs(int64(time.Since(dateTime))) > _TIMEOUT {
		return akopen, utils.ErrAkTimeout
	}

	if auth != sign(r.Method, r.Header.Get("Content-Type"), date, body) {
		return akopen, utils.ErrInvalidAKey
	}

	return akopen, nil
}
