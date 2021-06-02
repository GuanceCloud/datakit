package dialtesting

// HTTP dialer testing
// auth: tanb
// date: Fri Feb  5 13:17:00 CST 2021

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/cdproto/security"
	"github.com/chromedp/chromedp"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
)

type RecordedSteps struct {
	ActionName    string `json:"action_name"`
	ActionContent string `json:"action_content"`
}

type OptRequest struct {
	IgnoreServerCertificateError bool                   `json:"ignore_server_certificate_error"`
	Headers                      map[string]interface{} `json:"headers,omitempty"`
	Cookies                      string                 `json:"cookies,omitempty"`
	Proxy                        string                 `json:"proxy,omitempty"`
	Auth                         *HTTPOptAuth           `json:"auth,omitempty"`
	DisableCors                  bool                   `json:"disable_cors"`
}

type HeadlessAdvanceOption struct {
	RequestOptions *OptRequest `json:"request_options"`
	Secret         *HTTPSecret `json:"secret,omitempty"`
}

type HeadlessTask struct {
	ExternalID      string                 `json:"external_id"`
	Name            string                 `json:"name"`
	AK              string                 `json:"access_key"`
	Method          string                 `json:"method"`
	URL             string                 `json:"url"`
	PostURL         string                 `json:"post_url"`
	CurStatus       string                 `json:"status"`
	Frequency       string                 `json:"frequency"`
	Region          string                 `json:"region"` // 冗余进来，便于调试
	OwnerExternalID string                 `json:"owner_external_id,omitempty"`
	Steps           []*RecordedSteps       `json:"recorded_steps,omitempty"`
	Tags            map[string]string      `json:"tags,omitempty"`
	Labels          []string               `json:"labels,omitempty"`
	AdvanceOptions  *HeadlessAdvanceOption `json:"advance_options_headless,omitempty"`
	UpdateTime      int64                  `json:"update_time,omitempty"`

	ticker *time.Ticker

	hasAddSpecialSteps bool
	linedatas          string
}

func (t *HeadlessTask) UpdateTimeUs() int64 {
	return t.UpdateTime
}

func (t *HeadlessTask) Clear() {
	t.linedatas = ``
}

func (t *HeadlessTask) ID() string {
	if t.ExternalID == `` {
		return cliutils.XID("dtst_")
	}
	return fmt.Sprintf("%s_%s", t.AK, t.ExternalID)
}

func (t *HeadlessTask) GetOwnerExternalID() string {
	return t.OwnerExternalID
}

func (t *HeadlessTask) SetOwnerExternalID(exid string) {
	t.OwnerExternalID = exid
}

func (t *HeadlessTask) SetRegionId(regionId string) {
	t.Region = regionId
}

func (t *HeadlessTask) SetAk(ak string) {
	t.AK = ak
}

func (t *HeadlessTask) SetStatus(status string) {
	t.CurStatus = status
}

func (t *HeadlessTask) SetUpdateTime(ts int64) {
	t.UpdateTime = ts
}

func (t *HeadlessTask) Stop() error {
	return nil
}

func (t *HeadlessTask) Status() string {
	return t.CurStatus
}

func (t *HeadlessTask) Ticker() *time.Ticker {
	return t.ticker
}

func (t *HeadlessTask) Class() string {
	return ClassHeadless
}

func (t *HeadlessTask) MetricName() string {
	return `` //TODO
}

func (t *HeadlessTask) PostURLStr() string {
	return t.PostURL
}

func (t *HeadlessTask) GetFrequency() string {
	return t.Frequency
}

func (t *HeadlessTask) GetLineData() string {
	return t.linedatas
}

func (t *HeadlessTask) GetResults() (tags map[string]string, fields map[string]interface{}) {
	tags = map[string]string{
		"name": t.Name,
		"url":  t.URL,
	}

	return
}

func (t *HeadlessTask) RegionName() string {
	return t.Region
}

func (t *HeadlessTask) AccessKey() string {
	return t.AK
}

func (t *HeadlessTask) Check() error {
	// TODO: check task validity
	if t.ExternalID == "" {
		return fmt.Errorf("external ID missing")
	}

	// advance options
	opt := t.AdvanceOptions

	// proxy options
	if opt != nil && opt.RequestOptions != nil && opt.RequestOptions.Proxy != "" { // see https://stackoverflow.com/a/14663620/342348
		_, err := url.Parse(opt.RequestOptions.Proxy)
		if err != nil {
			log.Println(err)
			return err
		}
	}

	return nil
}

var allowHeaders = []string{
	"Content-Type",
	"Content-Length",
	"Accept-Encoding",
	"X-CSRF-Token",
	"Authorization",
	"accept",
	"origin",
	"Cache-Control",
	"X-Requested-With",
}

