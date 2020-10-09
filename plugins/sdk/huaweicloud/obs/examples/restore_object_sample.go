// Copyright 2019 Huawei Technologies Co.,Ltd.
// Licensed under the Apache License, Version 2.0 (the "License"); you may not use
// this file except in compliance with the License.  You may obtain a copy of the
// License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software distributed
// under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR
// CONDITIONS OF ANY KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations under the License.

/**
 * This sample demonstrates how to download an cold object
 * from OBS using the OBS SDK for Go.
 */
package examples

import (
	"fmt"
	"io/ioutil"
	"obs"
	"strings"
	"time"
)

type RestoreObjectSample struct {
	bucketName string
	objectKey  string
	location   string
	obsClient  *obs.ObsClient
}

func newRestoreObjectSample(ak, sk, endpoint, bucketName, objectKey, location string) *RestoreObjectSample {
	obsClient, err := obs.New(ak, sk, endpoint)
	if err != nil {
		panic(err)
	}
	return &RestoreObjectSample{obsClient: obsClient, bucketName: bucketName, objectKey: objectKey, location: location}
}

func (sample RestoreObjectSample) CreateColdBucket() {
	input := &obs.CreateBucketInput{}
	input.Bucket = sample.bucketName
	input.Location = sample.location
	input.StorageClass = obs.StorageClassCold
	_, err := sample.obsClient.CreateBucket(input)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Create cold bucket:%s successfully!\n", sample.bucketName)
	fmt.Println()
}

func (sample RestoreObjectSample) CreateObject() {
	input := &obs.PutObjectInput{}
	input.Bucket = sample.bucketName
	input.Key = sample.objectKey
	input.Body = strings.NewReader("Hello OBS")

	_, err := sample.obsClient.PutObject(input)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Create object:%s successfully!\n", sample.objectKey)
	fmt.Println()
}

func (sample RestoreObjectSample) RestoreObject() {
	input := &obs.RestoreObjectInput{}
	input.Bucket = sample.bucketName
	input.Key = sample.objectKey
	input.Days = 1
	input.Tier = obs.RestoreTierExpedited

	_, err := sample.obsClient.RestoreObject(input)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Create object:%s successfully!\n", sample.objectKey)
	fmt.Println()
}

func (sample RestoreObjectSample) GetObject() {
	input := &obs.GetObjectInput{}
	input.Bucket = sample.bucketName
	input.Key = sample.objectKey

	output, err := sample.obsClient.GetObject(input)
	if err != nil {
		panic(err)
	}
	defer func(){
		errMsg := output.Body.Close()
		if errMsg != nil{
			panic(errMsg)
		}
	}()
	fmt.Println("Object content:")
	body, err := ioutil.ReadAll(output.Body)
	if err != nil{
		panic(err)
	}
	fmt.Println(string(body))
	fmt.Println()
}

func (sample RestoreObjectSample) DeleteObject() {
	input := &obs.DeleteObjectInput{}
	input.Bucket = sample.bucketName
	input.Key = sample.objectKey
	_, err := sample.obsClient.DeleteObject(input)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Delete object:%s successfully!\n", input.Key)
	fmt.Println()
}

func RunRestoreObjectSample() {
	const (
		endpoint   = "https://your-endpoint"
		ak         = "*** Provide your Access Key ***"
		sk         = "*** Provide your Secret Key ***"
		bucketName = "bucket-test-cold"
		objectKey  = "object-test"
		location   = "yourbucketlocation"
	)

	sample := newRestoreObjectSample(ak, sk, endpoint, bucketName, objectKey, location)

	fmt.Println("Create a new cold bucket for demo")
	sample.CreateColdBucket()

	sample.CreateObject()

	sample.RestoreObject()

	// Wait 6 minutes to get the object
	time.Sleep(time.Duration(6*60) * time.Second)

	sample.GetObject()

	sample.DeleteObject()
}
