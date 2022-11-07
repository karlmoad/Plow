package secrets

import (
	"errors"
)

type KeyVaultConfiguration struct {
	Url string `mapstructure:"url"`
}

type KeyVault struct {
	config KeyVaultConfiguration
}

func (kv *KeyVault) GetSecret(key string) (string, error) {
	return "", errors.New("Not Implemented")
}
