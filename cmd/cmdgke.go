package cmd

import (
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func NewGkeCommand(streams genericclioptions.IOStreams) *cobra.Command {
	cmd := cobra.Command{
		Use: "gke",
	}

	cmd.InitDefaultHelpFlag()
	cmd.AddCommand(NewGetCredentialsCommand(streams))
	cmd.AddCommand(NewAuthCommand(streams))

	return &cmd
}
