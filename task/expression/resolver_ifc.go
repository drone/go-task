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
	if bytes.Contains(taskData, []byte("${{secrets")) {
		resolver := newCustomResolver(r.secrets)
		resolvedTaskData, err := resolver.Resolve(taskData)
		if err != nil {
			return nil, err
		}
		return resolvedTaskData, nil
	}
	//return taskData, nil
	resolver := newTemplateResolver(r.secrets)
	resolvedTaskData, err := resolver.Resolve(taskData)
	if err != nil {
		return nil, err
	}
	return resolvedTaskData, nil
}
