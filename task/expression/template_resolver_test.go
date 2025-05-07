package expression

import (
	"testing"

	"github.com/drone/go-task/task/common"
	"github.com/stretchr/testify/assert"
)

func TestTemplateResolver_Resolve(t *testing.T) {
	// Arrange
	secrets := []*common.Secret{
		{ID: "abc", Value: "111"},
		{ID: "xyz", Value: "222"},
	}
	resolver := newTemplateResolver(secrets)
	input := []byte("my secret value: #{{ .secrets.abc}} and another: #{{.secrets.xyz}}")
	expected := "my secret value: 111 and another: 222"

	// Act
	output, err := resolver.Resolve(input)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expected, string(output))
}

func TestTemplateResolver_Resolve_MissingSecret(t *testing.T) {
	// Arrange
	secrets := []*common.Secret{
		{ID: "abc", Value: "111"},
	}
	resolver := newTemplateResolver(secrets)
	input := []byte("my secret value: #{{ .secrets.missing }}")
	expected := "my secret value: <no value>"

	// Act
	output, err := resolver.Resolve(input)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expected, string(output))
}

func TestTemplateResolver_Resolve_InvalidTemplate(t *testing.T) {
	// Arrange
	secrets := []*common.Secret{
		{ID: "abc", Value: "111"},
	}
	resolver := newTemplateResolver(secrets)
	input := []byte("my secret value: #{{ .secrets.abc")

	// Act
	output, err := resolver.Resolve(input)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, output)
}

func TestTemplateResolver_Resolve_WithGetAsBase64(t *testing.T) {
	// Arrange
	secrets := []*common.Secret{
		{ID: "abc", Value: "mySecretValue"},
	}
	resolver := newTemplateResolver(secrets)
	input1 := []byte("my secret value in base64: #{{.secrets.abc | getAsBase64 }}")
	expected1 := "my secret value in base64: bXlTZWNyZXRWYWx1ZQ==" // Base64 encoding of "mySecretValue"

	// Act
	output1, err := resolver.Resolve(input1)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expected1, string(output1))
}
