package awsbill

import (
	"io/ioutil"
	"log"
	"os"
	"testing"
	"time"

	"github.com/influxdata/toml"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/costexplorer"
)

//https://docs.datadoghq.com/integrations/amazon_billing/#service-checks
//https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/monitor_estimated_charges_with_cloudwatch.html#turning_on_billing_metrics

//AWS billing metrics are available about once every 4 hours.

var (
	cloudwatchCli *cloudwatch.CloudWatch
	billClient    *costexplorer.CostExplorer
)

func defaultAuthProvider() client.ConfigProvider {

	accessKey := os.Getenv("AWS_AK")
	secretKey := os.Getenv("AWS_SK")

	cred := credentials.NewStaticCredentials(accessKey, secretKey, "")

	cfg := aws.NewConfig()
	cfg.WithCredentials(cred) //.WithRegion(endpoints.CnNorth1RegionID)

	sess, err := session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigDisable,
		Config:            *cfg,
	})

	if err != nil {
		log.Fatalf("auth failed: %s", err)
	}

	return sess
}

func getCostClient() *costexplorer.CostExplorer {
	if billClient != nil {
		return billClient
	}

	billClient = costexplorer.New(defaultAuthProvider(), aws.NewConfig())

	return billClient
}

func getCloudwatchClient() *cloudwatch.CloudWatch {

	if cloudwatchCli != nil {
		return cloudwatchCli
	}

	//cli := cloudwatch.New(defaultAuthProvider(), aws.NewConfig().WithRegion(endpoints.CnNorth1RegionID))
	cli := cloudwatch.New(defaultAuthProvider(), aws.NewConfig().WithRegion(endpoints.UsEast1RegionID))
	cloudwatchCli = cli

	return cli
}

func TestConfig(t *testing.T) {
	ag := &AwsInstance{
		AccessKey:    "xxx",
		AccessSecret: "xxx",
		RegionID:     "xxx",
	}

	if data, err := toml.Marshal(ag); err != nil {
		t.Errorf("%s", err)
	} else {
		log.Printf("%s", string(data))
	}
}

func TestListMetricsOfNamespce(t *testing.T) {

	//如果你没有使用该产品，则会返回空
	//aws中: namespace + metricname + dimension 确定一个具体的指标
	//metric := `CPUUtilization`
	namespace := `AWS/Billing`
	//namespace := `AWS/EC2`
	//dimension := `instanceId`

	var token *string
	params := &cloudwatch.ListMetricsInput{
		Namespace: aws.String(namespace),
		// Dimensions: []*cloudwatch.DimensionFilter{
		// 	&cloudwatch.DimensionFilter{
		// 		Name: aws.String(dimension),
		// 	},
		// },
		NextToken: token,
		//MetricName: aws.String(`EstimatedCharges`),
	}

	result, err := getCloudwatchClient().ListMetrics(params)

	if err != nil {
		log.Fatalf("fail to get namespace metrics, %s", err)
	}

	log.Printf("%d, %s", len(result.Metrics), result)

	// stub := &stubProvider{
	// 	creds: credentials.Value{
	// 		AccessKeyID:     "AKID",
	// 		SecretAccessKey: "SECRET",
	// 		SessionToken:    "",
	// 	},
	// 	expired: true,
	// }

	// c := credentials.NewCredentials(stub)
}

func TestMetricStatics(t *testing.T) {

	//https://docs.aws.amazon.com/AmazonCloudWatch/latest/APIReference/API_GetMetricStatistics.html

	svc := getCloudwatchClient()

	resp, err := svc.GetMetricStatistics(&cloudwatch.GetMetricStatisticsInput{
		MetricName: aws.String(`EstimatedCharges`),
		Namespace:  aws.String(`AWS/Billing`),
		Dimensions: []*cloudwatch.Dimension{
			{
				Name:  aws.String(`ServiceName`),
				Value: aws.String(`AmazonEC2`),
			},
			{
				Name:  aws.String(`Currency`),
				Value: aws.String(`USD`),
			},
		},
		EndTime:   aws.Time(time.Now().UTC().Truncate(time.Minute).Add(-1 * time.Hour)),
		StartTime: aws.Time(time.Now().UTC().Truncate(time.Minute).Add(-25*time.Hour - time.Second*40)),
		Period:    aws.Int64(60),
		Statistics: []*string{
			aws.String(`SampleCount`),
			aws.String(`Average`),
			aws.String(`Sum`),
			aws.String(`Minimum`),
			aws.String(`Maximum`),
		},
	})

	if err != nil {
		log.Fatalln(err)
	}

	log.Println(resp.String())

}

