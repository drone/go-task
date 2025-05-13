package expression

import (
	"encoding/base64"
	"text/template"
)

// TemplateFunctions provides custom functions for templates.
func TemplateFunctions() template.FuncMap {
	return template.FuncMap{
		"getAsBase64": resolveGetAsBase64,
		// Examples of additional template functions.
		// Be aware that putting more compute on Runner side is not a good idea. Pl. use template functions
		// only when it is really needed.
		// "toUpper":   strings.ToUpper,
		// "toLower":   strings.ToLower,
		// "trimSpace": strings.TrimSpace,
		// "contains":  strings.Contains,
		// "replace":   strings.Replace,
		// "split":     strings.Split,
		// "join":      strings.Join,
	}
}

func resolveGetAsBase64(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}
