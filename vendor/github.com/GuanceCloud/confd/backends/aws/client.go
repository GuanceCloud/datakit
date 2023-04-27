package aws

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

type Client struct {
	Client *secretsmanager.Client
	// KeyValue List. just json slice as []string{"{\"username\":\"david\",\"password\":\"EXAMPLE-PASSWORD\"}", ...} .
	KeyValues []string
	// cycle time interval, second.
	CircleInterval int
	ExitWatchCh    chan error
}

func NewAWSClient(accessKeyID, secretAccessKey, region string, circleInterval int) (c *Client, err error) {
	var conf aws.Config
	if accessKeyID == "" {
		// Should use secret file like ~/.aws/config .
		conf, err = config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	} else {
		// Should use accessKeyID & secretAccessKey.
		conf, err = config.LoadDefaultConfig(context.TODO(),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, "")),
			config.WithRegion(region),
		)
	}
	if err != nil {
		return nil, err
	}

	// Create Secrets Manager client.
	svc := secretsmanager.NewFromConfig(conf)

	c = &Client{
		Client:         svc,
		KeyValues:      make([]string, 0),
		CircleInterval: circleInterval,
		ExitWatchCh:    make(chan error, 1),
	}

	// Test get all secretNames.
	input := &secretsmanager.ListSecretsInput{}
	_, err = c.Client.ListSecrets(context.TODO(), input)

	if err != nil {
		return nil, fmt.Errorf("new aws client : %v", err)
	}

	return
}

// GetValues @keys: secretName.
func (c *Client) GetValues(keys []string) (map[string]string, error) {
	kvs := make(map[string]string)

	// Get all secretNames.
	input := &secretsmanager.ListSecretsInput{}
	result, err := c.Client.ListSecrets(context.TODO(), input)
	if err != nil {
		return nil, fmt.Errorf("aws get all secretNames : %v", err)
	}

	// Traverse all the secretNames.
	for _, l := range result.SecretList {
		name := *l.Name

		isPrefix := false
		for _, key := range keys {
			if strings.HasPrefix(name, key) {
				isPrefix = true
				break
			}
		}
		if isPrefix {
			value, err := c.getValue(name)
			if err != nil {
				return nil, err
			}
			kvs[name] = value
		}
	}

	return kvs, nil
}

// getValue get single value by key.
func (c *Client) getValue(name string) (string, error) {
	input := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(name),
		VersionStage: aws.String("AWSCURRENT"), // VersionStage defaults to AWSCURRENT if unspecified.
	}
	result, err := c.Client.GetSecretValue(context.TODO(), input)
	if err != nil {
		return "", fmt.Errorf("aws get getValue : %v", err)
	}

	return *result.SecretString, nil
}

// WatchPrefix will get all aws KV list every CircleInterval second.
// @prefix @keys will all useful
func (c *Client) WatchPrefix(prefix string, keys []string, waitIndex uint64, stopChan chan bool) (uint64, error) {
	prefixes := append([]string{}, keys...)
	if prefix != "" {
		prefixes = append(prefixes, prefix)
	}
	timeNow := time.Now().UTC()
	namesOld := make(map[string]bool) // Old names with the prefix

	// Cycle time interval.
	tick := time.NewTicker(time.Second * time.Duration(c.CircleInterval))
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
		case <-stopChan:
			return waitIndex, fmt.Errorf("stopChan")
		case err := <-c.ExitWatchCh:
			return waitIndex + 1, err
		}

		// Cycle get data to find new data.
		namesNow := make(map[string]bool)
		input := &secretsmanager.ListSecretsInput{}
		result, err := c.Client.ListSecrets(context.TODO(), input)
		if err != nil {
			return waitIndex + 1, fmt.Errorf("aws get ListSecrets : %v", err)
		}

		// Check if add or modify secret.Name.
		for _, secret := range result.SecretList {
			// Check prefix. And if secret.LastChangedDate is new or changed.
			for _, key := range prefixes {
				if strings.HasPrefix(*secret.Name, key) {
					if timeNow.Before(*secret.LastChangedDate) {
						// This secret.LastChangedDate changed.
						return waitIndex + 1, nil
					}
					// Add in namesNow to find deleted secret.Name later.
					namesNow[*secret.Name] = true
				}
			}
		}

		// Check if delete secret.Name.
		for k, _ := range namesOld {
			if _, ok := namesNow[k]; !ok {
				// some name be deleted
				return waitIndex + 1, nil
			}
		}

		// Deep copy.
		for k, _ := range namesNow {
			namesOld[k] = true
		}
	}
}

func (c *Client) Close() {}
