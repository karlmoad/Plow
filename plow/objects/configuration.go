package objects

type DirectoryType int

const (
	MemoryDirectory = 1
	LocalDirectory  = 2
)

type GitConfiguration struct {
	SSHKeyFile        string `yaml:"sshkey"`
	KeyPasswordSecret string `yaml:"passwordSecret"`
	Url               string `yaml:"url"`
	Branch            string `yaml:"branch"`
}

type Configuration struct {
	TargetType      string                 `yaml:"targetType"`
	SecretStoreType string                 `yaml:"secretStoreType"`
	SecretStore     map[string]interface{} `yaml:"secretStore"`
	Target          map[string]interface{} `yaml:"target"`
	GitConfig       GitConfiguration       `yaml:"git"`
	Vars            map[string]interface{} `yaml:"vars"`
}

type SystemConfiguration struct {
	Environments map[string]Configuration `yaml:"environments"`
}
