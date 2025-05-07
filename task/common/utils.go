package common

// SecretsToMap converts an array of Secret to a map with ID as the key and Value as the value.
func SecretsToMap(secrets []*Secret) map[string]any {
	secretMap := make(map[string]string)
	for _, secret := range secrets {
		secretMap[secret.ID] = secret.Value
	}

	return map[string]any{
		"secrets": secretMap,
	}
}
