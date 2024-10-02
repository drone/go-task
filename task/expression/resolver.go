package expression

import (
	"encoding/json"

	"github.com/drone/go-task/task/common"
	"github.com/drone/go-task/task/expression/evaler"
)

type Resolver struct {
	secrets []*common.Secret
}

func New(secrets []*common.Secret) *Resolver {
	return &Resolver{secrets: secrets}
}

func (r *Resolver) Resolve(data []byte) ([]byte, error) {
	v := map[string]any{}

	// unmarshal the task data into a map
	err := json.Unmarshal(data, &v)
	if err != nil {
		return nil, err
	}

	// evaluate the expressions
	evaler.Eval(v, r.secrets)

	// encode the map back to []byte
	resolved, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return resolved, nil
}
