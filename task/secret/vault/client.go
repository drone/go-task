package vault

import (
	vault "github.com/hashicorp/vault/api"
)

// Config provides the vault configuration.
type Config struct {
	Address   string  `json:"address"`
	Token     *string `json:"token"`
	Namespace *string `json:"namespace"`
}

// New returns a new vault client.
func New(in *Config) (*vault.Client, error) {
	config := vault.DefaultConfig()
	if in == nil {
		return vault.NewClient(config)

	}
	if in.Address != "" {
		config.Address = in.Address
	}
	client, err := vault.NewClient(config)
	if err != nil {
		return nil, err
	}
	if in.Token != nil {
		client.SetToken(*in.Token)
	}

	if in.Namespace != nil {
		client.SetNamespace(*in.Namespace)
	}
	return client, nil
}
