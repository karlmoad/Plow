package plow

import (
	"Plow/plow/objects"
	"Plow/plow/secrets"
	"Plow/plow/targets"
	"Plow/plow/targets/common"
	"context"
	"errors"
)

type Operation struct {
	config  Configuration
	options objects.Options
	target  common.Target
	repo    *Repo
}

func NewOperation(config Configuration, options objects.Options) (*Operation, error) {
	operation := &Operation{config: config, options: options}

	//unpack config and init secret store and target
	//secret store
	secretStr, err := secrets.InitKeyVault(config.SecretStoreType, config.SecretStore)
	if err != nil {
		return nil, err
	}
	target, err := targets.NewTarget(config.TargetType, config.Target, &operation.options, secretStr)
	if err != nil {
		return nil, err
	}
	operation.target = target

	var repo *Repo
	//init repo instance
	if !options.OptionFlags.Has(objects.UseLocalRepositorySetting) {
		repo, err = newMemoryRepo(&operation.config, &operation.options, secretStr)
	} else {
		repo, err = newLocalRepo(&operation.config, &operation.options, secretStr)
	}

	if err != nil {
		return nil, err
	}
	if repo == nil {
		return nil, errors.New("unable to obtain repository reference")
	}

	operation.repo = repo

	return operation, nil
}

func (o *Operation) Repository() *Repo {
	return o.repo
}

func (o *Operation) GenerateChangeLog() (*objects.ChangeLog, error) {
	if o.options.IsFileProvided() {
		changes := objects.NewChangeLog(o.target.GetObjectTypeTranslator())
		bundle := changes.AddManualBundle()
		err := bundle.AddItem(o.options.File.Bytes, objects.NewChangeMetaFromOptions(&o.options))
		if err != nil {
			return nil, err
		}
		return changes, nil
	} else {
		return o.listRepositoryChanges()
	}
}

func (o *Operation) listRepositoryChanges() (*objects.ChangeLog, error) {
	history, err := o.target.GetTrackingHistory(0)
	if err != nil {
		return nil, err
	}
	return o.repo.BuildChangeLog(history, o.target.GetObjectTypeTranslator())
}

func (o *Operation) ValidateChanges(changes *objects.ChangeLog) error {
	if o.options.OptionFlags.Has(objects.SkipValidationSetting) {
		return errors.New("cannot validate changes, skip validation option was set")
	}

	err := o.target.ValidateChangeLog(changes)
	if err != nil {
		return err
	}
	return nil
}

func (o *Operation) RenderChanges(changes *objects.ChangeLog) ([]*common.RenderedChange, error) {
	if !o.options.OptionFlags.Has(objects.SkipValidationSetting) {
		err := o.ValidateChanges(changes)
		if err != nil {
			return nil, err
		}
	}
	return o.target.RenderChangeLog(changes)
}

func (o *Operation) ApplyChanges(context context.Context, changes *objects.ChangeLog) error {
	if !o.options.OptionFlags.Has(objects.SkipValidationSetting) {
		err := o.ValidateChanges(changes)
		if err != nil {
			return err
		}
	}
	return o.target.ApplyChangeLog(context, changes)
}

func (o *Operation) GetExecutionOrder() []int64 {
	return o.target.GetObjectTypeExecutionOrder()
}
