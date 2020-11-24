package cmd

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"os"
	"strings"
	"time"

	"github.com/fsommar/kubectl-get_credentials/internal"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

type gcpCommand struct {
	streams         genericclioptions.IOStreams
	project         string
	location        string
	createContexts  bool
	contextTemplate string
	labelSelector   string
}

func NewGcpCommand(streams genericclioptions.IOStreams) *cobra.Command {
	gcp := gcpCommand{streams: streams}
	cmd := cobra.Command{
		Use:          "gcp",
		SilenceUsage: true,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if gcp.project != "" {
				return nil
			}
			ctx, cancel := context.WithTimeout(cmd.Context(), 100*time.Millisecond)
			defer cancel()
			if project, err := internal.GetGcpProject(ctx); err == nil && project != "" {
				gcp.project = project
				_, _ = fmt.Fprintf(streams.ErrOut, "using %s as project from application default", project)
				return nil
			}
			_ = cmd.Help()
			return errors.New("expected project to be readable from application default, or provided as a flag")
		},
		RunE: gcp.runE,
	}
	cmd.InitDefaultHelpFlag()
	cmd.Flags().StringVarP(&gcp.project, "project", "p", "", "GCP project (required)")
	cmd.Flags().StringVar(&gcp.location, "location", "-", "GCP location like region or zone")
	cmd.Flags().BoolVarP(&gcp.createContexts, "create-contexts", "x", false, "Creates contexts and a user; when this is false only clusters will be created/modified (which should be non-intrusive)")
	cmd.Flags().StringVarP(&gcp.contextTemplate, "template", "t", "gke_{{ .Project }}_{{ .Cluster.Location }}_{{ .Cluster.Name }}", "Context name template")
	cmd.Flags().StringVarP(&gcp.labelSelector, "selector", "l", "", "Selector (label query) to filter on, supports '=', '==', and '!='. (e.g. -l key1=value1,key2=value2)")

	cmd.AddCommand(NewGcpCredentialsCommand(streams))

	return &cmd
}

func (g *gcpCommand) runE(cmd *cobra.Command, _ []string) error {
	ctx, cancel := context.WithTimeout(cmd.Context(), 5*time.Second)
	defer cancel()

	t, err := template.New("context-template").Parse(g.contextTemplate)
	if err != nil {
		return err
	}

	clusters, err := internal.GetGoogleCloudClusters(ctx, g.project, g.location)
	if err != nil {
		return err
	}

	options := clientcmd.NewDefaultPathOptions()
	config, err := options.GetStartingConfig()
	if err != nil {
		return err
	}

	selector := labels.Everything()
	if g.labelSelector != "" {
		selector, err = labels.Parse(g.labelSelector)
		if err != nil {
			return err
		}
	}

	for _, cluster := range clusters {
		if !selector.Matches(labels.Set(cluster.ResourceLabels)) {
			continue
		}

		ca, _ := base64.StdEncoding.DecodeString(cluster.MasterAuth.ClusterCaCertificate)

		// reuse gcloud's context and cluster format to maintain some compatibility
		clusterName := fmt.Sprintf("gke_%s_%s_%s", g.project, cluster.Location, cluster.Name)
		configCluster, exists := config.Clusters[clusterName]
		if !exists {
			configCluster = clientcmdapi.NewCluster()
		}
		c := *configCluster
		c.CertificateAuthorityData = ca
		c.Server = fmt.Sprintf("https://%s", cluster.Endpoint)
		config.Clusters[clusterName] = &c

		if g.createContexts {
			type Cluster struct {
				Name     string
				Location string
			}
			type Data struct {
				Project string
				Cluster Cluster
			}
			contextBldr := strings.Builder{}
			if err := t.Execute(&contextBldr, &Data{
				Project: g.project,
				Cluster: Cluster{Name: cluster.Name, Location: cluster.Location},
			}); err != nil {
				return err
			}
			contextName := contextBldr.String()

			configContext, exists := config.Contexts[contextName]
			if !exists {
				configContext = clientcmdapi.NewContext()
			}
			c := *configContext
			c.Cluster = clusterName
			c.AuthInfo = "gcp"
			config.Contexts[contextName] = &c
		}
	}

	if g.createContexts {
		if err := createUser(ctx, config, "gcp"); err != nil {
			return err
		}
	}

	return clientcmd.ModifyConfig(options, *config, true)
}

func NewGcpCredentialsCommand(streams genericclioptions.IOStreams) *cobra.Command {
	cmd := cobra.Command{
		Use: "auth",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(cmd.Context(), 5*time.Second)
			defer cancel()

			auth, err := internal.GetGcpCredentials(ctx)
			if err != nil {
				return err
			}
			b, err := json.Marshal(auth)
			if err != nil {
				return err
			}
			_, err = fmt.Fprintln(streams.Out, string(b))
			return err
		}}
	return &cmd
}

// createUser mutates the config adding or modifying a user with the provided name.
//
// The user will use this command's auth mechanism for kubectl, to get the authentication token and its expiry time.
func createUser(_ context.Context, config *clientcmdapi.Config, user string) error {
	path, err := os.Executable()
	if err != nil {
		return err
	}

	authInfo, exists := config.AuthInfos[user]
	if !exists {
		authInfo = clientcmdapi.NewAuthInfo()
	}
	a := *authInfo
	a.AuthProvider = &clientcmdapi.AuthProviderConfig{
		Name: "gcp",
		Config: map[string]string{
			"cmd-args":   "gcp auth",
			"cmd-path":   path,
			"expiry-key": "{.token_expiry}",
			"token-key":  "{.access_token}",
		},
	}
	config.AuthInfos[user] = &a
	return nil
}
