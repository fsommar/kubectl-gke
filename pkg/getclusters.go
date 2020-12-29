package pkg

import (
	"context"
	"encoding/base64"
	"fmt"

	"google.golang.org/api/container/v1"

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
		clusters = append(clusters, into(cluster))
	}
	return clusters, nil
}

func GetCluster(
	ctx context.Context,
	project, location, name string,
) (*Cluster, error) {
	gkeCluster, err := internal.GetGoogleCloudCluster(ctx, project, location, name)
	if err != nil {
		return nil, err
	}
	cluster := into(gkeCluster)
	return &cluster, nil
}

func into(cluster *container.Cluster) Cluster {
	ca, _ := base64.StdEncoding.DecodeString(cluster.MasterAuth.ClusterCaCertificate)
	return Cluster{
		Name:                     cluster.Name,
		Server:                   fmt.Sprintf("https://%s", cluster.Endpoint),
		Location:                 cluster.Location,
		CertificateAuthorityData: ca,
	}
}
