package aws

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"

	"github.com/aws/aws-sdk-go/service/ec2"
)

//https://docs.aws.amazon.com/zh_cn/AmazonCloudWatch/latest/monitoring/viewing_metrics_with_cloudwatch.html
//https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/cloudwatch_concepts.html
//region: https://docs.aws.amazon.com/general/latest/gr/rande.html

var (
	accessKey   = `AKIAIKW5ZG4FBRDOYZTQ`
	secretKey   = `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`
	accessToken = ``
)

type stubProvider struct {
	creds   credentials.Value
	expired bool
	err     error
}

func (s *stubProvider) Retrieve() (credentials.Value, error) {
	s.expired = false
	s.creds.ProviderName = "stubProvider"
	return s.creds, s.err
}
func (s *stubProvider) IsExpired() bool {
	return s.expired
}

func getStaticSession() *session.Session {
	cred := credentials.NewStaticCredentials(accessKey, secretKey, accessToken)

	cfg := aws.NewConfig()
	cfg.WithCredentials(cred)
	cfg.WithRegion(`us-west-2`)

	sess, err := session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigDisable,
		Config:            *cfg,
	})

	if err != nil {
		log.Fatalln(err)
	}

	return sess
}

func TestListMetrics(t *testing.T) {

	svc := cloudwatch.New(getStaticSession())

	//metric := `CPUUtilization`
	namespace := `AWS/EC2`
	//dimension := `instanceId`

	result, err := svc.ListMetrics(&cloudwatch.ListMetricsInput{
		//MetricName: aws.String(metric),
		Namespace: aws.String(namespace),
		// Dimensions: []*cloudwatch.DimensionFilter{
		// 	&cloudwatch.DimensionFilter{
		// 		Name: aws.String(dimension),
		// 	},
		// },
	})

	if err != nil {
		log.Fatalln(err)
	}

	log.Println(result.Metrics)

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

	svc := cloudwatch.New(getStaticSession())

	resp, err := svc.GetMetricStatistics(&cloudwatch.GetMetricStatisticsInput{
		MetricName: aws.String(`CPUUtilization`),
		Namespace:  aws.String(`AWS/EC2`),
	})

	if err != nil {
		log.Fatalln(err)
	}

	log.Println(resp.GoString())

}

func TestListAllEC2(t *testing.T) {

	cred := credentials.NewStaticCredentials(accessKey, secretKey, accessToken)

	cfg := aws.NewConfig()
	cfg.WithCredentials(cred)
	cfg.WithRegion(`us-west-2`)

	sess, err := session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigDisable,
		Config:            *cfg,
	})

	if err != nil {
		log.Fatalln(err)
	}

	svc := ec2.New(sess)
	// input := &ec2.DescribeInstancesInput{
	// 	InstanceIds: []*string{
	// 		aws.String("i-1234567890abcdef0"),
	// 	},
	// }

	result, err := svc.DescribeInstances(nil)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
		return
	}

	fmt.Println(result)
}

func TestGetMetrics(t *testing.T) {

	//https://docs.aws.amazon.com/AmazonCloudWatch/latest/APIReference/API_GetMetricData.html

	svc := cloudwatch.New(getStaticSession())

	query1 := &cloudwatch.MetricDataQuery{
		MetricStat: &cloudwatch.MetricStat{
			Metric: &cloudwatch.Metric{
				MetricName: aws.String(`CPUUtilization`),
				Namespace:  aws.String(`AWS/EC2`),
				Dimensions: []*cloudwatch.Dimension{
					&cloudwatch.Dimension{
						Name:  aws.String(`InstanceId`),
						Value: aws.String(`i-0231ef3f85ff1e595`),
					},
				},
			},
			Period: aws.Int64(60),
			Stat:   aws.String(`Average`),
		},
		Id:         aws.String("a1"),
		ReturnData: aws.Bool(true),
	}

	query2 := &cloudwatch.MetricDataQuery{
		MetricStat: &cloudwatch.MetricStat{
			Metric: &cloudwatch.Metric{
				MetricName: aws.String(`DiskReadOps`),
				Namespace:  aws.String(`AWS/EC2`),
				// Dimensions: []*cloudwatch.Dimension{
				// 	&cloudwatch.Dimension{
				// 		Name:  aws.String(``),
				// 		Value: aws.String(``),
				// 	},
				// },
			},
			Period: aws.Int64(60),
			Stat:   aws.String(`Average`),
		},
		Id: aws.String("a2"),
		//ReturnData: aws.Bool(true),
	}
	_ = query2

	params := &cloudwatch.GetMetricDataInput{
		EndTime:           aws.Time(time.Now().UTC().Add(-1 * time.Minute)),
		StartTime:         aws.Time(time.Now().UTC().Add(-21 * time.Minute)),
		MetricDataQueries: []*cloudwatch.MetricDataQuery{query1}, //max 100
	}

	req, resp := svc.GetMetricDataRequest(params)

	err := req.Send()

	if err != nil {
		//log.Fatalln(err)
	}

	_ = resp

	log.Println(resp.GoString())

	//resp.MetricDataResults[0].
}
