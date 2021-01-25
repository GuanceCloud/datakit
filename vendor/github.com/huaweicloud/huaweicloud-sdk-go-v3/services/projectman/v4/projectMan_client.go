package v4

import (
	http_client "github.com/huaweicloud/huaweicloud-sdk-go-v3/core"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/projectman/v4/model"
)

type ProjectManClient struct {
	HcClient *http_client.HcHttpClient
}

func NewProjectManClient(hcClient *http_client.HcHttpClient) *ProjectManClient {
	return &ProjectManClient{HcClient: hcClient}
}

func ProjectManClientBuilder() *http_client.HcHttpClientBuilder {
	builder := http_client.NewHcHttpClientBuilder()
	return builder
}

//AGC调用 当前用户申请加入项目, 申请的用户id写在header中
func (c *ProjectManClient) AddApplyJoinProjectForAgc(request *model.AddApplyJoinProjectForAgcRequest) (*model.AddApplyJoinProjectForAgcResponse, error) {
	requestDef := GenReqDefForAddApplyJoinProjectForAgc()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.AddApplyJoinProjectForAgcResponse), nil
	}
}

//添加项目成员,可以添加跨租户成员
func (c *ProjectManClient) AddMemberV4(request *model.AddMemberV4Request) (*model.AddMemberV4Response, error) {
	requestDef := GenReqDefForAddMemberV4()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.AddMemberV4Response), nil
	}
}

//批量添加项目成员，只能添加和项目创建者同一租户下的成员，不正确的用户id会略过，添加的用户超过权限的，默认角色设置为7
func (c *ProjectManClient) BatchAddMembersV4(request *model.BatchAddMembersV4Request) (*model.BatchAddMembersV4Response, error) {
	requestDef := GenReqDefForBatchAddMembersV4()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.BatchAddMembersV4Response), nil
	}
}

//批量删除项目的迭代
func (c *ProjectManClient) BatchDeleteIterationsV4(request *model.BatchDeleteIterationsV4Request) (*model.BatchDeleteIterationsV4Response, error) {
	requestDef := GenReqDefForBatchDeleteIterationsV4()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.BatchDeleteIterationsV4Response), nil
	}
}

//批量删除项目成员
func (c *ProjectManClient) BatchDeleteMembersV4(request *model.BatchDeleteMembersV4Request) (*model.BatchDeleteMembersV4Response, error) {
	requestDef := GenReqDefForBatchDeleteMembersV4()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.BatchDeleteMembersV4Response), nil
	}
}

//更新项目
func (c *ProjectManClient) CheckProjectNameV4(request *model.CheckProjectNameV4Request) (*model.CheckProjectNameV4Response, error) {
	requestDef := GenReqDefForCheckProjectNameV4()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CheckProjectNameV4Response), nil
	}
}

//创建Scrum项目迭代
func (c *ProjectManClient) CreateIterationV4(request *model.CreateIterationV4Request) (*model.CreateIterationV4Response, error) {
	requestDef := GenReqDefForCreateIterationV4()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateIterationV4Response), nil
	}
}

//创建项目
func (c *ProjectManClient) CreateProjectV4(request *model.CreateProjectV4Request) (*model.CreateProjectV4Response, error) {
	requestDef := GenReqDefForCreateProjectV4()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateProjectV4Response), nil
	}
}

//删除项目迭代
func (c *ProjectManClient) DeleteIterationV4(request *model.DeleteIterationV4Request) (*model.DeleteIterationV4Response, error) {
	requestDef := GenReqDefForDeleteIterationV4()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteIterationV4Response), nil
	}
}

//删除项目
func (c *ProjectManClient) DeleteProjectV4(request *model.DeleteProjectV4Request) (*model.DeleteProjectV4Response, error) {
	requestDef := GenReqDefForDeleteProjectV4()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteProjectV4Response), nil
	}
}

