package expression

import (
	"bytes"
	"strings"
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

	// Extract and process template expressions
	input := string(data)
	start := 0
	processed := ""

	for {
		// Find start of template expression
		startIdx := strings.Index(input[start:], RESOLVER_DELIM_START)
		if startIdx == -1 {
			// No more template expressions found
			processed += input[start:]
			break
		}
		startIdx += start

		// Find end of template expression
		endIdx := strings.Index(input[startIdx:], RESOLVER_DELIM_END)
		if endIdx == -1 {
			// No matching end delimiter found
			processed += input[start:]
			break
		}
		endIdx += startIdx + len(RESOLVER_DELIM_END)

		// Add everything before the template expression
		processed += input[start:startIdx]

		// Get the template expression content
		expr := input[startIdx:endIdx]
		if strings.Contains(expr, "getAsBase64") {
			// Extract content between delimiters
			content := expr[len(RESOLVER_DELIM_START) : len(expr)-len(RESOLVER_DELIM_END)]
			// Split by pipe operator
			parts := strings.Split(content, "|")
			if len(parts) > 0 {
				// Get the quoted string part and trim spaces
				quotedStr := strings.TrimSpace(parts[0])
				
				// Handle the quoted string
				var unescaped string

                    // handl case use input contains quotes i.e. "abcd"
					if strings.HasPrefix(quotedStr, "\\\"") && strings.HasSuffix(quotedStr, "\\\"") {
						// Remove only first and last escaped quotes
						unescaped = "\"" + quotedStr[2:len(quotedStr)-2] + "\""
					} else if strings.HasPrefix(quotedStr, "\"") && strings.HasSuffix(quotedStr, "\"") {
						unescaped = quotedStr
					}else{
						unescaped = "\"" + quotedStr + "\""
					}
				
				// Reconstruct the expression
				expr = RESOLVER_DELIM_START + " " + unescaped
				if len(parts) > 1 {
					expr += " |" + parts[1]
				}
				expr += RESOLVER_DELIM_END
			}
		}
		processed += expr

		// Move start to after this template expression
		start = endIdx
	}

	// Parse the template with custom delimiters
	tmpl, err := template.New("resolver").
		Delims(RESOLVER_DELIM_START, RESOLVER_DELIM_END).
		Funcs(TemplateFunctions()). // Include template functions
		Parse(processed)

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
