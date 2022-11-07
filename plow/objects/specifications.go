package objects

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

type VariablesEntrySpec struct {
	Value string `yaml:"value,omitempty"`
	Key   string `yaml:"key,omitempty"`
}

type SecretEntrySpec struct {
	Source string `yaml:"source,omitempty"`
	Key    string `yaml:"key,omitempty"`
}

type VariablesSpec struct {
	Variables map[string]VariablesEntrySpec `yaml:"variables"`
	Secrets   map[string]SecretEntrySpec    `yaml:"secrets"`
}

type CodeBlockSpec struct {
	CodeBlockHeaderSpec `yaml:",inline"`
	Object              ObjectSpec    `yaml:"object"`
	Options             OptionsSpec   `yaml:"options"`
	Variables           VariablesSpec `yaml:"variables"`
	Spec                interface{}   `yaml:"spec"`
}

type CodeBlockHeaderSpec struct {
	DefinitionStyle string `yaml:"definitionStyle"`
	Type            string `yaml:"type"`
}
