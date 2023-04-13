package pkg

import (
	"context"
	"errors"
	"time"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/container/v1"
)

type Credentials struct {
	AccessToken string    `json:"access_token"`
	ExpiryTime  time.Time `json:"token_expiry"`
}

func GetGcpCredentials(ctx context.Context) (*Credentials, error) {
	creds, err := google.FindDefaultCredentials(ctx, container.CloudPlatformScope)
	if err != nil {
		return nil, err
	}

	if token, err := creds.TokenSource.Token(); err == nil && token.Valid() {
		return &Credentials{
			AccessToken: token.AccessToken,
			ExpiryTime:  token.Expiry.UTC(),
		}, nil
	}

	return nil, errors.New("unable to get credentials, consider running `gcloud auth application-default login`")
}
