package expression

import (
	"bytes"
	"text/template"

	"github.com/drone/go-task/task/common"
)

const (
	RESOLVER_DELIM_START = "<{"
	RESOLVER_DELIM_END   = "}>"
)

type TemplateResolver struct {
	secrets []*common.Secret
}

func newTemplateResolver(secrets []*common.Secret) *TemplateResolver {
	return &TemplateResolver{secrets: secrets}
}

func (r *TemplateResolver) Resolve(data []byte) ([]byte, error) {
	// Create a map of secrets for easy lookup by ID
	secretMap := make(map[string]string)
	for _, secret := range r.secrets {
		secretMap[secret.ID] = secret.Value
	}

	// Parse the template with custom delimiters
	tmpl, err := template.New("resolver").
		Delims(RESOLVER_DELIM_START, RESOLVER_DELIM_END).
		Funcs(TemplateFunctions()). // Include template functions
		Parse(string(data))
	if err != nil {
		return nil, err
	}

	// Execute the template with the secret map as context
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, common.SecretsToMap(r.secrets))
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
