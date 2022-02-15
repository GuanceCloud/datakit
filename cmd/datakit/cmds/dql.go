package cmds

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/c-bata/go-prompt"
	"github.com/influxdata/influxdb1-client/models"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	dkhttp "gitlab.jiagouyun.com/cloudcare-tools/datakit/http"
)

var (
	disableNil  = false
	echoExplain = false
	history     []string
	MaxHistory  = 5000

	defaultJSONIndent = "  "
	temporaryToken    = "" // 临时token，不允许跨工作空间 所以只能赋值一次
)

const (
	dqlraw = "/v1/query/raw"
)

type dqlCmd struct {
	json     bool
	autoJson bool
	verbose  bool

	csv           string
	forceWriteCSV bool

	dqlString string
	token     string
	host      string
	log       string

	dqlcli *http.Client
}

func runDQLFlags() error {
	dc := &dqlCmd{
		json:          *flagDQLJSON,
		autoJson:      *flagDQLAutoJSON,
		dqlString:     *flagDQLString,
		token:         *flagDQLToken,
		csv:           *flagDQLCSV,
		forceWriteCSV: *flagDQLForce,
		host:          *flagDQLDataKitHost,
		verbose:       *flagDQLVerbose,
		log:           *flagDQLLogPath,
	}

	if err := dc.prepare(); err != nil {
		return fmt.Errorf("dc.prepare: %s", err)
	}

	dc.run()
	return nil
}

func (dc *dqlCmd) prepare() error {
	// use localhost token configured in <datakit-install-path>/conf.d/datakit.conf
	if dc.token == "" {
		dc.token = config.GetToken()
	}

	// query local datakit(NOTE: it must be running at the time)
	if dc.host == "" {
		dc.host = config.Cfg.HTTPAPI.Listen
	}

	dc.dqlcli = &http.Client{}
	setCmdRootLog(dc.log)

	// TODO: we should ping the datakit here.
	return nil
}

type dqlResp struct {
	ErrorCode string         `json:"error_code"`
	Message   string         `json:"message"`
	Content   []*queryResult `json:"content"`
}

func (dc *dqlCmd) run() {
	if dc.dqlString != "" {
		dc.runDQL(dc.dqlString)
		return
	}

	// run DQL interactively
	c, _ := newCompleter()
	suggestions = append(suggestions, dqlSuggestions...)

	loadHistory()

	p := prompt.New(
		dc.runDQL,
		c.Complete,
		prompt.OptionTitle("DQL: query DQL in DataKit"),
		prompt.OptionPrefix("dql > "),
		prompt.OptionHistory(history),
	)

	p.Run()
}

func loadHistory() {
	homedir, err := os.UserHomeDir()
	if err != nil {
		l.Errorf("UserHomeDir(): %s", err.Error())
		return
	}

	histpath := filepath.Join(homedir, ".dql_history")

	if _, err = os.Stat(histpath); err != nil {
		l.Warnf("history file %s not found", histpath)
		return
	}

	data, err := ioutil.ReadFile(filepath.Clean(histpath))
	if err != nil {
		l.Warnf("read history failed: %s", err.Error())
		return
	}

	history = strings.Split(string(data), "\n")
}

func addHistory(s ...string) {
	history = append(history, s...)
	if len(history) > MaxHistory {
		dumpHistory()
	}
}

func dumpHistory() {
	homedir, err := os.UserHomeDir()
	if err != nil {
		l.Errorf("UserHomeDir(): %s", err.Error())
		return
	}

	if len(history) > MaxHistory {
		history = history[len(history)-MaxHistory/2:] // trim older-histories
	}

	if err := ioutil.WriteFile(filepath.Join(homedir, ".dql_history"),
		[]byte(strings.Join(history, "\n")), os.ModePerm); err != nil {
		l.Errorf("update history error: %s", err.Error())
	}
}

func updateHistoryOnExit() {
	if len(history) == 0 {
		return
	}

	dumpHistory()
}

func (dc *dqlCmd) runDQL(txt string) {
	s := strings.Join(strings.Fields(strings.TrimSpace(txt)), " ")
	if s == "" {
		return
	}

	addHistory(s)

	switch strings.ToLower(s) {
	case "":
		return
	case "q", "exit":
		output("Bye!\n")
		updateHistoryOnExit()
		os.Exit(0)

		// nil
	case "disable_nil":
		disableNil = true
		return
	case "enable_nil":
		disableNil = false
		return

		// explain
	case "echo_explain":
		echoExplain = true
		return
	case "echo_explain_off":
		echoExplain = false
		return

	default:
		lines := []string{}
		if strings.HasSuffix(s, "\\") {
			lines = append(lines, strings.TrimSuffix(s, "\\"))
		} else {
			lines = append(lines, s)
		}

		resp, err := dc.doDQL(strings.Join(lines, "\n"))
		if err == nil {
			dc.show(resp)
		}
	}
}

