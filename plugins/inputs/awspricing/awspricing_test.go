package awspricing

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/pricing"
)

var (
	accessKey = `AKIAJ6J5MR44T3DLI4IQ`
	secretKey = `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`
	//accessKey   = `AKIA2O3KWILDBBOMNHE3`
	//secretKey   = `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`
	accessToken = ``

	//priceClient *cloudwatch.CloudWatch
	priceClient *pricing.Pricing
)

func defaultAuthProvider() client.ConfigProvider {

	cred := credentials.NewStaticCredentials(accessKey, secretKey, accessToken)

	cfg := aws.NewConfig()
	cfg.WithCredentials(cred) //.WithRegion(`cn-north-1`)

	sess, err := session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigDisable,
		Config:            *cfg,
	})

	if err != nil {
		log.Fatalf("auth failed: %s", err)
	}

	return sess
}

func getPricingClient() *pricing.Pricing {
	if priceClient != nil {
		return priceClient
	}
	priceClient = pricing.New(defaultAuthProvider(), aws.NewConfig().WithRegion("us-east-1"))
	return priceClient
}

func TestDescribeServices(t *testing.T) {
	svc := getPricingClient()
	input := &pricing.DescribeServicesInput{
		FormatVersion: aws.String("aws_v1"),
		MaxResults:    aws.Int64(10),
		//ServiceCode:   aws.String("AmazonEC2"),
	}
	result, err := svc.DescribeServices(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case pricing.ErrCodeInternalErrorException:
				fmt.Println(pricing.ErrCodeInternalErrorException, aerr.Error())
			case pricing.ErrCodeInvalidParameterException:
				fmt.Println(pricing.ErrCodeInvalidParameterException, aerr.Error())
			case pricing.ErrCodeNotFoundException:
				fmt.Println(pricing.ErrCodeNotFoundException, aerr.Error())
			case pricing.ErrCodeInvalidNextTokenException:
				fmt.Println(pricing.ErrCodeInvalidNextTokenException, aerr.Error())
			case pricing.ErrCodeExpiredNextTokenException:
				fmt.Println(pricing.ErrCodeExpiredNextTokenException, aerr.Error())
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

	log.Printf("%s", result)
}
