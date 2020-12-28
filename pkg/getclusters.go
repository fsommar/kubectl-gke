package pkg

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/fsommar/kubectl-gke/internal"
	"k8s.io/apimachinery/pkg/labels"
)

type Cluster struct {
	Name                     string `json:"name"`
	Server                   string `json:"server"`
	Location                 string `json:"location"`
	CertificateAuthorityData []byte `json:"certificateAuthorityData"`
}

func GetClusters(
	ctx context.Context,
	selector labels.Selector,
	project, location string,
) ([]Cluster, error) {
	gkeClusters, err := internal.GetGoogleCloudClusters(ctx, project, location)
	if err != nil {
		return nil, err
	}

	clusters := make([]Cluster, 0)
	for _, cluster := range gkeClusters {
		if !selector.Matches(labels.Set(cluster.ResourceLabels)) {
			continue
		}

		ca, _ := base64.StdEncoding.DecodeString(cluster.MasterAuth.ClusterCaCertificate)
		c := Cluster{
			Name:                     cluster.Name,
			Server:                   fmt.Sprintf("https://%s", cluster.Endpoint),
			Location:                 cluster.Location,
			CertificateAuthorityData: ca,
		}
		clusters = append(clusters, c)
	}
	return clusters, nil
}