// runSingleDQL Perform a single DQL query statement.
func (dc *dqlCmd) runSingleDQL(s string) {
	if dc.csv != "" {
		if err := dc.prepareCsv(); err != nil {
			errorf("prepareCsv: %s", err)
			return
		}
	}

	resp, err := dc.doDQL(s)
	if err != nil {
		return
	}

	if dc.csv != "" {
		if len(resp) > 1 {
			errorf("CSV export only support single DQL.\n")
			return
		}

		if err := writeToCsv(resp[0].Series, dc.csv); err != nil {
			errorf("writeToCsv:%s", err)
			return
		}
	}

	if (dc.csv != "" && dc.verbose) || dc.csv == "" {
		dc.show(resp)
	}
}

func (dc *dqlCmd) prepareCsv() error {
	f, err := os.Stat(dc.csv)
	if err == nil {
		if f.IsDir() {
			return fmt.Errorf("the specified path is a directory")
		}

		if !dc.forceWriteCSV {
			return fmt.Errorf("file %s exists", dc.csv)
		}
	} else if err = os.MkdirAll(filepath.Dir(dc.csv), os.ModePerm); err != nil {
		return err
	}
	return nil
}

// convertStrings Convert Interface arrays to String array which in the query result.
func convertStrings(value []interface{}) []string {
	res := []string{}
	for k, v := range value {
		if v == nil {
			res = append(res, "")
			continue
		}
		switch x := v.(type) {
		case string:
			res = append(res, x)

		case json.Number:
			res = append(res, x.String())

		case bool:
			res = append(res, fmt.Sprintf("%v", x))

		default:
			warnf("unknown key value, %d:%v", k, x)
		}
	}
	return res
}

// writeToCsv Format the query results and write to CSV files.
func writeToCsv(series []*models.Row, csvPath string) error {
	csvFile, err := os.OpenFile(filepath.Clean(csvPath),
		os.O_WRONLY|os.O_TRUNC|os.O_CREATE, os.ModePerm)
	if err != nil {
		return err
	}
	defer csvFile.Close() //nolint:errcheck,gosec

	writer := csv.NewWriter(csvFile)

	sortColumns(series[0])

	columns := []string{"name"}
	columns = append(columns, series[0].Columns...)
	if err := writer.Write(columns); err != nil {
		l.Errorf("write file failed:%s", err)
	}

	for _, serial := range series {
		sortColumns(serial)

		for k := range serial.Values {
			values := []string{serial.Name}
			res := convertStrings(serial.Values[k])
			values = append(values, res...)
			if err := writer.Write(values); err != nil {
				l.Errorf("write file failed:%s", err)
			}
		}
	}

	writer.Flush()
	if err = writer.Error(); err != nil {
		warnf("writer flush failed,error: %s", err)
	}
	return nil
}

func (dc *dqlCmd) doDQL(s string) ([]*queryResult, error) {
	q := &dkhttp.QueryRaw{
		EchoExplain: echoExplain,
		Token:       dc.token,
		Queries: []*dkhttp.SingleQuery{
			{
				Query: s,
			},
		},
	}

	if dc.token != "" {
		q.Token = dc.token
	}

	if temporaryToken != "" && s != "show_workspaces()" {
		q.Token = temporaryToken
	}

	j, err := json.Marshal(q)
	if err != nil {
		errorf("%s\n", err.Error())
		return nil, err
	}

	l.Debugf("dql request: %s", string(j))

	if strings.HasPrefix(s, "use") {
		switchToken(s)
		return nil, fmt.Errorf("err not nil") // only switch token
	}

	req, err := http.NewRequest("POST",
		fmt.Sprintf("http://%s%s", dc.host, dqlraw), bytes.NewBuffer(j))
	if err != nil {
		errorf("http.NewRequest: %s\n", err.Error())
		return nil, err
	}

	resp, err := dc.dqlcli.Do(req)
	if err != nil {
		errorf("httpcli.Do: %s\n", err.Error())
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		errorf("ioutil.ReadAll: %s\n", err.Error())
		return nil, err
	}

	defer resp.Body.Close() //nolint:errcheck
	if resp.StatusCode/100 != 2 {
		r := struct {
			Err string `json:"error_code"`
			Msg string `json:"message"`
		}{}

		if err := json.Unmarshal(body, &r); err != nil {
			errorf("json.Unmarshal: %s\n", err.Error())
			errorf("body: %s\n", string(body))
			return nil, err
		}

		l.Warnf("body: %s", string(body))

		errorf("[%s] %s\n", r.Err, r.Msg)
		return nil, fmt.Errorf("%s", r.Err)
	}

	r := dqlResp{}

	jd := json.NewDecoder(bytes.NewReader(body))
	jd.UseNumber()

	if err := jd.Decode(&r); err != nil {
		errorf("%s\n", err.Error())
		return nil, err
	}
	content, _ := json.Marshal(r.Content)
	l.Debugf("json content:%s", string(content))

	if s == "show_workspaces()" {
		cacheWorkerSpace(r.Content)
	}
	return r.Content, nil
}