//获取租户没有加入的项目
func (c *ProjectManClient) ListDomainNotAddedProjectsV4(request *model.ListDomainNotAddedProjectsV4Request) (*model.ListDomainNotAddedProjectsV4Response, error) {
	requestDef := GenReqDefForListDomainNotAddedProjectsV4()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListDomainNotAddedProjectsV4Response), nil
	}
}

//获取bug统计信息，按模块统计
func (c *ProjectManClient) ListProjectBugStaticsV4(request *model.ListProjectBugStaticsV4Request) (*model.ListProjectBugStaticsV4Response, error) {
	requestDef := GenReqDefForListProjectBugStaticsV4()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListProjectBugStaticsV4Response), nil
	}
}

//获取需求统计信息
func (c *ProjectManClient) ListProjectDemandStaticV4(request *model.ListProjectDemandStaticV4Request) (*model.ListProjectDemandStaticV4Response, error) {
	requestDef := GenReqDefForListProjectDemandStaticV4()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListProjectDemandStaticV4Response), nil
	}
}

//获取项目迭代
func (c *ProjectManClient) ListProjectIterationsV4(request *model.ListProjectIterationsV4Request) (*model.ListProjectIterationsV4Response, error) {
	requestDef := GenReqDefForListProjectIterationsV4()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListProjectIterationsV4Response), nil
	}
}

//获取项目成员列表
func (c *ProjectManClient) ListProjectMembersV4(request *model.ListProjectMembersV4Request) (*model.ListProjectMembersV4Response, error) {
	requestDef := GenReqDefForListProjectMembersV4()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListProjectMembersV4Response), nil
	}
}

//查询项目列表
func (c *ProjectManClient) ListProjectsV4(request *model.ListProjectsV4Request) (*model.ListProjectsV4Response, error) {
	requestDef := GenReqDefForListProjectsV4()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListProjectsV4Response), nil
	}
}

//项目成员主动退出项目，项目创建者不能退出
func (c *ProjectManClient) RemoveProject(request *model.RemoveProjectRequest) (*model.RemoveProjectResponse, error) {
	requestDef := GenReqDefForRemoveProject()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.RemoveProjectResponse), nil
	}
}

//获取当前用户信息
func (c *ProjectManClient) ShowCurUserInfo(request *model.ShowCurUserInfoRequest) (*model.ShowCurUserInfoResponse, error) {
	requestDef := GenReqDefForShowCurUserInfo()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowCurUserInfoResponse), nil
	}
}

//获取用户在项目中的角色
func (c *ProjectManClient) ShowCurUserRole(request *model.ShowCurUserRoleRequest) (*model.ShowCurUserRoleResponse, error) {
	requestDef := GenReqDefForShowCurUserRole()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowCurUserRoleResponse), nil
	}
}

//查看迭代详情
func (c *ProjectManClient) ShowIterationV4(request *model.ShowIterationV4Request) (*model.ShowIterationV4Response, error) {
	requestDef := GenReqDefForShowIterationV4()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowIterationV4Response), nil
	}
}

//获取项目概览
func (c *ProjectManClient) ShowProjectSummaryV4(request *model.ShowProjectSummaryV4Request) (*model.ShowProjectSummaryV4Response, error) {
	requestDef := GenReqDefForShowProjectSummaryV4()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowProjectSummaryV4Response), nil
	}
}

//更新Scrum项目迭代
func (c *ProjectManClient) UpdateIterationV4(request *model.UpdateIterationV4Request) (*model.UpdateIterationV4Response, error) {
	requestDef := GenReqDefForUpdateIterationV4()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdateIterationV4Response), nil
	}
}

//更新成员在项目中的角色
func (c *ProjectManClient) UpdateMembesRoleV4(request *model.UpdateMembesRoleV4Request) (*model.UpdateMembesRoleV4Response, error) {
	requestDef := GenReqDefForUpdateMembesRoleV4()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdateMembesRoleV4Response), nil
	}
}

