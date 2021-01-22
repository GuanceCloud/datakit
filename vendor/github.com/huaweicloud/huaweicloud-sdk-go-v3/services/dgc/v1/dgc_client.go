package v1

import (
	http_client "github.com/huaweicloud/huaweicloud-sdk-go-v3/core"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/dgc/v1/model"
)

type DgcClient struct {
	HcClient *http_client.HcHttpClient
}

func NewDgcClient(hcClient *http_client.HcHttpClient) *DgcClient {
	return &DgcClient{HcClient: hcClient}
}

func DgcClientBuilder() *http_client.HcHttpClientBuilder {
	builder := http_client.NewHcHttpClientBuilder()
	return builder
}

//
func (c *DgcClient) CancelScript(request *model.CancelScriptRequest) (*model.CancelScriptResponse, error) {
	requestDef := GenReqDefForCancelScript()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CancelScriptResponse), nil
	}
}

//
func (c *DgcClient) CreateConnection(request *model.CreateConnectionRequest) (*model.CreateConnectionResponse, error) {
	requestDef := GenReqDefForCreateConnection()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateConnectionResponse), nil
	}
}

//
func (c *DgcClient) CreateJob(request *model.CreateJobRequest) (*model.CreateJobResponse, error) {
	requestDef := GenReqDefForCreateJob()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateJobResponse), nil
	}
}

//
func (c *DgcClient) CreateResource(request *model.CreateResourceRequest) (*model.CreateResourceResponse, error) {
	requestDef := GenReqDefForCreateResource()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateResourceResponse), nil
	}
}

//
func (c *DgcClient) CreateScript(request *model.CreateScriptRequest) (*model.CreateScriptResponse, error) {
	requestDef := GenReqDefForCreateScript()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateScriptResponse), nil
	}
}

//
func (c *DgcClient) DeleteConnction(request *model.DeleteConnctionRequest) (*model.DeleteConnctionResponse, error) {
	requestDef := GenReqDefForDeleteConnction()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteConnctionResponse), nil
	}
}

//
func (c *DgcClient) DeleteJob(request *model.DeleteJobRequest) (*model.DeleteJobResponse, error) {
	requestDef := GenReqDefForDeleteJob()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteJobResponse), nil
	}
}

//
func (c *DgcClient) DeleteResource(request *model.DeleteResourceRequest) (*model.DeleteResourceResponse, error) {
	requestDef := GenReqDefForDeleteResource()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteResourceResponse), nil
	}
}

//
func (c *DgcClient) DeleteScript(request *model.DeleteScriptRequest) (*model.DeleteScriptResponse, error) {
	requestDef := GenReqDefForDeleteScript()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteScriptResponse), nil
	}
}

//
func (c *DgcClient) ExecuteScript(request *model.ExecuteScriptRequest) (*model.ExecuteScriptResponse, error) {
	requestDef := GenReqDefForExecuteScript()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ExecuteScriptResponse), nil
	}
}

//
func (c *DgcClient) ExportConnections(request *model.ExportConnectionsRequest) (*model.ExportConnectionsResponse, error) {
	requestDef := GenReqDefForExportConnections()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ExportConnectionsResponse), nil
	}
}

//
func (c *DgcClient) ExportJob(request *model.ExportJobRequest) (*model.ExportJobResponse, error) {
	requestDef := GenReqDefForExportJob()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ExportJobResponse), nil
	}
}

//
func (c *DgcClient) ExportJobList(request *model.ExportJobListRequest) (*model.ExportJobListResponse, error) {
	requestDef := GenReqDefForExportJobList()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ExportJobListResponse), nil
	}
}

//
func (c *DgcClient) ImportConnections(request *model.ImportConnectionsRequest) (*model.ImportConnectionsResponse, error) {
	requestDef := GenReqDefForImportConnections()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ImportConnectionsResponse), nil
	}
}

//
func (c *DgcClient) ImportJob(request *model.ImportJobRequest) (*model.ImportJobResponse, error) {
	requestDef := GenReqDefForImportJob()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ImportJobResponse), nil
	}
}

//
func (c *DgcClient) ListConnections(request *model.ListConnectionsRequest) (*model.ListConnectionsResponse, error) {
	requestDef := GenReqDefForListConnections()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListConnectionsResponse), nil
	}
}

//
func (c *DgcClient) ListJobInstances(request *model.ListJobInstancesRequest) (*model.ListJobInstancesResponse, error) {
	requestDef := GenReqDefForListJobInstances()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListJobInstancesResponse), nil
	}
}

//
func (c *DgcClient) ListJobs(request *model.ListJobsRequest) (*model.ListJobsResponse, error) {
	requestDef := GenReqDefForListJobs()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListJobsResponse), nil
	}
}