type queryResult struct {
	Series   []*models.Row `json:"series"`
	RawQuery string        `json:"raw_query,omitempty"`
	Cost     string        `json:"cost,omitempty"`
}

func (dc *dqlCmd) show(content []*queryResult) {
	if content == nil {
		errorf("Empty result\n")
		return
	}
	for _, c := range content {
		l.Debugf("connent len: %d", len(content))
		dc.doShow(c)
	}
}

func (dc *dqlCmd) doShow(c *queryResult) {
	rows := 0

	switch {
	case dc.json:
		j, err := json.MarshalIndent(c, "", defaultJSONIndent)
		if err != nil {
			errorf("%s\n", err.Error())
			return
		}

		if len(j) == 0 {
			return
		}

		output("%s\n", j)
	default:
		rows = dc.prettyShow(c)
	}

	if c.RawQuery != "" {
		if json.Valid([]byte(c.RawQuery)) {
			var x map[string]interface{}
			if err := json.Unmarshal([]byte(c.RawQuery), &x); err != nil {
				errorf("%s\n", err)
			} else {
				j, err := json.MarshalIndent(x, "", "    ")
				if err != nil {
					errorf("%s\n", err)
					return
				}

				infof("---------\n")
				infof("explain:\n%s\n", string(j))
			}
		} else {
			infof("---------\n")
			infof("explain:%s\n", c.RawQuery)
		}
	}

	infof("---------\n%d rows, %d series, cost %s\n", rows, len(c.Series), c.Cost)
}

// Not used.
func sortColumns(r *models.Row) {
	colMap := map[string]int{}
	for i, col := range r.Columns {
		if _, ok := colMap[col]; ok {
			// duplicate colums(means tag/field with the same key)
			// terminate sorting
			return
		}

		colMap[col] = i
	}

	sort.Strings(r.Columns)
	valArray := [][]interface{}{}

	for _, vals := range r.Values {
		newVals := []interface{}{}
		for _, col := range r.Columns {
			newVals = append(newVals, vals[colMap[col]])
		}

		valArray = append(valArray, newVals)
	}

	r.Values = valArray
}

//nolint:deadcode,unused
func showRow(r *models.Row) {
	output("%d columns\n", len(r.Columns))

	for i, col := range r.Columns {
		if i < len(r.Columns)-1 {
			output("%s, ", col)
		} else {
			output("%s\n", col)
		}
	}

	output("%d values\n", len(r.Values))
	for _, vals := range r.Values {
		output("value width: %d\n", len(vals))
		for i, v := range vals {
			if i < len(vals)-1 {
				output("%v, ", v)
			} else {
				output("%v\n", v)
			}
		}
	}
}

// Not used
//nolint:deadcode,unused
func tableShow(resp *queryResult) int {
	nrows := 0

	if len(resp.Series) == 0 {
		warnf("no data\n")
		return 0
	}

	for _, row := range resp.Series {
		sortColumns(row)
		showRow(row)
		nrows += len(row.Values)
	}

	return nrows
}