//更新用户昵称
func (c *ProjectManClient) UpdateNickNameV4(request *model.UpdateNickNameV4Request) (*model.UpdateNickNameV4Response, error) {
	requestDef := GenReqDefForUpdateNickNameV4()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdateNickNameV4Response), nil
	}
}

//更新项目
func (c *ProjectManClient) UpdateProjectV4(request *model.UpdateProjectV4Request) (*model.UpdateProjectV4Response, error) {
	requestDef := GenReqDefForUpdateProjectV4()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdateProjectV4Response), nil
	}
}

//批量删除工作项
func (c *ProjectManClient) BatchDeleteIssuesV4(request *model.BatchDeleteIssuesV4Request) (*model.BatchDeleteIssuesV4Response, error) {
	requestDef := GenReqDefForBatchDeleteIssuesV4()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.BatchDeleteIssuesV4Response), nil
	}
}

//创建工作项
func (c *ProjectManClient) CreateIssueV4(request *model.CreateIssueV4Request) (*model.CreateIssueV4Response, error) {
	requestDef := GenReqDefForCreateIssueV4()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateIssueV4Response), nil
	}
}

//删除工作项
func (c *ProjectManClient) DeleteIssueV4(request *model.DeleteIssueV4Request) (*model.DeleteIssueV4Response, error) {
	requestDef := GenReqDefForDeleteIssueV4()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteIssueV4Response), nil
	}
}

//获取子工作项
func (c *ProjectManClient) ListChildIssuesV4(request *model.ListChildIssuesV4Request) (*model.ListChildIssuesV4Response, error) {
	requestDef := GenReqDefForListChildIssuesV4()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListChildIssuesV4Response), nil
	}
}

//获取工作项的评论
func (c *ProjectManClient) ListIssueCommentsV4(request *model.ListIssueCommentsV4Request) (*model.ListIssueCommentsV4Response, error) {
	requestDef := GenReqDefForListIssueCommentsV4()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListIssueCommentsV4Response), nil
	}
}

//获取项目成员列表
func (c *ProjectManClient) ListIssueRecordsV4(request *model.ListIssueRecordsV4Request) (*model.ListIssueRecordsV4Response, error) {
	requestDef := GenReqDefForListIssueRecordsV4()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListIssueRecordsV4Response), nil
	}
}

//高级查询工作项,根据筛选条件查询工作中
func (c *ProjectManClient) ListIssuesV4(request *model.ListIssuesV4Request) (*model.ListIssuesV4Response, error) {
	requestDef := GenReqDefForListIssuesV4()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListIssuesV4Response), nil
	}
}

//按用户查询工时（多项目）
func (c *ProjectManClient) ListProjectWorkHours(request *model.ListProjectWorkHoursRequest) (*model.ListProjectWorkHoursResponse, error) {
	requestDef := GenReqDefForListProjectWorkHours()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListProjectWorkHoursResponse), nil
	}
}

//查询工作项详情
func (c *ProjectManClient) ShowIssueV4(request *model.ShowIssueV4Request) (*model.ShowIssueV4Response, error) {
	requestDef := GenReqDefForShowIssueV4()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowIssueV4Response), nil
	}
}

//按用户查询工时（单项目）
func (c *ProjectManClient) ShowProjectWorkHours(request *model.ShowProjectWorkHoursRequest) (*model.ShowProjectWorkHoursResponse, error) {
	requestDef := GenReqDefForShowProjectWorkHours()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowProjectWorkHoursResponse), nil
	}
}

//获取工作项的完成率
func (c *ProjectManClient) ShowtIssueCompletionRate(request *model.ShowtIssueCompletionRateRequest) (*model.ShowtIssueCompletionRateResponse, error) {
	requestDef := GenReqDefForShowtIssueCompletionRate()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowtIssueCompletionRateResponse), nil
	}
}

//更新工作项
func (c *ProjectManClient) UpdateIssueV4(request *model.UpdateIssueV4Request) (*model.UpdateIssueV4Response, error) {
	requestDef := GenReqDefForUpdateIssueV4()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdateIssueV4Response), nil
	}
}
