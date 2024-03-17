package delegate

import (
	"github.com/kelseyhightower/envconfig"
)

// Sample config
type Config struct {
	Debug bool `envconfig:"DRONE_DEBUG"`
	Trace bool `envconfig:"DRONE_TRACE"`

	Delegate struct {
		AccountID       string `envconfig:"DRONE_DELEGATE_ACCOUNT_ID"`
		AccountSecret   string `envconfig:"DRONE_DELEGATE_ACCOUNT_SECRET"`
		ManagerEndpoint string `envconfig:"DRONE_DELEGATE_MANAGER_ENDPOINT"`
		Name            string `envconfig:"DRONE_DELEGATE_NAME"`
	}
}

func FromEnviron() (Config, error) {
	var config Config
	err := envconfig.Process("", &config)
	if err != nil {
		return config, err
	}

	return config, nil
}