func (t *HeadlessTask) Run() error {

	t.Clear()

	disableCors := false
	if t.AdvanceOptions != nil && t.AdvanceOptions.RequestOptions != nil {
		disableCors = t.AdvanceOptions.RequestOptions.DisableCors
	}

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("disable-web-security", disableCors),
		chromedp.Flag("disable-gpu", true),
	)

	if t.AdvanceOptions != nil && t.AdvanceOptions.RequestOptions != nil && t.AdvanceOptions.RequestOptions.Proxy != `` {
		opts = append(opts, chromedp.ProxyServer(t.AdvanceOptions.RequestOptions.Proxy))
	}

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	// create context
	ctx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	defer cancel()

	header := map[string]interface{}{}
	if t.AdvanceOptions != nil && t.AdvanceOptions.RequestOptions != nil {
		for k, v := range t.AdvanceOptions.RequestOptions.Headers {
			header[k] = v
			allowHeaders = append(allowHeaders, k)
		}
	}

	if t.AdvanceOptions != nil && t.AdvanceOptions.RequestOptions != nil && t.AdvanceOptions.RequestOptions.Auth != nil {
		auth := base64.StdEncoding.EncodeToString([]byte(t.AdvanceOptions.RequestOptions.Auth.Username + ":" + t.AdvanceOptions.RequestOptions.Auth.Password))
		header["Authorization"] = "Basic " + auth
	}

	if t.AdvanceOptions != nil && t.AdvanceOptions.RequestOptions != nil && t.AdvanceOptions.RequestOptions.Cookies != `` {
		header["Cookie"] = t.AdvanceOptions.RequestOptions.Cookies
	}

	actions := []chromedp.Action{}

	if len(header) != 0 {
		actions = append(actions, network.Enable())
		actions = append(actions, network.SetExtraHTTPHeaders(network.Headers(header)))
	}

	if t.AdvanceOptions != nil && t.AdvanceOptions.RequestOptions != nil {
		actions = append(actions, security.Enable())
		actions = append(actions, security.SetIgnoreCertificateErrors(t.AdvanceOptions.RequestOptions.IgnoreServerCertificateError))
	}

	actions = append(actions, chromedp.Navigate(t.URL))

	var res []string
	for _, step := range t.Steps {
		switch step.ActionName {
		case `insertjs`:

			expr := step.ActionContent
			actions = append(actions, chromedp.ActionFunc(func(ctx context.Context) error {
				_, exp, err := runtime.Evaluate(expr).Do(ctx)
				if err != nil {
					log.Println(err)
					return err
				}
				if exp != nil {
					return exp
				}
				return nil
			}))

		case `evaluate`:
			actions = append(actions, chromedp.Evaluate(step.ActionContent, &res))

		case `waitvisible`:
			actions = append(actions, chromedp.WaitVisible(step.ActionContent))

		case `waitready`:
			actions = append(actions, chromedp.WaitReady(step.ActionContent))

		case `sleep`:
			ts, err := time.ParseDuration(step.ActionContent)
			if err != nil {
				log.Println(err)
				return err
			}
			actions = append(actions, chromedp.Sleep(ts))

		case `click`, `sendkeys`, `screenshot`: //TODO
		default:
		}
	}

	// for _, action := range actions {
	// 	log.Printf("headless start run actions: %+#v  %d  %v", action, len(actions), disableCors)
	// }

	err := chromedp.Run(ctx, actions...)
	if err != nil {
		return err
	}

	t.linedatas = strings.Join(res, "\n")

	return nil
}

const insertjs = `(function (h, o, u, n, d) {
    h = h[d] = h[d] || {
      q: [],
      onReady: function (c) {
        h.q.push(c)
      }
    }
    d = o.createElement(u)
    d.async = 1
    d.src = n
    n = o.getElementsByTagName(u)[0]
    n.parentNode.insertBefore(d, n)
  })(
    window,
    document,
    'script',
    'https://zhuyun-static-files-production.oss-cn-hangzhou.aliyuncs.com/browser-sdk/v2/dataflux-rum-headless.js',
    'DATAFLUX_RUM_HEADLESS'
  )
  DATAFLUX_RUM_HEADLESS.onReady(function () {
    DATAFLUX_RUM_HEADLESS.init({
      trackLoadedVisibleElement: 'testheadless',
    })
  })`

const getdata = `window.DATAFLUX_RUM_HEADLESS && DATAFLUX_RUM_HEADLESS.getInternalData()`
const waitid = `#testheadless`

func (t *HeadlessTask) rumSpecialSteps() {

	t.Steps = append(t.Steps, &RecordedSteps{
		ActionName:    `insertjs`,
		ActionContent: insertjs,
	})

	t.Steps = append(t.Steps, &RecordedSteps{
		ActionName:    `waitvisible`,
		ActionContent: waitid,
	})

	t.Steps = append(t.Steps, &RecordedSteps{
		ActionName:    `evaluate`,
		ActionContent: getdata,
	})

	t.hasAddSpecialSteps = true
}

func (t *HeadlessTask) CheckResult() (reasons []string) {
	return
}

func (t *HeadlessTask) Init() error {

	// setup frequency
	du, err := time.ParseDuration(t.Frequency)
	if err != nil {
		return err
	}
	if t.ticker != nil {
		t.ticker.Stop()
	}
	t.ticker = time.NewTicker(du)

	if strings.ToLower(t.CurStatus) == StatusStop {
		return nil
	}

	//当前headless主要做browse rum 性能指标采集
	if !t.hasAddSpecialSteps {
		t.rumSpecialSteps()
	}

	// TODO: more checking on task validity

	return nil
}
