package objects

import "Plow/plow/utility"

type ObjectDesignation struct {
	ObjectType string `yaml:"type"`
	Identifier string `yaml:"id"`
}

type ObjectSpec struct {
	Name     string `yaml:"name,omitempty"`
	Database string `yaml:"database,omitempty"`
	Schema   string `yaml:"schema,omitempty"`
}

type OptionsSpec struct {
	CheckExists bool `yaml:"checkExists"`
	Validate    bool `yaml:"validate"`
	Drop        bool `yaml:"drop"`
}

type CodeBlockSpec struct {
	CodeBlockHeaderSpec   `yaml:",inline"`
	VariableDefinitionSet `yaml:",inline"`
	Object                ObjectSpec  `yaml:"object"`
	Options               OptionsSpec `yaml:"options"`
	Spec                  interface{} `yaml:"spec"`
}

type CodeBlockHeaderSpec struct {
	DefinitionStyle string `yaml:"definitionStyle"`
	Type            string `yaml:"type"`
}

func (cb *CodeBlockSpec) ExtractVariableValues(context *PlowContext) (map[string]interface{}, error) {
	output := make(map[string]interface{})
	for k, v := range cb.Variables {
		iv, err := cb.interpretVariableValue(v, context)
		if err != nil {
			return nil, err
		}
		output[k] = iv
	}

	output = cb.ExtractObjectValues(output)
	return output, nil
}

func (cb *CodeBlockSpec) ExtractObjectValues(input map[string]interface{}) map[string]interface{} {
	var output map[string]interface{}
	if input != nil {
		output = input
	} else {
		output = make(map[string]interface{})
	}

	if !utility.IsStringEmpty(&cb.Object.Name) {
		output["NAME"] = cb.Object.Name
	}
	if !utility.IsStringEmpty(&cb.Object.Database) {
		output["DATABASE"] = cb.Object.Database
	}
	if !utility.IsStringEmpty(&cb.Object.Schema) {
		output["SCHEMA"] = cb.Object.Schema
	}
	return output
}

func (cb *CodeBlockSpec) interpretVariableValue(def interface{}, context *PlowContext) (interface{}, error) {
	var typeInfo VariableDefinition
	err := utility.UnmarshalYamlSubObject(def, &typeInfo)
	if err != nil {
		return nil, err
	}

	switch StringToVariableType(typeInfo.VarType) {
	case ValueVariableType:
		{
			var valvar ValueVariableDefinition
			err := utility.UnmarshalYamlSubObject(def, &valvar)
			if err != nil {
				return nil, err
			}
			return valvar.Value, nil
		}
	case ConfigMapVariableType:
		{
			var configmap ConfigVariableMapDefinition
			err := utility.UnmarshalYamlSubObject(def, &configmap)
			if err != nil {
				return nil, err
			}
			if v, ok := context.Config.Vars[configmap.ConfigItem]; ok {
				if sv, ok := configmap.ValueMap[v]; ok {
					return cb.interpretVariableValue(sv, context)
				} else {
					return nil, ErrVariableUndefined
				}
			} else {
				return nil, ErrVariableUndefined
			}
		}
	case SecretVariableType:
		{
			var secvar SecretVariableDefinition
			err := utility.UnmarshalYamlSubObject(def, &secvar)
			if err != nil {
				return nil, err
			}
			return context.SecretStore.GetSecret(secvar.Secret)
		}
	}
	return nil, nil
}
