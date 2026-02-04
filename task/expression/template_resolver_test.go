package expression

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTemplateResolver_Resolve(t *testing.T) {
	input := []byte("my secret value: <{.secrets.abc}> and another: <{.secrets.xyz}>")
	expected := "my secret value: <no value> and another: <no value>"

	// Act
	output, _, err := ResolveWithTemplateFunctions(input, []string{})

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expected, string(output))
}

func TestTemplateResolver_Resolve_WithGetAsBase64(t *testing.T) {
	input1 := []byte("my secret value in base64: <{ \"username: mySecretValue\" | getAsBase64 }>")
	expected1 := "my secret value in base64: dXNlcm5hbWU6IG15U2VjcmV0VmFsdWU=" // gitleaks:allow

	// Act
	output1, _, err := ResolveWithTemplateFunctions(input1, []string{})

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expected1, string(output1))
}

func TestTemplateResolver_Resolve_NestedGetAsBase64(t *testing.T) {
	// Test simple nested expression: encode "something" to base64, then encode "hello <base64>" to base64
	input := []byte("ohmy <{ hello <{ something | getAsBase64 }> | getAsBase64 }>")
	// Inner: "something" -> "c29tZXRoaW5n"
	// Outer: "hello c29tZXRoaW5n" -> "aGVsbG8gYzI5dFpYUm9hVzVu"
	expected := "ohmy aGVsbG8gYzI5dFpYUm9hVzVu"

	// Act
	output, _, err := ResolveWithTemplateFunctions(input, []string{})

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expected, string(output))
}

func TestTemplateResolver_Resolve_DeeplyNestedGetAsBase64(t *testing.T) {
	// Test 3-level nested expression
	input := []byte("<{ outer <{ middle <{ inner | getAsBase64 }> | getAsBase64 }> | getAsBase64 }>")
	// Level 1 (innermost): "inner" -> "aW5uZXI="
	// Level 2: "middle aW5uZXI=" -> "bWlkZGxlIGFXNXVaWEk9"
	// Level 3 (outermost): "outer bWlkZGxlIGFXNXVaWEk9" -> "b3V0ZXIgYldsa1pHeGxJR0ZYTlhWYVdFazk="
	expected := "b3V0ZXIgYldsa1pHeGxJR0ZYTlhWYVdFazk="

	// Act
	output, _, err := ResolveWithTemplateFunctions(input, []string{})

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expected, string(output))
}

func TestTemplateResolver_Resolve_MultipleNestedExpressions(t *testing.T) {
	// Test multiple nested expressions in same input
	input := []byte("first: <{ a <{ b | getAsBase64 }> | getAsBase64 }> and second: <{ x <{ y | getAsBase64 }> | getAsBase64 }>")
	// First nested: "b" -> "Yg==", then "a Yg==" -> "YSBZZz09"
	// Second nested: "y" -> "eQ==", then "x eQ==" -> "eCBlUT09"
	expected := "first: YSBZZz09 and second: eCBlUT09"

	// Act
	output, _, err := ResolveWithTemplateFunctions(input, []string{})

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expected, string(output))
}

func TestTemplateResolver_DerivedSecretMasking_SimpleBase64(t *testing.T) {
	// Test that when an expression contains a secret, the base64-encoded result is returned as a mask
	secretValue := "mySecretPassword123"
	input := []byte("<{ username: " + secretValue + " | getAsBase64 }>")
	// Expected base64 of "username: mySecretPassword123"
	expectedOutput := "dXNlcm5hbWU6IG15U2VjcmV0UGFzc3dvcmQxMjM="

	// Act
	output, masks, err := ResolveWithTemplateFunctions(input, []string{secretValue})

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedOutput, string(output))
	assert.Len(t, masks, 1, "Should return 1 derived mask")
	assert.Equal(t, expectedOutput, masks[0], "The base64-encoded result should be in the masks list")
}

func TestTemplateResolver_DerivedSecretMasking_NestedExpressions(t *testing.T) {
	// Test that nested expressions containing secrets return multiple derived masks
	secretValue := "secret123"
	input := []byte("result: <{ outer <{ inner " + secretValue + " | getAsBase64 }> | getAsBase64 }>")

	// Inner: "inner secret123" -> base64 = "aW5uZXIgc2VjcmV0MTIz"
	innerBase64 := "aW5uZXIgc2VjcmV0MTIz"
	// Outer: "outer aW5uZXIgc2VjcmV0MTIz" -> base64 = "b3V0ZXIgYVc1dVpYSWdjMlZqY21WME1USXo="
	outerBase64 := "b3V0ZXIgYVc1dVpYSWdjMlZqY21WME1USXo="
	expectedOutput := "result: " + outerBase64

	// Act
	output, masks, err := ResolveWithTemplateFunctions(input, []string{secretValue})

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedOutput, string(output))
	assert.Len(t, masks, 2, "Should return 2 derived masks (inner and outer)")
	assert.Contains(t, masks, innerBase64, "Inner base64 result should be in masks")
	assert.Contains(t, masks, outerBase64, "Outer base64 result should be in masks")
}
