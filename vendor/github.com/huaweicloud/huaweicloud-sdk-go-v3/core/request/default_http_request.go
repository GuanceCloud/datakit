// Copyright 2020 Huawei Technologies Co.,Ltd.
//
// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package request

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"
)

type DefaultHttpRequest struct {
	endpoint             string
	path                 string
	method               string
	queryParams          map[string]interface{}
	pathParams           map[string]string
	autoFilledPathParams map[string]string
	headerParams         map[string]string
	body                 interface{}
}

func (httpRequest *DefaultHttpRequest) fillParamsInPath() *DefaultHttpRequest {
	for key, value := range httpRequest.pathParams {
		httpRequest.path = strings.ReplaceAll(httpRequest.path, "{"+key+"}", value)
	}
	for key, value := range httpRequest.autoFilledPathParams {
		httpRequest.path = strings.ReplaceAll(httpRequest.path, "{"+key+"}", value)
	}
	return httpRequest
}

func (httpRequest *DefaultHttpRequest) Builder() *HttpRequestBuilder {
	httpRequestBuilder := HttpRequestBuilder{httpRequest: httpRequest}
	return &httpRequestBuilder
}

func (httpRequest *DefaultHttpRequest) GetEndpoint() string {
	return httpRequest.endpoint
}

func (httpRequest *DefaultHttpRequest) GetPath() string {
	return httpRequest.path
}

func (httpRequest *DefaultHttpRequest) GetMethod() string {
	return httpRequest.method
}

func (httpRequest *DefaultHttpRequest) GetQueryParams() map[string]interface{} {
	return httpRequest.queryParams
}

func (httpRequest *DefaultHttpRequest) GetHeaderParams() map[string]string {
	return httpRequest.headerParams
}

func (httpRequest *DefaultHttpRequest) GetPathPrams() map[string]string {
	return httpRequest.pathParams
}

func (httpRequest *DefaultHttpRequest) GetBody() interface{} {
	return httpRequest.body
}

func (httpRequest *DefaultHttpRequest) GetBodyToBytes() (*bytes.Buffer, error) {
	buf := &bytes.Buffer{}

	if httpRequest.body != nil {
		err := json.NewEncoder(buf).Encode(httpRequest.body)
		if err != nil {
			return nil, err
		}
	}

	return buf, nil
}

func (httpRequest *DefaultHttpRequest) AddQueryParam(key string, value string) {
	httpRequest.queryParams[key] = value
}

func (httpRequest *DefaultHttpRequest) AddPathParam(key string, value string) {
	httpRequest.pathParams[key] = value
}

func (httpRequest *DefaultHttpRequest) AddHeaderParam(key string, value string) {
	httpRequest.headerParams[key] = value
}

func (httpRequest *DefaultHttpRequest) ConvertRequest() (*http.Request, error) {
	buf, err := httpRequest.GetBodyToBytes()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(httpRequest.GetMethod(), httpRequest.GetEndpoint(), buf)
	if err != nil {
		return nil, err
	}

	httpRequest.fillPath(req)
	httpRequest.fillQueryParams(req)
	httpRequest.fillHeaderParams(req)

	return req, nil
}

func (httpRequest *DefaultHttpRequest) fillHeaderParams(req *http.Request) {
	if len(httpRequest.GetHeaderParams()) == 0 {
		return
	}

	for key, value := range httpRequest.GetHeaderParams() {
		req.Header.Add(key, value)
	}
}

func (httpRequest *DefaultHttpRequest) fillQueryParams(req *http.Request) {
	if len(httpRequest.GetQueryParams()) == 0 {
		return
	}

	q := req.URL.Query()
	for key, value := range httpRequest.GetQueryParams() {
		valueWithType := value.(reflect.Value)

		if valueWithType.Kind() == reflect.Slice {
			params := httpRequest.CanonicalSliceQueryParamsToMulti(valueWithType)
			for _, param := range params {
				q.Add(key, param)
			}
		} else if valueWithType.Kind() == reflect.Map {
			params := httpRequest.CanonicalMapQueryParams(key, valueWithType)
			for _, param := range params {
				for k, v := range param {
					q.Add(k, v)
				}
			}
		} else {
			q.Add(key, httpRequest.CanonicalStringQueryParams(valueWithType))
		}
	}

	req.URL.RawQuery = strings.ReplaceAll(strings.Trim(q.Encode(), "="), "=&", "&")
}

func (httpRequest *DefaultHttpRequest) CanonicalStringQueryParams(value reflect.Value) string {
	return fmt.Sprintf("%v", value)
}

