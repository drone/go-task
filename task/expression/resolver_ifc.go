package expression

import (
	"bytes"

	"github.com/drone/go-task/task/common"
)

type ResolverIfc interface {
	Resolve(data []byte) ([]byte, []string, error)
}

type Resolver struct {
	secrets []*common.Secret
}

func New(secrets []*common.Secret) *Resolver {
	return &Resolver{secrets: secrets}
}

func (r *Resolver) Resolve(taskData []byte) ([]byte, []string, error) {
	// Start with the original task data
	currentData := taskData

	// First pass: Handle custom secrets syntax (${{secrets...}})
	if bytes.Contains(currentData, []byte("${{secrets")) {
		customResolver := newCustomResolver(r.secrets)
		resolvedData, err := customResolver.Resolve(currentData)
		if err != nil {
			return nil, nil, err
		}
		currentData = resolvedData // Update current data with resolved result
	}

	// Second pass: Handle template resolver syntax
	// This is used solely to resolve the `<{ content | getAsBase64 }>` expressions
	finalResolvedData, additionalMasks, err := ResolveWithTemplateFunctions(currentData, r.extractSecretValues())
	if err != nil {
		return nil, nil, err
	}

	return finalResolvedData, additionalMasks, nil
}

func (r *Resolver) extractSecretValues() []string {
	secretValues := make([]string, 0, len(r.secrets))
	for _, secret := range r.secrets {
		if secret != nil && secret.Value != "" {
			secretValues = append(secretValues, secret.Value)
		}
	}
	return secretValues
}
