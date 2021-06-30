package azurecms

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"

	"github.com/Azure/azure-sdk-for-go/services/preview/monitor/mgmt/2019-06-01/insights"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/azure/auth"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

func TestGetMetricMetas(t *testing.T) {

	//insights.DefaultBaseURI
	//cli := insights.NewMetricsClientWithBaseURI(`https://management.chinacloudapi.cn`, `7b9b5f30-1590-4e09-b1d0-90547d257b6b`)

	cli := insights.NewMetricDefinitionsClientWithBaseURI(`https://management.chinacloudapi.cn`, `7b9b5f30-1590-4e09-b1d0-90547d257b6b`)

	var err error

	settings, err := auth.GetSettingsFromEnvironment()
	if err != nil {
		return
	}
	settings.Environment = azure.ChinaCloud
	settings.Values[auth.Resource] = settings.Environment.ResourceManagerEndpoint
	cli.Authorizer, err = settings.GetAuthorizer()
	if err != nil {
		log.Fatalf("auth failed, %s", err)
	}

	resourceID := `subscriptions/7b9b5f30-1590-4e09-b1d0-90547d257b6b/resourceGroups/gp1/providers/Microsoft.Compute/virtualMachines/aaa`
	metricnamespace := "" // `Microsoft.Compute/virtualMachines`

	ctx, cancelfun := context.WithCancel(context.Background())
	_ = cancelfun
	res, err := cli.List(ctx, resourceID, metricnamespace)

	if err != nil {
		log.Printf("endpoint: %s", cli.BaseURI)
		log.Fatalf("****%s", err)
	} else {
		log.Printf("Value legth: %v", len(*res.Value))

		if res.Value != nil {
			for _, md := range *res.Value {
				line := []string{}
				if md.Name != nil {
					line = append(line, fmt.Sprintf("name=%s", *(*md.Name).Value))
				}
				if md.Namespace != nil {
					line = append(line, fmt.Sprintf("namespace=%s", *md.Namespace))
				}
				line = append(line, fmt.Sprintf("Unit=%s", md.Unit))
				line = append(line, fmt.Sprintf("PrimaryAggregationType=%s", md.PrimaryAggregationType))
				if md.SupportedAggregationTypes != nil {
					line = append(line, fmt.Sprintf("SupportedAggregationTypes=%s", *md.SupportedAggregationTypes))
				}
				if md.IsDimensionRequired != nil {
					line = append(line, fmt.Sprintf("IsDimensionRequired=%v", *md.IsDimensionRequired))
				}
				if md.Dimensions != nil {
					dimensions := []string{}
					for _, dm := range *md.Dimensions {
						dimensions = append(dimensions, *dm.Value)
					}
					line = append(line, fmt.Sprintf("Dimensions=%s", dimensions))
				}
				if md.MetricAvailabilities != nil {
					timegrains := []string{}
					for _, ma := range *md.MetricAvailabilities {
						//retation := *ma.Retention
						timegrain := *ma.TimeGrain
						timegrains = append(timegrains, timegrain)
					}
					line = append(line, fmt.Sprintf("MetricAvailabilities=%s", timegrains))
				}

				log.Printf("%s", strings.Join(line, ","))
			}
		} else {
			log.Printf("empty response")
		}

	}
}

