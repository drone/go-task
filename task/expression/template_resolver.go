package expression

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

const (
	RESOLVER_DELIM_START = "<{"
	RESOLVER_DELIM_END   = "}>"
)

// Resolve processes template expressions in the input data, evaluating nested expressions from innermost to outermost.
// It supports nested expressions like: <{ outer <{ inner | getAsBase64 }> | getAsBase64 }>
//
// Example:
//
//	Input:  "hello <{ world | getAsBase64 }>"
//	Output: "hello d29ybGQ="
//
//	Input:  "<{ <{ inner | getAsBase64 }> | getAsBase64 }>"
//	Output: "aVc1dVpYST0=" (base64 of "aW5uZXI=", which is base64 of "inner")
func ResolveWithTemplateFunctions(data []byte) ([]byte, error) {
	input := string(data)
	maxIterations := 100

	for i := 0; i < maxIterations; i++ {
		start, end, found := findInnermostExpression(input)
		if !found {
			break
		}

		expr := input[start:end]
		preprocessed := preprocessExpression(expr)

		result, err := evaluateSingleExpression(preprocessed)
		if err != nil {
			return nil, fmt.Errorf("error evaluating expression %q: %w", expr, err)
		}

		input = input[:start] + result + input[end:]
	}

	return []byte(input), nil
}

// findInnermostExpression locates the deepest nested template expression in the input string.
// It uses depth tracking to find expressions that contain no other nested expressions.
//
// Example:
//
//	Input:  "<{ outer <{ inner }> }>"
//	Returns: start=9, end=21, found=true (pointing to "<{ inner }>")
//
//	Input:  "no expressions here"
//	Returns: start=0, end=0, found=false
func findInnermostExpression(input string) (start, end int, found bool) {
	depth := 0
	var innermostStart, innermostEnd int
	maxDepth := 0

	i := 0
	for i < len(input) {
		if i+len(RESOLVER_DELIM_START) <= len(input) && input[i:i+len(RESOLVER_DELIM_START)] == RESOLVER_DELIM_START {
			depth++
			if depth > maxDepth {
				maxDepth = depth
				innermostStart = i
			}
			i += len(RESOLVER_DELIM_START)
		} else if i+len(RESOLVER_DELIM_END) <= len(input) && input[i:i+len(RESOLVER_DELIM_END)] == RESOLVER_DELIM_END {
			if depth > 0 {
				if depth == maxDepth {
					innermostEnd = i + len(RESOLVER_DELIM_END)
					return innermostStart, innermostEnd, true
				}
				depth--
			}
			i += len(RESOLVER_DELIM_END)
		} else {
			i++
		}
	}

	return 0, 0, false
}

// preprocessExpression wraps unquoted content in quotes for getAsBase64 expressions.
// This is necessary because Go's template engine requires string literals to be quoted.
// Without this preprocessing, expressions like <{ something | getAsBase64 }> would fail
// because "something" would be interpreted as a command, not a string.
//
// Example:
//
//	Input:  "<{ something | getAsBase64 }>"
//	Output: "<{ \"something\" | getAsBase64 }>"
//
//	Input:  "<{ \"already-quoted\" | getAsBase64 }>"
//	Output: "<{ \"already-quoted\" | getAsBase64 }>" (unchanged)
func preprocessExpression(expr string) string {
	if !strings.Contains(expr, "getAsBase64") {
		return expr
	}

	content := expr[len(RESOLVER_DELIM_START) : len(expr)-len(RESOLVER_DELIM_END)]
	parts := strings.Split(content, "|")
	if len(parts) == 0 {
		return expr
	}

	quotedStr := strings.TrimSpace(parts[0])
	var unescaped string

	if strings.HasPrefix(quotedStr, "\\\"") && strings.HasSuffix(quotedStr, "\\\"") {
		unescaped = "\"" + quotedStr[2:len(quotedStr)-2] + "\""
	} else if strings.HasPrefix(quotedStr, "\"") && strings.HasSuffix(quotedStr, "\"") {
		unescaped = quotedStr
	} else {
		unescaped = "\"" + quotedStr + "\""
	}

	expr = RESOLVER_DELIM_START + " " + unescaped
	if len(parts) > 1 {
		expr += " |" + parts[1]
	}
	expr += RESOLVER_DELIM_END

	return expr
}

// evaluateSingleExpression evaluates a single template expression using Go's text/template engine.
// It parses the expression with custom delimiters and executes it.
//
// Example:
//
//	Input:  "<{ \"hello\" | getAsBase64 }>", secretsMap={}
//	Output: "aGVsbG8="
func evaluateSingleExpression(expr string) (string, error) {
	tmpl, err := template.New("resolver").
		Delims(RESOLVER_DELIM_START, RESOLVER_DELIM_END).
		Funcs(TemplateFunctions()).
		Parse(expr)

	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, make(map[string]interface{}))
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
