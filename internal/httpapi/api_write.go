// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/GuanceCloud/cliutils/pipeline/ptinput/ipdb"
	"github.com/GuanceCloud/cliutils/point"

	uhttp "github.com/GuanceCloud/cliutils/network/http"
	plmanager "github.com/GuanceCloud/cliutils/pipeline/manager"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/bufpool"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/plval"
)

type IAPIWrite interface {
	Feed(point.Category, []*point.Point, []dkio.FeedOption) error
}

type apiWriteImpl struct{}

func (x *apiWriteImpl) Feed(category point.Category, pts []*point.Point, opt []dkio.FeedOption) error {
	f := dkio.DefaultFeeder()
	return f.FeedV2(category, pts, opt...)
}

const (
	DefaultPrecision = "n"
)

func apiWrite(_ http.ResponseWriter, req *http.Request, x ...interface{}) (interface{}, error) {
	if x == nil || len(x) != 1 {
		l.Errorf("invalid handler")
		return nil, ErrInvalidAPIHandler
	}

	h, ok := x[0].(IAPIWrite)
	if !ok {
		l.Errorf("not IApiWrite, got %s", reflect.TypeOf(x).String())
		return nil, ErrInvalidAPIHandler
	}

	wr := GetAPIWriteResult()
	defer PutAPIWriteResult(wr)

	if err := wr.APIV1Write(req); err != nil {
		return nil, err
	} else {
		if len(wr.Points) != 0 {
			if err := h.Feed(wr.Category, wr.Points, wr.FeedOptions); err != nil {
				l.Warnf("feed: %s, ignored", err.Error())
			}
		}

		return wr.RespBody, nil
	}
}

func HandleWriteBody(body []byte, encTyp point.Encoding, opts ...point.Option) ([]*point.Point, error) {
	switch encTyp {
	case point.JSON, point.PBJSON:
		dec := point.GetDecoder(point.WithDecEncoding(point.JSON))
		defer point.PutDecoder(dec)

		if pts, err := dec.Decode(body, opts...); err != nil {
			l.Warnf("dec.Decode: %s", err.Error())

			return nil, uhttp.Errorf(ErrInvalidJSONPoint, "%s", err)
		} else {
			return pts, nil
		}

	case point.Protobuf:
		dec := point.GetDecoder(
			point.WithDecEncoding(point.Protobuf),

			// do NOT enable easyproto for memory safe.
			point.WithDecEasyproto(false),
		)

		defer point.PutDecoder(dec)

		pts, err := dec.Decode(body, opts...)
		if err != nil {
			l.Warnf("dec.Decode: %s", err.Error())
			return nil, uhttp.Errorf(ErrInvalidProtobufPoint, "%s", err)
		}
		return pts, nil

	case point.LineProtocol:
		dec := point.GetDecoder(point.WithDecEncoding(point.LineProtocol))
		defer point.PutDecoder(dec)

		pts, err := dec.Decode(body, opts...)
		if err != nil {
			l.Warnf("dec.Decode: %s", err.Error())
			return nil, uhttp.Errorf(ErrInvalidLinePoint, "%s", err)
		}
		return pts, nil

	default:
		dec := point.GetDecoder(point.WithDecEncoding(point.LineProtocol))
		defer point.PutDecoder(dec)

		pts, err := dec.Decode(body, opts...)
		if err != nil {
			l.Warnf("dec.Decode: %s", err.Error())
			return nil, uhttp.Errorf(ErrInvalidLinePoint, "%s", err)
		}
		return pts, nil
	}
}

// GetPointEncoding performs additional processing of request headers
// when processing "rum" data.
func GetPointEncoding(hdr http.Header) point.Encoding {
	cntTyp := hdr.Get("Content-Type")

	switch {
	case strings.Contains(cntTyp, "application/json"):
		return point.JSON
	default:
		return point.HTTPContentType(cntTyp)
	}
}

var wrp sync.Pool

// APIWriteResult used to wrap /v1/write/:category handle result.
type APIWriteResult struct {
	Category    point.Category
	Points      []*point.Point
	FeedOptions []dkio.FeedOption
	RespBody    []byte

	IPInfo *ipdb.IPdbRecord

	PointCallback point.Callback

	input,
	inputVersion,
	plName,
	SrcIP,
	APPID,
	IPStatus,
	LocateStatus string

	IPQuerier plval.IIPQuerier

	globalElectionTags,
	ignoreGlobalTags bool
}

// GetAPIWriteResult get write request handler.
func GetAPIWriteResult() *APIWriteResult {
	if x := wrp.Get(); x == nil {
		return defaultAPIWriteResult()
	} else {
		return x.(*APIWriteResult)
	}
}

