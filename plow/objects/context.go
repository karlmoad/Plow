package objects

import (
	"Plow/plow/secrets"
	"context"
)

type PlowContext struct {
	Config      Configuration
	Options     Options
	SecretStore secrets.SecretStore
	context     context.Context
}

func NewPlowContext(context context.Context, config Configuration, options Options) (*PlowContext, error) {
	secretStr, err := secrets.InitKeyVault(config.SecretStoreType, config.SecretStore)
	if err != nil {
		return nil, err
	}
	return &PlowContext{context: context, Config: config, SecretStore: secretStr, Options: options}, nil
}
