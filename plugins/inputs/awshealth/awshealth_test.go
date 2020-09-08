package awshealth

import (
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/health"
)

//https://docs.aws.amazon.com/zh_cn/AmazonCloudWatch/latest/monitoring/viewing_metrics_with_cloudwatch.html
//https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/cloudwatch_concepts.html
//region: https://docs.aws.amazon.com/general/latest/gr/rande.html

var (
	//accessKey   = `AKIAJ6J5MR44T3DLI4IQ`
	//secretKey   = `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`
	//accessKey   = `AKIA2O3KWILDBBOMNHE3`
	//secretKey   = `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`
	accessKey   = `AKIA2O3KWILDFXX6F72U`
	secretKey   = `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`
	accessToken = ``

	cloudwatchCli *cloudwatch.CloudWatch
	healthCli     *health.Health
)

func defaultAuthProvider() client.ConfigProvider {

	cred := credentials.NewStaticCredentials(accessKey, secretKey, accessToken)

	cfg := aws.NewConfig()
	cfg.WithCredentials(cred).WithRegion(endpoints.CnNorth1RegionID)

	sess, err := session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigDisable,
		Config:            *cfg,
	})

	if err != nil {
		log.Fatalf("auth failed: %s", err)
	}

	return sess
}

func getCloudwatchClient() *cloudwatch.CloudWatch {

	if cloudwatchCli != nil {
		return cloudwatchCli
	}

	cli := cloudwatch.New(defaultAuthProvider())
	cloudwatchCli = cli

	return cli
}

func getHealthClient() *health.Health {
	if healthCli != nil {
		return healthCli
	}
	healthCli = health.New(defaultAuthProvider(), aws.NewConfig().WithRegion(endpoints.CnNorth1RegionID))
	return healthCli
}

func TestGetEvents(t *testing.T) {

	var maxres int64 = 20
	evCat1 := "issue"
	var token *string
	params := &health.DescribeEventTypesInput{
		MaxResults: &maxres,
		Filter: &health.EventTypeFilter{
			EventTypeCategories: []*string{&evCat1},
		},
		NextToken: token,
	}
	_ = params
	types, err := getHealthClient().DescribeEventTypes(params)
	if err != nil {
		log.Fatalf("DescribeEventTypes failed: %s", err)
	}
	log.Printf("%s", types.String())
}

func TestListMetricsOfNamespce(t *testing.T) {

	//如果你没有使用该产品，则会返回空
	//metric := `CPUUtilization`
	namespace := `AWS/EC2`
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
		//MetricName: nil,
	}

	result, err := getCloudwatchClient().ListMetrics(params)

	if err != nil {
		log.Fatalf("fail to get namespace metrics, %s", err)
	}

	log.Printf("count: %d", len(result.Metrics))

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

// func TestMetricStatics(t *testing.T) {

// 	//https://docs.aws.amazon.com/AmazonCloudWatch/latest/APIReference/API_GetMetricStatistics.html

// 	svc := cloudwatch.New(getStaticSession())

// 	resp, err := svc.GetMetricStatistics(&cloudwatch.GetMetricStatisticsInput{
// 		MetricName: aws.String(`CPUUtilization`),
// 		Namespace:  aws.String(`AWS/EC2`),
// 	})

// 	if err != nil {
// 		log.Fatalln(err)
// 	}

// 	log.Println(resp.GoString())

// }

// func TestListAllEC2(t *testing.T) {

// 	cred := credentials.NewStaticCredentials(accessKey, secretKey, accessToken)

// 	cfg := aws.NewConfig()
// 	cfg.WithCredentials(cred)
// 	cfg.WithRegion(`us-west-2`)

// 	sess, err := session.NewSessionWithOptions(session.Options{
// 		SharedConfigState: session.SharedConfigDisable,
// 		Config:            *cfg,
// 	})

// 	if err != nil {
// 		log.Fatalln(err)
// 	}

// 	svc := ec2.New(sess)
// 	// input := &ec2.DescribeInstancesInput{
// 	// 	InstanceIds: []*string{
// 	// 		aws.String("i-1234567890abcdef0"),
// 	// 	},
// 	// }

// 	result, err := svc.DescribeInstances(nil)
// 	if err != nil {
// 		if aerr, ok := err.(awserr.Error); ok {
// 			switch aerr.Code() {
// 			default:
// 				fmt.Println(aerr.Error())
// 			}
// 		} else {
// 			// Print the error, cast err to awserr.Error to get the Code and
// 			// Message from an error.
// 			fmt.Println(err.Error())
// 		}
// 		return
// 	}

// 	fmt.Println(result)
// }

// func TestGetMetrics(t *testing.T) {

// 	//https://docs.aws.amazon.com/AmazonCloudWatch/latest/APIReference/API_GetMetricData.html

// 	svc := cloudwatch.New(getStaticSession())

// 	query1 := &cloudwatch.MetricDataQuery{
// 		MetricStat: &cloudwatch.MetricStat{
// 			Metric: &cloudwatch.Metric{
// 				MetricName: aws.String(`CPUUtilization`),
// 				Namespace:  aws.String(`AWS/EC2`),
// 				Dimensions: []*cloudwatch.Dimension{
// 					&cloudwatch.Dimension{
// 						Name:  aws.String(`InstanceId`),
// 						Value: aws.String(`i-0231ef3f85ff1e595`),
// 					},
// 				},
// 			},
// 			Period: aws.Int64(60),
// 			Stat:   aws.String(`Average`),
// 		},
// 		Id:         aws.String("a1"),
// 		ReturnData: aws.Bool(true),
// 	}

// 	query2 := &cloudwatch.MetricDataQuery{
// 		MetricStat: &cloudwatch.MetricStat{
// 			Metric: &cloudwatch.Metric{
// 				MetricName: aws.String(`DiskReadOps`),
// 				Namespace:  aws.String(`AWS/EC2`),
// 				// Dimensions: []*cloudwatch.Dimension{
// 				// 	&cloudwatch.Dimension{
// 				// 		Name:  aws.String(``),
// 				// 		Value: aws.String(``),
// 				// 	},
// 				// },
// 			},
// 			Period: aws.Int64(60),
// 			Stat:   aws.String(`Average`),
// 		},
// 		Id: aws.String("a2"),
// 		//ReturnData: aws.Bool(true),
// 	}
// 	_ = query2

// 	params := &cloudwatch.GetMetricDataInput{
// 		EndTime:           aws.Time(time.Now().UTC().Add(-1 * time.Minute)),
// 		StartTime:         aws.Time(time.Now().UTC().Add(-21 * time.Minute)),
// 		MetricDataQueries: []*cloudwatch.MetricDataQuery{query1}, //max 100
// 	}

// 	req, resp := svc.GetMetricDataRequest(params)

// 	err := req.Send()

// 	if err != nil {
// 		//log.Fatalln(err)
// 	}

// 	_ = resp

// 	log.Println(resp.GoString())

// 	//resp.MetricDataResults[0].
// }
