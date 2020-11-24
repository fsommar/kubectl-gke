package cmd

import (
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func NewGetCredentialsCommand(streams genericclioptions.IOStreams) *cobra.Command {
	cmd := cobra.Command{
		Use: "get-credentials",
	}

	cmd.InitDefaultHelpFlag()
	cmd.AddCommand(NewGcpCommand(streams))

	return &cmd
}
