package internal

import (
	"context"
	"errors"
	"fmt"
	"time"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/container/v1"
	"google.golang.org/api/option"
)

func GetGoogleCloudClusters(ctx context.Context, project string, location string) ([]*container.Cluster, error) {
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

type Credentials struct {
	AccessToken string `json:"access_token"`
	ExpiryTime  string `json:"token_expiry"`
}

func GetGcpCredentials(ctx context.Context) (*Credentials, error) {
	creds, err := google.FindDefaultCredentials(ctx, container.CloudPlatformScope)
	if err != nil {
		return nil, err
	}
	if token, err := creds.TokenSource.Token(); err == nil && token.Valid() {
		return &Credentials{
			AccessToken: token.AccessToken,
			ExpiryTime:  token.Expiry.UTC().Format(time.RFC3339),
		}, nil
	}
	return nil, errors.New("unable to get credentials")
}
