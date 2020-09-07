package http

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"sort"
	"strings"
)

type SignOption struct {
	AuthorizationType string
	SignHeaders       []string // in order
	SK                string
	AK                string
	Sign              string
	SignStr           string
}

func DefaultSignOption(authType string, headers []string) *SignOption {

	sort.Strings(headers)

	return &SignOption{
		AuthorizationType: authType,
		SignHeaders:       headers,
	}
}

func (o *SignOption) SignReq(r *http.Request) string {
	signElems := []string{
		r.Method,
	}

	sort.Strings(o.SignHeaders)

	for _, v := range o.SignHeaders {
		signElems = append(signElems, r.Header.Get(v))
	}

	o.SignStr = strings.Join(signElems, "\n")

	h := hmac.New(sha256.New, []byte(o.SK))
	if _, err := h.Write([]byte(o.SignStr)); err != nil {
		panic(err)
	}

	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func (o *SignOption) ParseAuth(r *http.Request) error {

	authHeader := r.Header.Get(`Authorization`)

	parts := strings.Split(authHeader, " ")
	switch len(parts) {
	case 2: //nolint:gomnd
		if parts[0] != o.AuthorizationType {
			goto failed
		}

		signParts := strings.Split(parts[1], `:`)
		if len(signParts) != 2 { //nolint:gomnd
			goto failed
		}

		o.AK = signParts[0]
		o.Sign = signParts[1]
		return nil

	default:
		goto failed
	}

failed:
	return ErrBadAuthHeader
}