func TestGetMetrics(t *testing.T) {

	ag := loadCfg("test.conf").(*azureInstance)

	ev := azure.PublicCloud

	cli := insights.NewMetricsClientWithBaseURI(ev.ResourceManagerEndpoint, ag.SubscriptionID)

	var err error

	settings := auth.EnvironmentSettings{
		Values: map[string]string{},
	}
	settings.Values[auth.SubscriptionID] = ag.SubscriptionID
	settings.Values[auth.TenantID] = ag.TenantID
	settings.Values[auth.ClientID] = ag.ClientID
	settings.Values[auth.ClientSecret] = ag.ClientSecret

	settings.Environment = ev
	settings.Values[auth.Resource] = settings.Environment.ResourceManagerEndpoint

	cli.Authorizer, err = settings.GetAuthorizer()
	if err != nil {
		log.Fatalf("auth failed, %s", err)
	}

	resourceID := ag.Resource[0].ResourceID
	timespan := "2021-01-14T09:34:03Z/2021-01-14T10:34:03Z" // `2020-03-11T06:38:00Z/2020-03-11T06:39:00Z` //默认最近一小时
	interval := `PT1M`
	metricnames := "Percentage CPU" //`Disk Read Bytes` // `Network In Total` //
	aggregation := ""
	//var top int32 = 10
	orderby := ""
	filter := ""
	var resultType insights.ResultType = ""
	//apiVersion := "2018-01-01"
	metricnamespace := "Microsoft.Compute/virtualMachines" // `Microsoft.Compute/virtualMachines`

	ctx, cancelfun := context.WithTimeout(context.Background(), time.Second*40)
	_ = cancelfun
	log.Printf("start list")
	res, err := cli.List(ctx, resourceID, timespan, &interval, metricnames, aggregation, nil, orderby, filter, resultType, metricnamespace)

	if err != nil {
		log.Fatalf("%s", err)
	}

	summary := []string{}
	if res.Resourceregion != nil {
		summary = append(summary, fmt.Sprintf("Resourceregion=%s", *res.Resourceregion))
	}

	if res.Interval != nil {
		summary = append(summary, fmt.Sprintf("Interval=%s", *res.Interval))
	}
	log.Printf("summary: %s", summary)

	for _, m := range *res.Value {
		metricName := *(*m.Name).Value
		metricUnit := string(m.Unit)

		tms := *m.Timeseries

		log.Printf("Timeseries(%s) length: %v", metricName, len(tms))

		for _, tm := range tms {

			if tm.Metadatavalues != nil {
				// log.Printf("Metadatavalues len=%v", len(*tm.Metadatavalues))

				for _, metaitem := range *tm.Metadatavalues {
					kv := *(*metaitem.Name).Value
					kv += "=" + *(metaitem.Value)
					//line = append(line, kv)
					log.Printf("Metadatavalues: %s", kv)
				}
			}

			if *tm.Data == nil {
				continue
			}

			for _, mv := range *tm.Data {

				line := []string{metricName, "unit:" + metricUnit}

				if mv.Average != nil {
					line = append(line, fmt.Sprintf("Average=%v", *mv.Average))
				}
				if mv.Count != nil {
					line = append(line, fmt.Sprintf("Count=%v", *mv.Count))
				}
				if mv.Total != nil {
					line = append(line, fmt.Sprintf("Total=%v", *mv.Total))
				}
				if mv.Maximum != nil {
					line = append(line, fmt.Sprintf("Maximum=%v", *mv.Maximum))
				}
				if mv.Minimum != nil {
					line = append(line, fmt.Sprintf("Minimum=%v", *mv.Minimum))
				}
				//metricTime := time.Now()
				if mv.TimeStamp != nil {
					line = append(line, fmt.Sprintf("TimeStamp=%v", *mv.TimeStamp))
					//metricTime, _ = time.Parse(`2006-01-02T15:04:05Z`, "2020-03-10T10:30:00Z")
					//log.Printf("time is %v", metricTime)
				}
				log.Printf("%s\n", line)
			}
		}
	}
}

func TestInput(t *testing.T) {

	ag := loadCfg("test.conf").(*azureInstance)
	ag.mode = "debug"
	ag.Run()
}

func loadCfg(file string) inputs.Input {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatalf("read config file failed: %s", err.Error())
		return nil
	}

	tbl, err := toml.Parse(data)
	if err != nil {
		log.Fatalf("parse toml file failed, %s", err)
		return nil
	}

	creator := func() inputs.Input {
		ac := &azureInstance{}
		ac.ctx, ac.cancelFun = context.WithCancel(context.Background())
		return ac
	}

	for field, node := range tbl.Fields {

		switch field {
		case "inputs":
			stbl, ok := node.(*ast.Table)
			if !ok {
				log.Fatalf("ignore bad toml node")
			} else {
				for inputName, v := range stbl.Fields {
					input, err := tryUnmarshal(v, inputName, creator)
					if err != nil {
						log.Fatalf("unmarshal input %s failed: %s", inputName, err.Error())
						return nil
					}
					return input
				}
			}

		default: // compatible with old version: no [[inputs.xxx]] header
			input, err := tryUnmarshal(node, "aa", creator)
			if err != nil {
				log.Fatalf("unmarshal input failed: %s", err.Error())
				return nil
			}
			return input
		}
	}

	return nil
}

func tryUnmarshal(tbl interface{}, name string, creator inputs.Creator) (inputs.Input, error) {

	tbls := []*ast.Table{}

	switch t := tbl.(type) {
	case []*ast.Table:
		tbls = tbl.([]*ast.Table)
	case *ast.Table:
		tbls = append(tbls, tbl.(*ast.Table))
	default:
		err := fmt.Errorf("invalid toml format: %v", t)
		return nil, err
	}

	for _, t := range tbls {
		input := creator()

		err := toml.UnmarshalTable(t, input)
		if err != nil {
			log.Fatalf("toml unmarshal failed: %v", err)
			return nil, err
		}
		return input, nil

	}
	return nil, nil
}