func (dc *dqlCmd) prettyShowRow(s *models.Row, val []interface{}, fmtStr string) {
	for colIdx := range s.Columns {
		if disableNil && val[colIdx] == nil {
			continue
		}

		col := s.Columns[colIdx]

		if v, ok := s.Tags[col]; ok {
			output(fmtStr+" %s\n", col, "#", v) // decorate tag key with a `#'
			addSug(col)
			continue
		}

		if col == "time" {
			if _, ok := val[colIdx].(json.Number); !ok {
				l.Error("invalid time: %v", val[colIdx])
				val[colIdx] = fmt.Sprintf("%v", val[colIdx])
			} else {
				i, err := val[colIdx].(json.Number).Int64()
				if err != nil {
					l.Error("parse time failed: %v", err)
					continue
				}

				// convert ms to second
				val[colIdx] = time.Unix(i/1000, 0) //nolint
			}
		}

		valFmt := ""
		switch v := val[colIdx].(type) {
		case time.Time:
			valFmt = "%v\n"
		case json.Number:
			i, err := v.Int64()
			if err != nil {
				f, err := v.Float64()
				if err != nil {
					l.Warn(err)
				} else {
					valFmt = "%.6f\n"
					val[colIdx] = f
				}
			} else {
				val[colIdx] = i
				valFmt = "%d\n"
			}

		case string:
			valFmt = "'%s'\n"
			if dc.autoJson {
				dst := &bytes.Buffer{}
				if err := json.Indent(dst, []byte(v), "", defaultJSONIndent); err == nil {
					val[colIdx] = dst.String()
					valFmt = "<<<<< json \n" + "%s\n" + ">>>>> end of json\n"
				}
			}
		case bool:
			valFmt = "%v\n"
		default:
			valFmt = "%v\n"
			// pass
		}

		output(fmtStr+valFmt, col, " ", val[colIdx])
		addSug(s.Columns[colIdx])
	}
}

func (dc *dqlCmd) prettyShow(resp *queryResult) int {
	nrows := 0

	if len(resp.Series) == 0 {
		warnf("no data\n")
		return 0
	}

	for si, s := range resp.Series {
		switch len(s.Columns) {
		case 1:

			if s.Name == "" {
				output("<unknown>\n")
			} else {
				output("%s\n", s.Name)
			}

			output("%s\n", "-------------------")
			for _, val := range s.Values {
				if len(val) == 0 {
					continue
				}

				if str, ok := val[0].(string); ok {
					addSug(str)
				}

				output("%s\n", val[0])
				nrows++
			}

		default:
			colWidth := getMaxColWidth(s)
			sortColumns(s)

			fmtStr := fmt.Sprintf("%%%ds%%s", colWidth)

			for _, val := range s.Values {
				output("-----------------[ r%d.%s.s%d ]-----------------\n", nrows+1, s.Name, si+1)
				nrows++
				dc.prettyShowRow(s, val, fmtStr)
			}
		}
		if s.Name == "show_workspaces" {
			infof("do dql to change workSpace 'use tkn_xxxxxxxxxx' \n")
		}
	}

	return nrows
}

func getMaxColWidth(r *models.Row) int {
	max := 0
	for k := range r.Tags {
		if len(k) > max {
			max = len(k)
		}
	}

	for _, col := range r.Columns {
		if len(col) > max {
			max = len(col)
		}
	}

	return max
}

