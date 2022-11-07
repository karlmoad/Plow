package secrets

import (
	"errors"
	"fmt"
	"os"
)

type EnvironmentSecretsConfiguration struct {
	Prefix string `mapstructure:"namespace"`
}

type EnvironmentSecretStore struct {
	config EnvironmentSecretsConfiguration
}

func (env *EnvironmentSecretStore) GetSecret(key string) (string, error) {
	// temp
	if val, ok := os.LookupEnv(fmt.Sprintf("%s-%s", env.config.Prefix, key)); ok {
		return val, nil
	} else {
		return "", errors.New("invalid secret key, key not found")
	}
}
