package main

import (
	"github.com/go-logr/logr"
	"k8s.io/klog/v2"
	"os"

	"github.com/fsommar/kubectl-gke/cmd"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

const version = "v0.2.0"

func main() {
	klog.SetLogger(logr.Discard())

	getCredentialsCmd := cmd.NewGkeCommand(genericclioptions.IOStreams{
		In:     os.Stdin,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	})
	getCredentialsCmd.Version = version
	if err := getCredentialsCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
