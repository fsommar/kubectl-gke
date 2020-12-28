package config

import (
	"github.com/fsommar/kubectl-gke/pkg"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

func UpsertCluster(cfg *clientcmdapi.Config, clusterName string, cluster pkg.Cluster) {
	configCluster, exists := cfg.Clusters[clusterName]
	if !exists {
		configCluster = clientcmdapi.NewCluster()
	}
	c := *configCluster
	c.CertificateAuthorityData = cluster.CertificateAuthorityData
	c.Server = cluster.Server
	cfg.Clusters[clusterName] = &c
}

func UpsertContext(cfg *clientcmdapi.Config, contextName, clusterName, user string) {
	configContext, exists := cfg.Contexts[contextName]
	if !exists {
		configContext = clientcmdapi.NewContext()
	}
	c := *configContext
	c.Cluster = clusterName
	c.AuthInfo = user
	cfg.Contexts[contextName] = &c
}

func UpsertUser(cfg *clientcmdapi.Config, user, cmdPath, cmdArgs string) {
	authInfo, exists := cfg.AuthInfos[user]
	if !exists {
		authInfo = clientcmdapi.NewAuthInfo()
	}
	a := *authInfo
	// AuthProvider is a built-in KubeConfig concept where vendors provide in-tree authentication mechanisms. The `gcp`
	// auth provider is one of the existing in-tree ones, and it expects the cmd-path and cmd-args to produce a JSON
	// output with an access token (`.access_token`) and expiration time (`.expiry_time`).
	a.AuthProvider = &clientcmdapi.AuthProviderConfig{
		Name: "gcp",
		Config: map[string]string{
			"cmd-path": cmdPath,
			"cmd-args": cmdArgs,
		},
	}
	cfg.AuthInfos[user] = &a
}
