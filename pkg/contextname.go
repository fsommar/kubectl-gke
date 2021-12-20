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

type TemplateFactory struct {
	t      *template.Template
	Format string
}

type data struct {
	Project string
	Cluster Cluster
}

func (c TemplateFactory) For(project string, cluster Cluster) (string, error) {
	contextName := strings.Builder{}
	if err := c.t.Execute(&contextName, &data{
		Project: project,
		Cluster: cluster,
	}); err != nil {
		return "", err
	}

	return contextName.String(), nil
}

var _ ContextNameFactory = (*TemplateFactory)(nil)

// NewContextNameFactory returns a new ContextNameFactory, preferring the format string in `$KUBECTL_GKE_CONTEXT_FORMAT`
// over the provided string.
func NewContextNameFactory(fallbackFormat string) (*TemplateFactory, error) {
	var format string
	if env, exists := os.LookupEnv("KUBECTL_GKE_CONTEXT_FORMAT"); exists {
		format = env
	} else if fallbackFormat == "" {
		format = GcloudDefaultFormat
	} else {
		format = fallbackFormat
	}

	t, err := template.New("context-name").Parse(format)
	if err != nil {
		return nil, err
	}

	return &TemplateFactory{t: t, Format: format}, nil
}

// DefaultContextName returns the default context name used by the gcloud get-credentials command.
func DefaultContextName(project string, cluster Cluster) (string, error) {
	return gcloudDefaultFactory.For(project, cluster)
}
