package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/pkg/apis/clientauthentication/v1"
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

			exp := metav1.NewTime(auth.ExpiryTime)
			cred := v1.ExecCredential{
				TypeMeta: metav1.TypeMeta{
					APIVersion: v1.SchemeGroupVersion.String(),
					Kind:       "ExecCredential",
				},
				Status: &v1.ExecCredentialStatus{
					ExpirationTimestamp: &exp,
					Token:               auth.AccessToken,
				},
			}

			b, err := json.Marshal(cred)
			if err != nil {
				return err
			}
			_, err = fmt.Fprintln(streams.Out, string(b))
			return err
		}}

	return &cmd
}