func (httpRequest *DefaultHttpRequest) CanonicalSliceQueryParamsToMulti(value reflect.Value) []string {
	params := make([]string, 0)
	for i := 0; i < value.Len(); i++ {
		if value.Index(i).Kind() == reflect.Struct {
			v, e := json.Marshal(value.Interface())
			if e == nil {
				if strings.HasPrefix(string(v), "\"") {
					params = append(params, strings.Trim(string(v), "\""))
				} else {
					params = append(params, string(v))
				}
			}
		} else {
			params = append(params, httpRequest.CanonicalStringQueryParams(value.Index(i)))
		}
	}
	return params
}

func (httpRequest *DefaultHttpRequest) CanonicalMapQueryParams(key string, value reflect.Value) []map[string]string {
	queryParams := make([]map[string]string, 0)

	for _, k := range value.MapKeys() {
		if value.MapIndex(k).Kind() == reflect.Struct {
			v, e := json.Marshal(value.Interface())
			if e == nil {
				if strings.HasPrefix(string(v), "\"") {
					queryParams = append(queryParams, map[string]string{
						key: strings.Trim(string(v), "\""),
					})
				} else {
					queryParams = append(queryParams, map[string]string{
						key: string(v),
					})
				}
			}
		} else if value.MapIndex(k).Kind() == reflect.Slice {
			params := httpRequest.CanonicalSliceQueryParamsToMulti(value.MapIndex(k))
			if len(params) == 0 {
				queryParams = append(queryParams, map[string]string{
					fmt.Sprintf("%s[%s]", key, k): "",
				})
				continue
			}
			for _, paramValue := range httpRequest.CanonicalSliceQueryParamsToMulti(value.MapIndex(k)) {
				queryParams = append(queryParams, map[string]string{
					fmt.Sprintf("%s[%s]", key, k): paramValue,
				})
			}
		} else if value.MapIndex(k).Kind() == reflect.Map {
			queryParams = append(queryParams, httpRequest.CanonicalMapQueryParams(fmt.Sprintf("%s[%s]", key, k), value.MapIndex(k))...)
		} else {
			queryParams = append(queryParams, map[string]string{
				fmt.Sprintf("%s[%s]", key, k): httpRequest.CanonicalStringQueryParams(value.MapIndex(k)),
			})
		}
	}

	return queryParams
}

func (httpRequest *DefaultHttpRequest) fillPath(req *http.Request) {
	if "" != httpRequest.GetPath() {
		req.URL.Path = httpRequest.GetPath()
	}
}

type HttpRequestBuilder struct {
	httpRequest *DefaultHttpRequest
}

func NewHttpRequestBuilder() *HttpRequestBuilder {
	httpRequest := &DefaultHttpRequest{
		queryParams:          make(map[string]interface{}),
		headerParams:         make(map[string]string),
		pathParams:           make(map[string]string),
		autoFilledPathParams: make(map[string]string),
	}
	httpRequestBuilder := &HttpRequestBuilder{
		httpRequest: httpRequest,
	}
	return httpRequestBuilder
}

func (builder *HttpRequestBuilder) WithEndpoint(endpoint string) *HttpRequestBuilder {
	builder.httpRequest.endpoint = endpoint
	return builder
}

func (builder *HttpRequestBuilder) WithPath(path string) *HttpRequestBuilder {
	builder.httpRequest.path = path
	return builder
}

func (builder *HttpRequestBuilder) WithMethod(method string) *HttpRequestBuilder {
	builder.httpRequest.method = method
	return builder
}

func (builder *HttpRequestBuilder) AddQueryParam(key string, value interface{}) *HttpRequestBuilder {
	builder.httpRequest.queryParams[key] = value
	return builder
}

func (builder *HttpRequestBuilder) AddPathParam(key string, value string) *HttpRequestBuilder {
	builder.httpRequest.pathParams[key] = value
	return builder
}

func (builder *HttpRequestBuilder) AddAutoFilledPathParam(key string, value string) *HttpRequestBuilder {
	builder.httpRequest.autoFilledPathParams[key] = value
	return builder
}

func (builder *HttpRequestBuilder) AddHeaderParam(key string, value string) *HttpRequestBuilder {
	builder.httpRequest.headerParams[key] = value
	return builder
}

func (builder *HttpRequestBuilder) WithBody(body interface{}) *HttpRequestBuilder {
	builder.httpRequest.body = body
	return builder
}

func (builder *HttpRequestBuilder) Build() *DefaultHttpRequest {
	return builder.httpRequest.fillParamsInPath()
}