func (wr *APIWriteResult) reset() {
	wr.Category = point.UnknownCategory
	wr.Points = wr.Points[:0]
	wr.FeedOptions = wr.FeedOptions[:0]

	wr.PointCallback = nil

	wr.IPQuerier = nil
	wr.IPInfo = nil

	wr.globalElectionTags = false
	wr.input = ""
	wr.inputVersion = ""
	wr.ignoreGlobalTags = false
	wr.SrcIP = ""
	wr.APPID = ""
	wr.IPStatus = ""
	wr.LocateStatus = ""
	wr.plName = ""
}

// PutAPIWriteResult put-back write request handler.
func PutAPIWriteResult(wr *APIWriteResult) {
	wr.reset()
	wrp.Put(wr)
}

func defaultAPIWriteResult() *APIWriteResult {
	return &APIWriteResult{}
}

func (wr *APIWriteResult) ipTags() map[string]string {
	res := map[string]string{
		"city":     plval.IPInfoUnknow,
		"province": plval.IPInfoUnknow,
		"country":  plval.IPInfoUnknow,
		"isp":      plval.IPInfoUnknow,
		"ip":       wr.SrcIP,
	}

	if wr.IPInfo == nil {
		return res
	}

	if len(wr.IPInfo.City) > 0 {
		res["city"] = wr.IPInfo.City
	}
	if len(wr.IPInfo.Region) > 0 {
		res["province"] = wr.IPInfo.Region
	}
	if len(wr.IPInfo.Country) > 0 {
		res["country"] = wr.IPInfo.Country
	}

	if len(wr.IPInfo.Isp) > 0 {
		res["isp"] = wr.IPInfo.Isp
	}
	return res
}

func (wr *APIWriteResult) getIPInfo(req *http.Request) error {
	if wr.IPQuerier == nil {
		return nil
	}

	wr.SrcIP, wr.IPStatus = wr.IPQuerier.GetSourceIP(req)
	if wr.SrcIP == "" {
		return nil
	}

	info, err := wr.IPQuerier.Query(wr.SrcIP)
	if err != nil {
		wr.LocateStatus = plval.LocateStatusGEOFailure
		l.Warnf("IP query failed: %s, ignored", err)
		return err
	}

	if info == nil { // no IP info found
		wr.LocateStatus = plval.LocateStatusGEONil
		return nil
	}

	l.Debugf("IP info(%s): %+#v", wr.SrcIP, info)
	wr.LocateStatus = plval.LocateStatusGEOSuccess
	wr.IPInfo = info
	return nil
}

const (
	argPrecision            = "precision"
	argInput                = "input"
	argIgnoreGlobalTags     = "ignore_global_tags"      // deprecated, use IGNORE_GLOBAL_HOST_TAGS
	argIgnoreGlobalHostTags = "ignore_global_host_tags" // default enabled
	argGlobalElectionTags   = "global_election_tags"    // default disabled
	argNoBlocking           = "noblocking"              // no blocking on feed

	// echo point in line-protocol or JSON for debugging.
	argEchoLineProtoDeprecated = "echo_line_proto"
	argEchoJSONDeprecated      = "echo_json"
	argEcho                    = "echo"
	argDryRun                  = "dry"

	argVersion        = "version"
	argPipelineSource = "source"
	argStrict         = "strict"
)