var (
	liveSug        = map[string]bool{}
	dqlSuggestions = []prompt.Suggest{
		{Text: "AND", Description: "..."},
		{Text: "AS", Description: "..."},
		{Text: "ASC", Description: "..."},
		{Text: "BY", Description: "..."},
		{Text: "DESC", Description: "..."},
		{Text: "FALSE", Description: "..."},
		{Text: "FILL()", Description: "fill default value"},
		{Text: "LIMIT", Description: "..."},
		{Text: "LINEAR", Description: "..."},
		{Text: "NIL", Description: "..."},
		{Text: "OFFSET", Description: "..."},
		{Text: "OR", Description: "..."},
		{Text: "ORDER", Description: "..."},
		{Text: "PREVIOUS", Description: "..."},
		{Text: "re()", Description: "regex expressionn"},
		{Text: "SLIMIT", Description: "..."},
		{Text: "SOFFSET", Description: "..."},
		{Text: "TRUE", Description: "..."},
		{Text: "tz()", Description: "timezone function"},

		{Text: "metric::", Description: "Metric namespace"},
		{Text: "object::", Description: "Object namespace"},
		{Text: "custom_object::", Description: "Custom object namespace"},
		{Text: "event::", Description: "Event namespace"},
		{Text: "logging::", Description: "Logging namespace"},
		{Text: "tracing::", Description: "Tracing namespace"},
		{Text: "rum::", Description: "RUM namespace"},
		{Text: "security::", Description: "Security namespace"},
		{Text: "network::", Description: "eBPF-network namespace"},

		{Text: "M::", Description: "metric namespace"},
		{Text: "N::", Description: "eBPF-network namespace"},
		{Text: "O::", Description: "object namespace"},
		{Text: "CO::", Description: "custom object namespace"},
		{Text: "E::", Description: "event namespace"},
		{Text: "L::", Description: "logging namespace"},
		{Text: "T::", Description: "tracing namespace"},
		{Text: "R::", Description: "RUM namespace"},
		{Text: "S::", Description: "Security namespace"},

		// functions
		{Text: "show_measurement()", Description: "show all metric names"},
		{Text: "show_field_key()", Description: "show metric fields"},
		{Text: "show_tag_key()", Description: "show metric tags"},
		{Text: "show_tag_value(keyin=[])", Description: "show metric tag values"},

		{Text: "show_object_class()", Description: "show object classes"},
		{Text: "show_object_field()", Description: "show object fields"},

		{Text: "show_logging_source()", Description: "show logging sources"},
		{Text: "show_logging_field()", Description: "show logging fields"},

		{Text: "show_event_source()", Description: "show event sources"},
		{Text: "show_event_field()", Description: "show event fields"},

		{Text: "show_tracing_service()", Description: "show tracing services"},
		{Text: "show_tracing_field()", Description: "show tracing fields"},

		{Text: "show_rum_type()", Description: "show RUM types"},
		{Text: "show_rum_field(rum-type-value)", Description: "show RUM type fields"},

		{Text: "show_security_source()", Description: "show security categories, same as show_security_category()"},
		{Text: "show_security_category()", Description: "show security categories"},
		{Text: "show_security_field()", Description: "show security fields"},

		{Text: "show_network_source()", Description: "show eBPF network source"},
		{Text: "show_network_field()", Description: "show eBPF network fields"},
		{Text: "show_workspaces()", Description: "Show all default and readonly workspaces"},
		{Text: "use", Description: "use other workspace,as :'use tkn_xxxxxx' "},

		{Text: "show_workspaces()", Description: "Show all default and readonly workspaces"},
		{Text: "use", Description: "use other workspace,as :'use tkn_xxxxxx' "},

		{Text: "avg()", Description: ""},
		{Text: "bottom()", Description: ""},
		{Text: "count()", Description: ""},
		{Text: "derivative()", Description: ""},
		{Text: "difference()", Description: ""},
		{Text: "distinct()", Description: ""},
		{Text: "first()", Description: ""},
		{Text: "float()", Description: ""},
		{Text: "int()", Description: ""},
		{Text: "last()", Description: ""},
		{Text: "log()", Description: ""},
		{Text: "match()", Description: ""},
		{Text: "max()", Description: ""},
		{Text: "min()", Description: ""},
		{Text: "moving_average()", Description: ""},
		{Text: "non_negative_derivative()", Description: ""},
		{Text: "percent()", Description: ""},
		{Text: "sum()", Description: ""},
		{Text: "top()", Description: ""},
		{Text: "dataflux__dql.CHAIN()", Description: ""},
		{Text: "dataflux__dql.EXEC_EXPR()", Description: ""},
		{Text: "dataflux__dql.EXEC_FORMULA()", Description: ""},
		{Text: "dataflux__dql.ABS()", Description: ""},
		{Text: "dataflux__dql.CUMSUM()", Description: ""},
		{Text: "dataflux__dql.INTEGRAL()", Description: ""},
		{Text: "dataflux__dql.LOG2()", Description: ""},
		{Text: "dataflux__dql.LOG10()", Description: ""},
		{Text: "dataflux__dql.TOP()", Description: ""},
		{Text: "dataflux__dql.BOTTOM()", Description: ""},
		{Text: "dataflux__dql.DIFF()", Description: ""},
		{Text: "dataflux__dql.MIN()", Description: ""},
		{Text: "dataflux__dql.MAX()", Description: ""},
		{Text: "dataflux__dql.AVG()", Description: ""},
		{Text: "dataflux__dql.SUM()", Description: ""},
		{Text: "dataflux__dql.FIRST()", Description: ""},
		{Text: "dataflux__dql.LAST()", Description: ""},

		// settings
		{Text: "echo_explain", Description: "echo backend query"},
		{Text: "echo_explain_off", Description: "disable echo backend query"},
		{Text: "disable_nil", Description: "disable show nil values"},
		{Text: "enable_nil", Description: "show nil values"},
		{Text: "exit", Description: "exit dfcli"},

		// new outer funcs
		{Text: "abs", Description: "math.abs"},
		{Text: "cumsum", Description: "cumsum"},
		{Text: "log10", Description: "log10"},
		{Text: "log2", Description: "log2"},
		{Text: "non_negative_difference", Description: "positive difference"},
	}
)

func addSug(key string) {
	if ok := liveSug[key]; !ok {
		suggestions = append(suggestions, prompt.Suggest{
			Text: key, Description: "",
		})
		liveSug[key] = true
	}
}
