package expression

import (
	"bytes"

	"github.com/drone/go-task/task/common"
)

type ResolverIfc interface {
	Resolve(data []byte) ([]byte, error)
}

type Resolver struct {
	secrets []*common.Secret
}

func New(secrets []*common.Secret) *Resolver {
	return &Resolver{secrets: secrets}
}

func (r *Resolver) Resolve(taskData []byte) ([]byte, error) {
	// Start with the original task data
	currentData := taskData

	// First pass: Handle custom secrets syntax (${{secrets...}})
	if bytes.Contains(currentData, []byte("${{secrets")) {
		customResolver := newCustomResolver(r.secrets)
		resolvedData, err := customResolver.Resolve(currentData)
		if err != nil {
			return nil, err
		}
		currentData = resolvedData // Update current data with resolved result
	}

	// Second pass: Handle template resolver syntax
	templateResolver := newTemplateResolver(r.secrets)
	finalResolvedData, err := templateResolver.Resolve(currentData)
	if err != nil {
		return nil, err
	}

	return finalResolvedData, nil
}