func TestGetMetrics(t *testing.T) {

	//https://docs.aws.amazon.com/AmazonCloudWatch/latest/APIReference/API_GetMetricData.html

	svc := getCloudwatchClient()

	query1 := &cloudwatch.MetricDataQuery{
		MetricStat: &cloudwatch.MetricStat{
			Metric: &cloudwatch.Metric{
				MetricName: aws.String(`EstimatedCharges`),
				Namespace:  aws.String(`AWS/Billing`),
				Dimensions: []*cloudwatch.Dimension{
					&cloudwatch.Dimension{
						Name:  aws.String(`ServiceName`),
						Value: aws.String(`AmazonEC2`),
					},
					&cloudwatch.Dimension{
						Name:  aws.String(`Currency`),
						Value: aws.String(`USD`),
					},
				},
			},
			Period: aws.Int64(60 * 60 * 6),
			Stat:   aws.String(`Maximum`),
		},
		Id:         aws.String("a1"),
		ReturnData: aws.Bool(true),
	}

	query2 := &cloudwatch.MetricDataQuery{
		MetricStat: &cloudwatch.MetricStat{
			Metric: &cloudwatch.Metric{
				MetricName: aws.String(`EstimatedCharges`),
				Namespace:  aws.String(`AWS/Billing`),
				Dimensions: []*cloudwatch.Dimension{
					&cloudwatch.Dimension{
						Name:  aws.String(`ServiceName`),
						Value: aws.String(`AWSCostExplorer`),
					},
					&cloudwatch.Dimension{
						Name:  aws.String(`Currency`),
						Value: aws.String(`USD`),
					},
				},
			},
			Period: aws.Int64(60 * 60 * 6),
			Stat:   aws.String(`Maximum`),
		},
		Id:         aws.String("a2"),
		ReturnData: aws.Bool(true),
	}

	params := &cloudwatch.GetMetricDataInput{
		EndTime:           aws.Time(time.Now().UTC().Truncate(time.Minute)),
		StartTime:         aws.Time(time.Now().UTC().Truncate(time.Minute).Add(-30 * 24 * time.Hour)),
		MetricDataQueries: []*cloudwatch.MetricDataQuery{query1, query2}, //max 100
	}

	resp, err := svc.GetMetricData(params)

	if err != nil {
		log.Fatalln(err)
	}

	log.Printf("%s", resp.MetricDataResults)
}

func TestGetDimensionValues(t *testing.T) {
	svc := getCostClient()
	params := &costexplorer.GetDimensionValuesInput{
		Dimension: aws.String(costexplorer.DimensionBillingEntity),
		TimePeriod: &costexplorer.DateInterval{
			Start: aws.String(`2020-02-20`),
			End:   aws.String(`2020-04-20`),
		},
	}

	result, err := svc.GetDimensionValues(params)
	if err != nil {
		log.Fatalf("GetDimensionValues failed, %s", err)
	}

	log.Printf("%s", result)
}

func TestGetCostAndUsage(t *testing.T) {
	svc := getCostClient()
	_ = svc

	params := &costexplorer.GetCostAndUsageInput{
		// Filter: &costexplorer.Expression{
		// 	Dimensions: &costexplorer.DimensionValues{
		// 		Key: aws.String(costexplorer.DimensionUsageType),
		// 		Values: []*string{
		// 			aws.String(``),
		// 		},
		// 	},
		// },
		Metrics: []*string{
			aws.String("UsageQuantity"),
			aws.String("BlendedCost"),
		},
		Granularity: aws.String(`DAILY`),
		TimePeriod: &costexplorer.DateInterval{
			Start: aws.String(`2020-01-02`),
			End:   aws.String(`2020-09-10`),
		},
	}

	result, err := svc.GetCostAndUsage(params)
	if err != nil {
		log.Fatalf("GetCostAndUsage failed, %s", err)
	}

	log.Printf("%s", result)
}

func TestSvr(t *testing.T) {

	ag := newInstance()
	ag.debugMode = true

	if data, err := ioutil.ReadFile("./test.conf"); err != nil {
		log.Fatalf("%s", err)
	} else {
		if toml.Unmarshal(data, ag); err != nil {
			log.Fatalf("%s", err)
		}
	}

	ag.Run()
}
