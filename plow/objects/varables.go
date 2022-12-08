package objects

import (
	"errors"
	"strings"
)

type VariableDefinitionType uint

const (
	ValueVariableType VariableDefinitionType = iota
	SecretVariableType
	ConfigMapVariableType
)

var (
	ErrVariableUndefined = errors.New("undefined variable")
)

func StringToVariableType(input string) VariableDefinitionType {
	switch strings.TrimSpace(strings.ToLower(input)) {
	case "secret":
		return SecretVariableType
	case "configmap":
		return ConfigMapVariableType
	default:
		return ValueVariableType
	}
}

type VariableDefinition struct {
	VarType string `yaml:"type,omitempty"`
}

type ValueVariableDefinition struct {
	VariableDefinition `yaml:",inline"`
	Value              interface{} `yaml:"value"`
}

type SecretVariableDefinition struct {
	VariableDefinition `yaml:",inline"`
	Secret             string `yaml:"secret"`
}

type ConfigVariableMapDefinition struct {
	VariableDefinition `yaml:",inline"`
	ConfigItem         string                      `yaml:"configItem"`
	ValueMap           map[interface{}]interface{} `yaml:"map"`
}

type VariableDefinitionSet struct {
	Variables map[string]interface{} `yaml:"vars"`
}
