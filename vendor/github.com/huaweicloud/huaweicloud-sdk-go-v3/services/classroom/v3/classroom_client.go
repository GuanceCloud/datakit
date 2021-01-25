package v3

import (
	http_client "github.com/huaweicloud/huaweicloud-sdk-go-v3/core"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/classroom/v3/model"
)

type ClassroomClient struct {
	HcClient *http_client.HcHttpClient
}

func NewClassroomClient(hcClient *http_client.HcHttpClient) *ClassroomClient {
	return &ClassroomClient{HcClient: hcClient}
}

func ClassroomClientBuilder() *http_client.HcHttpClientBuilder {
	builder := http_client.NewHcHttpClientBuilder()
	return builder
}

//根据课堂ID获取指定课堂的课堂成员列表，支持分页，搜索字段默认同时匹配姓名，学号，用户名，班级。
func (c *ClassroomClient) ListClassroomMembers(request *model.ListClassroomMembersRequest) (*model.ListClassroomMembersResponse, error) {
	requestDef := GenReqDefForListClassroomMembers()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListClassroomMembersResponse), nil
	}
}

//获取当前用户的课堂列表，课堂课表分为我创建的课堂，我加入的课堂以及所有课堂，支持分页查询。
func (c *ClassroomClient) ListClassrooms(request *model.ListClassroomsRequest) (*model.ListClassroomsResponse, error) {
	requestDef := GenReqDefForListClassrooms()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListClassroomsResponse), nil
	}
}

//根据课堂ID获取指定课堂的详细信息
func (c *ClassroomClient) ShowClassroomDetail(request *model.ShowClassroomDetailRequest) (*model.ShowClassroomDetailResponse, error) {
	requestDef := GenReqDefForShowClassroomDetail()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowClassroomDetailResponse), nil
	}
}

//查询课堂下指定成员的作业信息
func (c *ClassroomClient) ListClassroomMemberJobs(request *model.ListClassroomMemberJobsRequest) (*model.ListClassroomMemberJobsResponse, error) {
	requestDef := GenReqDefForListClassroomMemberJobs()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListClassroomMemberJobsResponse), nil
	}
}

//查询指定课堂下的作业列表信息，支持分页查询。
func (c *ClassroomClient) ListJobs(request *model.ListJobsRequest) (*model.ListJobsResponse, error) {
	requestDef := GenReqDefForListJobs()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListJobsResponse), nil
	}
}

//查询学生指定作业的习题提交记录信息(针对函数习题)
func (c *ClassroomClient) ListMemberJobRecords(request *model.ListMemberJobRecordsRequest) (*model.ListMemberJobRecordsResponse, error) {
	requestDef := GenReqDefForListMemberJobRecords()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListMemberJobRecordsResponse), nil
	}
}

//根据作业ID，查询指定作业的信息
func (c *ClassroomClient) ShowJobDetail(request *model.ShowJobDetailRequest) (*model.ShowJobDetailResponse, error) {
	requestDef := GenReqDefForShowJobDetail()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowJobDetailResponse), nil
	}
}

//查询指定作业下的习题信息
func (c *ClassroomClient) ShowJobExercises(request *model.ShowJobExercisesRequest) (*model.ShowJobExercisesResponse, error) {
	requestDef := GenReqDefForShowJobExercises()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowJobExercisesResponse), nil
	}
}
