package plow

import (
	"Plow/plow/objects"
	"Plow/plow/targets"
	"Plow/plow/targets/common"
	"context"
	"errors"
)

type Operation struct {
	context *objects.PlowContext
	target  common.Target
	repo    *Repo
}

func NewOperation(context *objects.PlowContext) (*Operation, error) {
	operation := &Operation{context: context}

	target, err := targets.NewTarget(context.Config.TargetType, context)
	if err != nil {
		return nil, err
	}
	operation.target = target

	var repo *Repo
	//init repo instance
	if !operation.context.Options.OptionFlags.Has(objects.UseLocalRepositorySetting) {
		repo, err = newMemoryRepo(operation.context)
	} else {
		repo, err = newLocalRepo(operation.context)
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
	if o.context.Options.IsFileProvided() {
		changes := objects.NewChangeLog(o.target.GetObjectTypeTranslator())
		bundle := changes.AddManualBundle()
		err := bundle.AddItem(o.context.Options.File.Bytes, objects.NewChangeMetaFromOptions(&o.context.Options))
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
	if o.context.Options.OptionFlags.Has(objects.SkipValidationSetting) {
		return errors.New("cannot validate changes, skip validation option was set")
	}

	err := o.target.ValidateChangeLog(changes)
	if err != nil {
		return err
	}
	return nil
}

func (o *Operation) RenderChanges(changes *objects.ChangeLog) ([]*common.RenderedChange, error) {
	if !o.context.Options.OptionFlags.Has(objects.SkipValidationSetting) {
		err := o.ValidateChanges(changes)
		if err != nil {
			return nil, err
		}
	}
	return o.target.RenderChangeLog(changes)
}

func (o *Operation) ApplyChanges(context context.Context, changes *objects.ChangeLog) error {
	if !o.context.Options.OptionFlags.Has(objects.SkipValidationSetting) {
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
