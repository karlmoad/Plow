package utility

import "gopkg.in/yaml.v3"

func UnmarshalYamlSubObject(in interface{}, out interface{}) error {
	//TODO This is a hack rework this
	bytes, err := yaml.Marshal(in)
	if err != nil {
		return nil
	}
	return yaml.Unmarshal(bytes, out)
}
