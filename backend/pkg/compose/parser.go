package compose

import (
	"encoding/json"

	"github.com/ghodss/yaml"
)

func FromString(s string, target interface{}) error {
	err := json.Unmarshal([]byte(s), target)
	if err != nil {
		return err
	}
	return nil
}

func (dc *DockerCompose) ToJSONString() (string, error) {
	b, err := json.Marshal(dc)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (dc *DockerCompose) FromStringToYAML(s string) (string, error) {
	b, err := yaml.JSONToYAML([]byte(s))
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (dc *DockerCompose) ToYAML() (string, error) {
	b, err := yaml.Marshal(dc)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
