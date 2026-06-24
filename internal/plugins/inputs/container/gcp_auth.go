// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package container

import (
	"context"
	"fmt"
	"net/http"
	"os"

	gcpauth "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/cloudprovider/gcp/auth"
	gcpmetadata "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/cloudprovider/gcp/metadata"
)

func newGCPHTTPClient() *http.Client {
	return gcpauth.NewHTTPClient()
}

func newGCPMetadataClient() *gcpmetadata.Client {
	return gcpmetadata.NewClient()
}

func resolveGCPCloudConfig(ctx context.Context, ipt *Input, metadata *gcpmetadata.Client) error {
	cfg, err := metadata.Resolve(ctx, gcpmetadata.Config{
		ProjectID:       ipt.GCPProjectID,
		ClusterName:     ipt.GCPClusterName,
		ClusterLocation: ipt.GCPClusterLocation,
	})
	if err != nil {
		return err
	}

	ipt.GCPProjectID = cfg.ProjectID
	ipt.GCPClusterName = cfg.ClusterName
	ipt.GCPClusterLocation = cfg.ClusterLocation

	if clusterName := os.Getenv("ENV_CLUSTER_NAME_K8S"); clusterName == "" || clusterName == "default" {
		if err := os.Setenv("ENV_CLUSTER_NAME_K8S", ipt.GCPClusterName); err != nil {
			return fmt.Errorf("set discovered kubernetes cluster name: %w", err)
		}
	}
	return nil
}
