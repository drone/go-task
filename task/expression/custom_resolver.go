package expression

import (
	"bytes"
	"encoding/json"

	"github.com/drone/go-task/task/common"
	"github.com/drone/go-task/task/expression/evaler"
)


/**
 * CustomResolver is a struct that implements the ResolverIfc interface.
 * It is used to resolve CEL-like expressions for secrets in task data.
 * It's using a custom resolver to evaluate expressions in the task data.
 * The resolver replaces the expressions with the actual secret values.
 * The expressions are expected to be in the format of "${{secrets.<secret_id>}}".
 * The secret values are provided in the form of a slice of common.Secret.
 */
type CustomResolver struct {
	secrets []*common.Secret
}

func newCustomResolver(secrets []*common.Secret) *CustomResolver {
	return &CustomResolver{secrets: secrets}
}

func (r *CustomResolver) Resolve(data []byte) ([]byte, error) {
	v := map[string]any{}

	// unmarshal the task data into a map
	err := json.Unmarshal(data, &v)
	if err != nil {
		return nil, err
	}

	// evaluate the expressions
	evaler.Eval(v, r.secrets)

	// encode the map back to []byte using a custom encoder that doesn't escape HTML
	buf := &bytes.Buffer{}
	encoder := json.NewEncoder(buf)
	encoder.SetEscapeHTML(false)
	err = encoder.Encode(v)
	if err != nil {
		return nil, err
	}

	// trim the trailing newline that Encode adds
	resolved := bytes.TrimSpace(buf.Bytes())
	return resolved, nil
}
