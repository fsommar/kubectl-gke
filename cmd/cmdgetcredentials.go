package cmd

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/fsommar/kubectl-gke/pkg"
	"github.com/fsommar/kubectl-gke/pkg/config"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/tools/clientcmd"
)

var labelSelector string

type getCredentialsCommand struct {
	streams          genericclioptions.IOStreams
	project          string
	location         string
	noCreateContexts bool
	contextTemplate  string
	selector         labels.Selector
}

func NewGetCredentialsCommand(streams genericclioptions.IOStreams) *cobra.Command {
	gcp := getCredentialsCommand{streams: streams}
	cmd := cobra.Command{
		Use:          "get-credentials PROJECT",
		SilenceUsage: true,
		Args:         cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) (err error) {
			gcp.project = args[0]
			if labelSelector != "" {
				gcp.selector, err = labels.Parse(labelSelector)
			} else {
				gcp.selector = labels.Everything()
			}
			return
		},
		RunE: gcp.runE,
	}
	cmd.InitDefaultHelpFlag()
	cmd.Flags().StringVar(&gcp.location, "location", "-", "GCP location like region or zone")
	cmd.Flags().BoolVar(&gcp.noCreateContexts, "no-create-contexts", false, "When this is true, no contexts and users will be created, only clusters (which should be non-intrusive)")
	cmd.Flags().StringVarP(&gcp.contextTemplate, "format", "f", pkg.GcloudDefaultFormat, "Format of the context name")
	cmd.Flags().StringVarP(&labelSelector, "selector", "l", "", "Selector (label query) to filter on, supports '=', '==', and '!='. (e.g. -l key1=value1,key2=value2)")

	return &cmd
}

func (g *getCredentialsCommand) runE(cmd *cobra.Command, _ []string) error {
	ctx, cancel := context.WithTimeout(cmd.Context(), 5*time.Second)
	defer cancel()

	contextNameFactory, err := pkg.NewContextNameFactory(g.contextTemplate)
	if err != nil {
		return err
	}

	options := clientcmd.NewDefaultPathOptions()
	cfg, err := options.GetStartingConfig()
	if err != nil {
		return err
	}

	clusters, err := pkg.GetClusters(ctx, g.selector, g.project, g.location)
	if err != nil {
		return err
	}

	for _, cluster := range clusters {
		// Maintain compatibility with gcloud's default cluster names.
		clusterName, _ := pkg.DefaultContextName(g.project, cluster)
		config.UpsertCluster(cfg, clusterName, cluster)
		if g.shouldCreateContexts() {
			contextName, err := contextNameFactory.For(g.project, cluster)
			if err != nil {
				return err
			}
			config.UpsertContext(cfg, contextName, clusterName, "kubectl-gke")
		}
	}

	if g.shouldCreateContexts() {
		path, err := os.Executable()
		if err != nil {
			return err
		}
		authCmd, _, err := cmd.Root().Find([]string{authCmdName})
		if err != nil {
			return err
		}
		// CommandPath includes name of the root command, which is not used in the invocation.
		args := strings.Replace(authCmd.CommandPath(), cmd.Root().Name()+" ", "", 1)

		config.UpsertUser(cfg, "kubectl-gke", map[string]string{
			"cmd-path": path,
			"cmd-args": args,
		})
	}

	return clientcmd.ModifyConfig(options, *cfg, true)
}

func (g getCredentialsCommand) shouldCreateContexts() bool {
	return !g.noCreateContexts
}
