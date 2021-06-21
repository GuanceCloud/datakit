package cmds

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/c-bata/go-prompt"
	"github.com/influxdata/influxdb1-client/models"

	dkhttp "gitlab.jiagouyun.com/cloudcare-tools/datakit/http"
)

func DQL() {
	switch runtime.GOOS {
	case "windows":
		fmt.Println("\n[E] --man do not support Windows")
		return
	}

	// TODO: add suggestions

	c, _ := newCompleter()

	p := prompt.New(
		runDQL,
		c.Complete,
		prompt.OptionTitle("DQL: query DQL in DataKit"),
		prompt.OptionPrefix("dql > "),
	)

	p.Run()
}

func runDQL(txt string) {

	s := strings.Join(strings.Fields(strings.TrimSpace(txt)), " ")
	if s == "" {
		return
	}

	switch strings.ToUpper(s) {
	case "":
		return
	case "Q", "EXIT":
		fmt.Println("Bye!")
		os.Exit(0)
	default:
		lines := []string{}
		if strings.HasSuffix(s, "\\") {
			lines = append(lines, strings.TrimSuffix(s, "\\"))
		} else {
			lines = append(lines, s)
		}

		doDQL(strings.Join(lines, "\n"))
	}
}

func doDQL(s string) {

	q := &dkhttp.QueryRaw{
		Queries: []*dkhttp.SingleQuery{
			&dkhttp.SingleQuery{
				Query: s,
			},
		},
	}

	j, err := json.Marshal(q)
	if err != nil {
		fmt.Printf("[E] %s\n", err)
		return
	}

	l.Debugf("query: %s", string(j))

	resp, err := http.Post("http://localhost:9529/v1/query/raw", "application/json", bytes.NewBuffer(j))
	if err != nil {
		fmt.Printf("[E] %s\n", err)
		return
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("[E] read response body: %s\n", err)
		return
	}

	show(respBody)
}

type queryResult struct {
	Series   []*models.Row `json:"series"`
	RawQuery string        `json:"raw_query,omitempty"`
	Cost     string        `json:"cost,omitempty"`
}

func show(body []byte) {
	r := struct {
		ErrorCode string         `json:"error_code"`
		Message   string         `json:"message"`
		Content   []*queryResult `json:"content"`
	}{}

	jd := json.NewDecoder(bytes.NewReader(body))
	jd.UseNumber()
	if err := jd.Decode(&r); err != nil {
		fmt.Printf("[E] %s\n", err)
		return
	}

	if r.Content == nil {
		fmt.Println("[W] Empty result")
		return
	}

	for _, c := range r.Content {
		doShow(c)
	}
}

func doShow(c *queryResult) {

	rows := 0

	rows = prettyShow(c)

	if c.RawQuery != "" {
		if json.Valid([]byte(c.RawQuery)) {
			var x map[string]interface{}
			if err := json.Unmarshal([]byte(c.RawQuery), &x); err != nil {
				fmt.Printf("[E] %s\n", err)
			} else {
				j, err := json.MarshalIndent(x, "", "    ")
				if err != nil {
					fmt.Printf("[E] %s\n", err)
				} else {
					fmt.Println("---------")
					fmt.Printf("explain:\n%s\n", string(j))
				}
			}
		} else {
			fmt.Println("---------")
			fmt.Printf("explain: %s\n", c.RawQuery)
		}
	}

	fmt.Printf("---------\n%d rows, cost %s\n", rows, c.Cost)
}

func prettyShow(resp *queryResult) int {
	nrows := 0

	if len(resp.Series) == 0 {
		fmt.Println("no data")
		return 0
	}

	for _, s := range resp.Series {
		switch len(s.Columns) {
		case 1:

			if s.Name == "" {
				fmt.Printf("<unknown>\n")
			} else {
				fmt.Printf("%s\n", s.Name)
			}

			fmt.Printf("%s\n", "-------------------")
			for _, val := range s.Values {
				if len(val) == 0 {
					continue
				}

				switch val[0].(type) {
				case string:
					//AddSug(val[0].(string))
				}

				fmt.Printf("%s\n", val[0])
				nrows++
			}

		default:
			colWidth := getMaxColWidth(s)
			fmtStr := fmt.Sprintf("%%%ds%%s", colWidth)
			for _, val := range s.Values {
				nrows++
				fmt.Printf("-----------------[ %d.%s ]-----------------\n", nrows, s.Name)

				for k, v := range s.Tags {
					fmt.Printf(fmtStr+" %s\n", k, "#", v)
					//AddSug(k)
				}

				for colIdx, _ := range s.Columns {
					//if DisableNil && val[colIdx] == nil {
					//	continue
					//}

					col := s.Columns[colIdx]
					if _, ok := s.Tags[col]; !ok {
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
								val[colIdx] = time.Unix(i/1000, 0)
							}
						}

						valFmt := ""
						switch val[colIdx].(type) {
						case time.Time:
							valFmt = "%v\n"
						case json.Number:
							i, err := val[colIdx].(json.Number).Int64()
							if err != nil {
								f, err := val[colIdx].(json.Number).Float64()
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
						case bool:
							valFmt = "%v\n"
						default:
							valFmt = "%v\n"
							// pass
						}

						fmt.Printf(fmtStr+valFmt, col, " ", val[colIdx])
					}
					//AddSug(s.Columns[colIdx])
				}
			}
		}
	}

	return nrows
}

func getMaxColWidth(r *models.Row) int {
	max := 0
	for k, _ := range r.Tags {
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