//
func (c *DgcClient) ListResources(request *model.ListResourcesRequest) (*model.ListResourcesResponse, error) {
	requestDef := GenReqDefForListResources()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListResourcesResponse), nil
	}
}

//
func (c *DgcClient) ListScriptResults(request *model.ListScriptResultsRequest) (*model.ListScriptResultsResponse, error) {
	requestDef := GenReqDefForListScriptResults()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListScriptResultsResponse), nil
	}
}

//
func (c *DgcClient) ListScripts(request *model.ListScriptsRequest) (*model.ListScriptsResponse, error) {
	requestDef := GenReqDefForListScripts()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListScriptsResponse), nil
	}
}

//
func (c *DgcClient) ListSystemTasks(request *model.ListSystemTasksRequest) (*model.ListSystemTasksResponse, error) {
	requestDef := GenReqDefForListSystemTasks()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListSystemTasksResponse), nil
	}
}

//
func (c *DgcClient) ModifyJob(request *model.ModifyJobRequest) (*model.ModifyJobResponse, error) {
	requestDef := GenReqDefForModifyJob()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ModifyJobResponse), nil
	}
}

//
func (c *DgcClient) ModifyResource(request *model.ModifyResourceRequest) (*model.ModifyResourceResponse, error) {
	requestDef := GenReqDefForModifyResource()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ModifyResourceResponse), nil
	}
}

//
func (c *DgcClient) ModifyScript(request *model.ModifyScriptRequest) (*model.ModifyScriptResponse, error) {
	requestDef := GenReqDefForModifyScript()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ModifyScriptResponse), nil
	}
}

//
func (c *DgcClient) RestoreJobInstance(request *model.RestoreJobInstanceRequest) (*model.RestoreJobInstanceResponse, error) {
	requestDef := GenReqDefForRestoreJobInstance()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.RestoreJobInstanceResponse), nil
	}
}

//
func (c *DgcClient) RunOnce(request *model.RunOnceRequest) (*model.RunOnceResponse, error) {
	requestDef := GenReqDefForRunOnce()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.RunOnceResponse), nil
	}
}

//
func (c *DgcClient) ShowConnection(request *model.ShowConnectionRequest) (*model.ShowConnectionResponse, error) {
	requestDef := GenReqDefForShowConnection()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowConnectionResponse), nil
	}
}

//
func (c *DgcClient) ShowFileInfo(request *model.ShowFileInfoRequest) (*model.ShowFileInfoResponse, error) {
	requestDef := GenReqDefForShowFileInfo()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowFileInfoResponse), nil
	}
}

//
func (c *DgcClient) ShowJob(request *model.ShowJobRequest) (*model.ShowJobResponse, error) {
	requestDef := GenReqDefForShowJob()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowJobResponse), nil
	}
}

//
func (c *DgcClient) ShowJobInstance(request *model.ShowJobInstanceRequest) (*model.ShowJobInstanceResponse, error) {
	requestDef := GenReqDefForShowJobInstance()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowJobInstanceResponse), nil
	}
}

//
func (c *DgcClient) ShowJobStatus(request *model.ShowJobStatusRequest) (*model.ShowJobStatusResponse, error) {
	requestDef := GenReqDefForShowJobStatus()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowJobStatusResponse), nil
	}
}

//
func (c *DgcClient) ShowResource(request *model.ShowResourceRequest) (*model.ShowResourceResponse, error) {
	requestDef := GenReqDefForShowResource()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowResourceResponse), nil
	}
}

//
func (c *DgcClient) ShowScript(request *model.ShowScriptRequest) (*model.ShowScriptResponse, error) {
	requestDef := GenReqDefForShowScript()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowScriptResponse), nil
	}
}

//
func (c *DgcClient) StartJob(request *model.StartJobRequest) (*model.StartJobResponse, error) {
	requestDef := GenReqDefForStartJob()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.StartJobResponse), nil
	}
}

//
func (c *DgcClient) StopJob(request *model.StopJobRequest) (*model.StopJobResponse, error) {
	requestDef := GenReqDefForStopJob()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.StopJobResponse), nil
	}
}

//
func (c *DgcClient) StopJobInstance(request *model.StopJobInstanceRequest) (*model.StopJobInstanceResponse, error) {
	requestDef := GenReqDefForStopJobInstance()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.StopJobInstanceResponse), nil
	}
}

//
func (c *DgcClient) UpdateConnection(request *model.UpdateConnectionRequest) (*model.UpdateConnectionResponse, error) {
	requestDef := GenReqDefForUpdateConnection()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdateConnectionResponse), nil
	}
}