// APIV1Write handle API /v1/write/:category.
func (wr *APIWriteResult) APIV1Write(req *http.Request) (err error) {
	var (
		categoryURL = req.URL.Path
		body        []byte
		inputName   = "datakit-http"

		opts = []point.Option{
			point.WithTime(time.Now()),
			point.WithCallback(wr.PointCallback),
		}
	)

	switch categoryURL {
	// set specific options on them
	case point.Metric.URL():
		opts = append(opts, point.DefaultMetricOptions()...)
	case point.Logging.URL():
		opts = append(opts, point.DefaultLoggingOptions()...)
	case point.Object.URL():
		opts = append(opts, point.DefaultObjectOptions()...)

	case point.Network.URL(),
		point.Tracing.URL(),
		point.KeyEvent.URL(): // pass

		// set input-name for them
	case point.CustomObject.URL():
		inputName = "custom_object"
	case point.Security.URL():
		inputName = "scheck"

	case point.RUM.URL():
		inputName = "rum"
		if err := wr.getIPInfo(req); err != nil {
			l.Warnf("getIPInfo failed: %s, ignored", err)
		}

		opts = append(opts, point.WithExtraTags(wr.ipTags()))

	default:
		l.Debugf("invalid category: %q", categoryURL)
		return uhttp.Errorf(ErrInvalidCategory, "invalid URL %q", categoryURL)
	}

	wr.Category = point.CatURL(categoryURL)

	q := req.URL.Query()
	if x := q.Get(argInput); x != "" {
		inputName = x
	}

	wr.FeedOptions = append(wr.FeedOptions, dkio.WithInputName(inputName))

	if q.Get(argNoBlocking) == "" {
		wr.FeedOptions = append(wr.FeedOptions, dkio.WithBlocking(true)) // default set block, see #2300.
	}

	wr.input = inputName

	if x := q.Get(argPrecision); x != "" {
		opts = append(opts, point.WithPrecision(point.PrecStr(x)))
	} else {
		opts = append(opts, point.WithPrecision(point.PrecDyn))
	}

	// feed option
	if x := q.Get(argVersion); x != "" {
		wr.FeedOptions = append(wr.FeedOptions, dkio.WithInputVersion(x))
		wr.inputVersion = x
	}

	if x := q.Get(argPipelineSource); x != "" {
		wr.plName = x + ".p"
		wr.FeedOptions = append(wr.FeedOptions, dkio.WithPipelineOption(&plmanager.Option{
			ScriptMap: map[string]string{x: x + ".p"},
		}))
	}

	// read POST body
	var (
		pts    []*point.Point
		cntTyp = GetPointEncoding(req.Header)
	)

	if x := q.Get(argGlobalElectionTags); x != "" {
		wr.FeedOptions = append(wr.FeedOptions, dkio.WithElection(true))
		wr.globalElectionTags = true
	} else {
		// disable global host tags?
		for _, arg := range []string{
			argIgnoreGlobalHostTags,
			argIgnoreGlobalTags, // deprecated
		} {
			if x := q.Get(arg); x != "" {
				wr.FeedOptions = append(wr.FeedOptions, dkio.DisableGlobalTags(true))
				wr.ignoreGlobalTags = true
			}
		}
	}

	buf := bufpool.GetBuffer()
	defer bufpool.PutBuffer(buf)

	if _, err := io.Copy(buf, req.Body); err != nil {
		return err
	}

	body = buf.Bytes()
	if len(body) == 0 {
		return ErrEmptyBody
	}

	if pts, err = HandleWriteBody(body, cntTyp, opts...); err != nil {
		if errors.Is(err, point.ErrInvalidLineProtocol) {
			return uhttp.Errorf(ErrInvalidLinePoint, "%s: body(%d bytes)", err.Error(), len(body))
		} else {
			return err
		}
	}

	if len(pts) == 0 {
		return ErrNoPoints
	}

	if x, ok := pts[0].Get("app_id").(string); ok {
		wr.APPID = x
	}

	l.Debugf("received %d(%s) points from %s, pipeline: %s",
		len(pts), categoryURL, inputName, wr.plName)

	// under strict mode, any point warning are errors
	if q.Get(argStrict) != "" {
		for _, pt := range pts {
			if arr := pt.Warns(); len(arr) > 0 {
				switch cntTyp {
				case point.JSON, point.PBJSON:
					l.Warnf("point warnning: %s", arr[0].Msg)
					return uhttp.Errorf(ErrStrictPoint, "%s: %s", arr[0].Type, arr[0].Msg)
				case point.LineProtocol:
					l.Warnf("point warnning: %s", arr[0].Msg)
					return uhttp.Errorf(ErrStrictPoint, "%s: %s", arr[0].Type, arr[0].Msg)
				case point.Protobuf:
					l.Warnf("point warnning: %s", arr[0].Msg)
					return uhttp.Errorf(ErrStrictPoint, "%s: %s", arr[0].Type, arr[0].Msg)
				}
			}
		}
	}

	wr.RespBody = getEchoOption(pts, q)

	if x := q.Get(argDryRun); x == "" {
		wr.Points = pts
	}

	return nil
}

func getEchoOption(pts []*point.Point, q url.Values) []byte {
	var arr []string

	//nolint: gocritic
	if x := q.Get(argEcho); x != "" {
		switch x {
		case "lp": // line-protocol
		case "json": // simple json
			goto echoSimpleJSON
		case "pbjson": // advanced json(protobuf json)
			goto echoPBJSON
		default:
			return nil
		}
	} else if q.Get(argEchoLineProtoDeprecated) != "" {
		goto echoLineProtocol
	} else if q.Get(argEchoJSONDeprecated) != "" {
		goto echoSimpleJSON
	} else {
		return nil
	}

echoLineProtocol:
	for _, x := range pts {
		arr = append(arr, x.LineProto())
	}

	return []byte(strings.Join(arr, "\n"))

echoSimpleJSON:
	for _, pt := range pts {
		pt.ClearFlag(point.Ppb)
	}

	if j, err := json.Marshal(pts); err != nil {
		return []byte(err.Error())
	} else {
		return j
	}

echoPBJSON:
	for _, pt := range pts {
		pt.SetFlag(point.Ppb)
	}

	if j, err := json.Marshal(pts); err != nil {
		return []byte(err.Error())
	} else {
		return j
	}
}
