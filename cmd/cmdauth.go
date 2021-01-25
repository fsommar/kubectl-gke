package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/fsommar/kubectl-gke/pkg"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

const authCmdName = "auth"

func NewAuthCommand(streams genericclioptions.IOStreams) *cobra.Command {
	cmd := cobra.Command{
		Use: authCmdName,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(cmd.Context(), 5*time.Second)
			defer cancel()

			auth, err := pkg.GetGcpCredentials(ctx)
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
