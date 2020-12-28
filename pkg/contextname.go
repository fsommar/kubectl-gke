package pkg

import (
	"os"
	"strings"
	"text/template"
)

// GcloudDefaultFormat is the default context format used by the gcloud get-credentials command.
const GcloudDefaultFormat = "gke_{{ .Project }}_{{ .Cluster.Location }}_{{ .Cluster.Name }}"

var gcloudDefaultFactory ContextNameFactory

func init() {
	gcloudDefaultFactory, _ = NewContextNameFactory(GcloudDefaultFormat)
}

type ContextNameFactory interface {
	// For returns the context name for the combined project and cluster properties.
	For(string, Cluster) (string, error)
}

type contextNameFactory struct {
	t      *template.Template
	format string
}

type data struct {
	Project string
	Cluster Cluster
}

func (c contextNameFactory) For(project string, cluster Cluster) (string, error) {
	contextName := strings.Builder{}
	if err := c.t.Execute(&contextName, &data{
		Project: project,
		Cluster: cluster,
	}); err != nil {
		return "", err
	}
	return contextName.String(), nil
}

var _ ContextNameFactory = (*contextNameFactory)(nil)

// NewContextNameFactory returns a new ContextNameFactory, preferring the format string in `$KUBECTL_GKE_CONTEXT_FORMAT`
// over the provided string.
func NewContextNameFactory(fallback string) (ContextNameFactory, error) {
	var format string
	if env, exists := os.LookupEnv("KUBECTL_GKE_CONTEXT_FORMAT"); exists {
		format = env
	} else if fallback == "" {
		format = GcloudDefaultFormat
	} else {
		format = fallback
	}
	t, err := template.New("context-name").Parse(format)
	if err != nil {
		return nil, err
	}
	return contextNameFactory{t: t}, nil
}

// DefaultContextName returns the default context name used by the gcloud get-credentials command.
func DefaultContextName(project string, cluster Cluster) (string, error) {
	return gcloudDefaultFactory.For(project, cluster)
}
