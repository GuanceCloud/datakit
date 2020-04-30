package awstrail

import (
	"log"
	"testing"
	"time"

	"github.com/influxdata/toml"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudtrail"
)

/*refs:
https://docs.amazonaws.cn/awscloudtrail/latest/userguide/cloudtrail-concepts.html#cloudtrail-concepts-events
*/

var (
	//accessKey   = `AKIAJ6J5MR44T3DLI4IQ`
	//secretKey   = `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`
	accessKey   = `AKIA2O3KWILDBBOMNHE3`
	secretKey   = `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`
	accessToken = ``

	apiClient = getClient()
)

func TestConfig(t *testing.T) {
	ag := &AwsTrailAgent{
		AwsTrailInstance: []*AwsTrailInstance{
			&AwsTrailInstance{
				AccessKey:    "xxx",
				AccessSecret: "xxx",
				AccessToken:  "xxx",
				RegionID:     "xxx",
				MetricName:   "xxx",
			},
		},
	}

	if data, err := toml.Marshal(ag); err != nil {
		t.Errorf("%s", err)
	} else {
		log.Printf("%s", string(data))
	}
}

func getClient() *cloudtrail.CloudTrail {

	cred := credentials.NewStaticCredentials(accessKey, secretKey, accessToken)

	cfg := aws.NewConfig()
	cfg.WithCredentials(cred).WithRegion(`cn-north-1`)

	sess, err := session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigDisable,
		Config:            *cfg,
	})

	if err != nil {
		log.Fatalf("auth failed: %s", err)
	}

	cli := cloudtrail.New(sess)

	return cli
}

func TestGetEvents(t *testing.T) {

	now := time.Now().UTC().Truncate(time.Minute)
	param := &cloudtrail.LookupEventsInput{
		EndTime:    aws.Time(now),
		StartTime:  aws.Time(now.Add(-24 * time.Hour)),
		MaxResults: aws.Int64(500),
	}
	resp, err := apiClient.LookupEvents(param)
	if err != nil {
		t.Errorf("LookupEvents failed, %s", err)
	}
	log.Printf("%s", resp.String())
}
