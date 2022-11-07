package secrets

import (
	"errors"
	"github.com/mitchellh/mapstructure"
	"strings"
)

var (
	ErrInvalidKeyStoreType = errors.New("invalid keystore type")
)

type SecretStore interface {
	GetSecret(key string) (string, error)
}

func InitKeyVault(kvtype string, config map[string]interface{}) (SecretStore, error) {
	switch strings.TrimSpace(strings.ToUpper(kvtype)) {
	//TODO replace with keyvault solution when ready
	case "KEYVAULT", "ENV":
		{
			var envConfig EnvironmentSecretsConfiguration
			err := mapstructure.Decode(config, &envConfig)
			if err != nil {
				return nil, err
			}

			return &EnvironmentSecretStore{config: envConfig}, nil
		}
	default:
		{
			return nil, ErrInvalidKeyStoreType
		}
	}
}
