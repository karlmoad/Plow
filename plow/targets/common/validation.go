package common

import (
	"Plow/plow/objects"
)

type Validator interface {
	Init() error
	Designation() string
	Validate(change *objects.ChangeItem) error
	Destroy() error
}

type ValidationHandler struct {
	typeValidators   map[int64][]Validator
	globalValidators []Validator
	typeMapper       func(string) int64
}

func (v *ValidationHandler) Initialize() error {
	//global validators init
	for _, validator := range v.globalValidators {
		err := validator.Init()
		if err != nil {
			return err
		}
	}

	//global validators init
	for _, validators := range v.typeValidators {
		for _, validator := range validators {
			err := validator.Init()
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (v *ValidationHandler) Close() error {
	//global validators destroy
	for _, validator := range v.globalValidators {
		err := validator.Destroy()
		if err != nil {
			return err
		}
	}

	//global validators destroy
	for _, validators := range v.typeValidators {
		for _, validator := range validators {
			err := validator.Destroy()
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (v *ValidationHandler) RegisterGlobalValidator(validator Validator) {
	v.globalValidators = append(v.globalValidators, validator)
}

func (v *ValidationHandler) RegisterTypeValidator(t int64, validator Validator) {
	if _, ok := v.typeValidators[t]; !ok {
		v.typeValidators[t] = make([]Validator, 0)
	}
	v.typeValidators[t] = append(v.typeValidators[t], validator)
}

func (v *ValidationHandler) Validate(change *objects.ChangeItem) error {

	//apply global validators in order
	for _, validator := range v.globalValidators {
		if err := validator.Validate(change); err != nil {
			return err
		}
	}

	//apply type validators in order
	if typeValidators, ok := v.typeValidators[v.typeMapper(change.ObjectType)]; ok {
		for _, validator := range typeValidators {
			if err := validator.Validate(change); err != nil {
				return err
			}
		}
	}

	return nil
}

func NewValidationHandler(mapper objects.ObjectTypeTranslator) *ValidationHandler {
	return &ValidationHandler{
		globalValidators: make([]Validator, 0),
		typeValidators:   make(map[int64][]Validator),
		typeMapper:       mapper,
	}
}
