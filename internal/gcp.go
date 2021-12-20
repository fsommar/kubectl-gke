package internal

import (
	"context"
	"fmt"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/container/v1"
	"google.golang.org/api/option"
)

func GetGoogleCloudClusters(ctx context.Context, project, location string) ([]*container.Cluster, error) {
	client, err := google.DefaultClient(ctx, container.CloudPlatformScope)
	if err != nil {
		return nil, err
	}

	containerService, err := container.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}

	parent := fmt.Sprintf("projects/%s/locations/%s", project, location)

	resp, err := containerService.Projects.Locations.Clusters.List(parent).Context(ctx).Do()
	if err != nil {
		return nil, err
	}

	return resp.Clusters, nil
}

func GetGoogleCloudCluster(ctx context.Context, project, location, cluster string) (*container.Cluster, error) {
	client, err := google.DefaultClient(ctx, container.CloudPlatformScope)
	if err != nil {
		return nil, err
	}

	containerService, err := container.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}

	name := fmt.Sprintf("projects/%s/locations/%s/clusters/%s", project, location, cluster)

	return containerService.Projects.Locations.Clusters.Get(name).Context(ctx).Do()
}
