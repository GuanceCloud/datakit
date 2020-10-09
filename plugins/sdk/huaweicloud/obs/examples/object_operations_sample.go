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
 * This sample demonstrates how to do object-related operations
 * (such as create/delete/get/copy object, do object ACL)
 * on OBS using the OBS SDK for Go.
 */
package examples

import (
	"fmt"
	"io/ioutil"
	"obs"
	"strings"
)

type ObjectOperationsSample struct {
	bucketName string
	objectKey  string
	location   string
	obsClient  *obs.ObsClient
}

func newObjectOperationsSample(ak, sk, endpoint, bucketName, objectKey, location string) *ObjectOperationsSample {
	obsClient, err := obs.New(ak, sk, endpoint)
	if err != nil {
		panic(err)
	}
	return &ObjectOperationsSample{obsClient: obsClient, bucketName: bucketName, objectKey: objectKey, location: location}
}

func (sample ObjectOperationsSample) CreateBucket() {
	input := &obs.CreateBucketInput{}
	input.Bucket = sample.bucketName
	input.Location = sample.location
	_, err := sample.obsClient.CreateBucket(input)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Create bucket:%s successfully!\n", sample.bucketName)
	fmt.Println()
}

func (sample ObjectOperationsSample) GetObjectMeta() {
	input := &obs.GetObjectMetadataInput{}
	input.Bucket = sample.bucketName
	input.Key = sample.objectKey
	output, err := sample.obsClient.GetObjectMetadata(input)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Object content-type:%s\n", output.ContentType)
	fmt.Printf("Object content-length:%d\n", output.ContentLength)
	fmt.Println()
}

func (sample ObjectOperationsSample) CreateObject() {
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

func (sample ObjectOperationsSample) GetObject() {
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

func (sample ObjectOperationsSample) CopyObject() {
	input := &obs.CopyObjectInput{}
	input.Bucket = sample.bucketName
	input.Key = sample.objectKey + "-back"
	input.CopySourceBucket = sample.bucketName
	input.CopySourceKey = sample.objectKey

	_, err := sample.obsClient.CopyObject(input)
	if err != nil {
		panic(err)
	}
	fmt.Println("Copy object successfully!")
	fmt.Println()
}

func (sample ObjectOperationsSample) DoObjectAcl() {
	input := &obs.SetObjectAclInput{}
	input.Bucket = sample.bucketName
	input.Key = sample.objectKey
	input.ACL = obs.AclPublicRead

	_, err := sample.obsClient.SetObjectAcl(input)
	if err != nil {
		panic(err)
	}
	fmt.Println("Set object acl successfully!")
	fmt.Println()

	output, err := sample.obsClient.GetObjectAcl(&obs.GetObjectAclInput{Bucket: sample.bucketName, Key: sample.objectKey})
	if err != nil {
		panic(err)
	}
	fmt.Printf("Object owner - ownerId:%s, ownerName:%s\n", output.Owner.ID, output.Owner.DisplayName)
	for index, grant := range output.Grants {
		fmt.Printf("Grant[%d]\n", index)
		fmt.Printf("GranteeUri:%s, GranteeId:%s, GranteeName:%s\n", grant.Grantee.URI, grant.Grantee.ID, grant.Grantee.DisplayName)
		fmt.Printf("Permission:%s\n", grant.Permission)
	}
}

func (sample ObjectOperationsSample) DeleteObject() {
	input := &obs.DeleteObjectInput{}
	input.Bucket = sample.bucketName
	input.Key = sample.objectKey

	_, err := sample.obsClient.DeleteObject(input)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Delete object:%s successfully!\n", input.Key)
	fmt.Println()

	input.Key = sample.objectKey + "-back"

	_, err = sample.obsClient.DeleteObject(input)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Delete object:%s successfully!\n", input.Key)
	fmt.Println()
}

func RunObjectOperationsSample() {
	const (
		endpoint   = "https://your-endpoint"
		ak         = "*** Provide your Access Key ***"
		sk         = "*** Provide your Secret Key ***"
		bucketName = "bucket-test"
		objectKey  = "object-test"
		location   = "yourbucketlocation"
	)

	sample := newObjectOperationsSample(ak, sk, endpoint, bucketName, objectKey, location)

	fmt.Println("Create a new bucket for demo")
	sample.CreateBucket()

	sample.CreateObject()

	sample.GetObjectMeta()

	sample.GetObject()

	sample.CopyObject()

	sample.DoObjectAcl()

	sample.DeleteObject()
}
