// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package httpapi

import (
	"errors"
	"net/http"
	"reflect"
	"strings"
	"time"

	lp "github.com/GuanceCloud/cliutils/lineproto"
	uhttp "github.com/GuanceCloud/cliutils/network/http"
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
)

type IAPIWrite interface {
	feed(string, point.Category, []*point.Point, ...*io.Option) error
	geoInfo(string) map[string]string
}

type apiWriteImpl struct{}

func (x *apiWriteImpl) feed(input string, category point.Category, pts []*point.Point, opt ...*io.Option) error {
	f := io.DefaultFeeder()
	return f.Feed(input, category, pts, opt...)
}

func (x *apiWriteImpl) geoInfo(ip string) map[string]string {
	return geoTags(ip)
}

//nolint:stylecheck
const (
	ArgPrecision            = "precision"
	ArgInput                = "input"
	ArgIgnoreGlobalTags     = "ignore_global_tags"      // deprecated, use IGNORE_GLOBAL_HOST_TAGS
	ArgIgnoreGlobalHostTags = "ignore_global_host_tags" // default enabled
	ArgGlobalElectionTags   = "global_election_tags"    // default disabled

	// echo point in line-protocol or JSON for debugging.
	ArgEchoLineProto = "echo_line_proto"
	ArgEchoJSON      = "echo_json"

	ArgVersion        = "version"
	ArgPipelineSource = "source"

	ArgLoose  = "loose" // Deprecated: default are loose mode
	ArgStrict = "strict"

	defaultInput     = "datakit-http" // 当 API 调用方未亮明自己身份时，默认它作为数据源名称
	DefaultPrecision = "n"
)

func apiWrite(w http.ResponseWriter, req *http.Request, x ...interface{}) (interface{}, error) {
	var body []byte
	var err error

	if x == nil || len(x) != 1 {
		l.Errorf("invalid handler")
		return nil, ErrInvalidAPIHandler
	}

	h, ok := x[0].(IAPIWrite)
	if !ok {
		l.Errorf("not IApiWrite, got %s", reflect.TypeOf(x).String())
		return nil, ErrInvalidAPIHandler
	}

	input := defaultInput
	categoryURL := req.URL.Path

	opts := []point.Option{
		point.WithPrecision(point.NS),
		point.WithTime(time.Now()),
	}

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
		input = "custom_object"
	case point.Security.URL():
		input = "scheck"

	default:
		l.Debugf("invalid category: %q", categoryURL)
		return nil, uhttp.Errorf(ErrInvalidCategory, "invalid URL %q", categoryURL)
	}

	q := req.URL.Query()

	if x := q.Get(ArgInput); x != "" {
		input = x
	}

	if x := q.Get(ArgPrecision); x != "" {
		opts = append(opts, point.WithPrecision(point.PrecStr(x)))
	}

	var version string
	if x := q.Get(ArgVersion); x != "" {
		version = x
	}

	var pipelineSource string
	if x := q.Get(ArgPipelineSource); x != "" {
		pipelineSource = x
	}

	body, err = uhttp.ReadBody(req)
	if err != nil {
		return nil, err
	}

	if len(body) == 0 {
		return nil, ErrEmptyBody
	}

	cntTyp := GetPointEncoding(req.Header)

	pts, err := HandleWriteBody(body, cntTyp, opts...)
	if err != nil {
		if errors.Is(err, point.ErrInvalidLineProtocol) {
			return nil, uhttp.Errorf(ErrInvalidLinePoint, "%s: body(%d bytes)", err.Error(), len(body))
		}
		return nil, err
	}

	if len(pts) == 0 {
		return nil, ErrNoPoints
	}

	// add extra tags
	ignoreGlobalTags := false
	for _, arg := range []string{
		ArgIgnoreGlobalHostTags,
		ArgIgnoreGlobalTags, // deprecated
	} {
		if x := q.Get(arg); x != "" {
			ignoreGlobalTags = true
		}
	}

	if !ignoreGlobalTags {
		appendTags(pts, datakit.GlobalHostTags())
	}

	if x := q.Get(ArgGlobalElectionTags); x != "" {
		appendTags(pts, datakit.GlobalElectionTags())
	}

	l.Debugf("received %d(%s) points from %s, pipeline source: %v",
		len(pts), categoryURL, input, pipelineSource)

	// under strict mode, any point warning are errors
	strict := false
	if q.Get(ArgStrict) != "" {
		strict = true
	}
	if strict {
		for _, pt := range pts {
			if arr := pt.Warns(); len(arr) > 0 {
				switch cntTyp {
				case point.JSON:
					return nil, uhttp.Errorf(ErrInvalidJSONPoint, "%s: %s", arr[0].Type, arr[0].Msg)
				case point.LineProtocol:
					return nil, uhttp.Errorf(ErrInvalidLinePoint, "%s: %s", arr[0].Type, arr[0].Msg)
				case point.Protobuf:
					return nil, uhttp.Errorf(ErrInvalidProtobufPoint, "%s: %s", arr[0].Type, arr[0].Msg)
				}
			}
		}
	}

	feedOpt := &io.Option{Version: version}
	if pipelineSource != "" {
		feedOpt.PlScript = map[string]string{pipelineSource: pipelineSource + ".p"}
	}

	if err := h.feed(input, point.CatURL(categoryURL), pts, feedOpt); err != nil {
		return err, nil
	}

	if q.Get(ArgEchoLineProto) != "" {
		var arr []string
		for _, x := range pts {
			arr = append(arr, x.LineProto())
		}

		return strings.Join(arr, "\n"), nil
	}

	if q.Get(ArgEchoJSON) != "" {
		return pts, nil
	}

	return nil, nil
}

func appendTags(pts []*point.Point, tags map[string]string) {
	for k, v := range tags {
		for _, pt := range pts {
			pt.AddTag(k, v)
		}
	}
}

func HandleWriteBody(body []byte,
	encTyp point.Encoding,
	opts ...point.Option,
) ([]*point.Point, error) {
	switch encTyp {
	case point.JSON:
		dec := point.GetDecoder(point.WithDecEncoding(point.JSON))
		defer point.PutDecoder(dec)

		if pts, err := dec.Decode(body, opts...); err != nil {
			l.Warnf("dec.Decode: %s", err.Error())

			return nil, uhttp.Errorf(ErrInvalidJSONPoint, "%s", err)
		} else {
			return pts, nil
		}
	case point.Protobuf:
		dec := point.GetDecoder(point.WithDecEncoding(point.Protobuf))
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

func getTimeFromInt64(n int64, opt *lp.Option) time.Time {
	if opt != nil {
		switch opt.Precision {
		case "h":
			return time.Unix(n*3600, 0).UTC()
		case "m":
			return time.Unix(n*60, 0).UTC()
		case "s":
			return time.Unix(n, 0).UTC()
		case "ms":
			return time.Unix(0, n*1000).UTC()
		case "u":
			return time.Unix(0, n*1000000).UTC()
		default:
		}
	}

	// nanoseconds
	return time.Unix(0, n).UTC()
}
