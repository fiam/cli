package context

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v3"
)

func parseOrSetupConfigFile(fn string) (Config, error) {
	config, err := parseConfig(fn)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		return setupConfigFile(fn)
	}
	return config, err
}

func ParseDefaultConfig() (Config, error) {
	return parseConfig(configFile())
}

var readConfig = func(fn string) ([]byte, error) {
	f, err := os.Open(fn)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func parseConfigFile(fn string) ([]byte, *yaml.Node, error) {
	data, err := readConfig(fn)
	if err != nil {
		return nil, nil, err
	}

	var root yaml.Node
	err = yaml.Unmarshal(data, &root)
	if err != nil {
		return data, nil, err
	}
	if len(root.Content) < 1 {
		return data, &root, fmt.Errorf("malformed config")
	}
	if root.Content[0].Kind != yaml.MappingNode {
		return data, &root, fmt.Errorf("expected a top level map")
	}

	return data, &root, nil
}

func isLegacy(root *yaml.Node) bool {
	for _, v := range root.Content[0].Content {
		if v.Value == "hosts" {
			return false
		}
	}

	return true
}

func parseConfig(fn string) (Config, error) {
	_, root, err := parseConfigFile(fn)
	if err != nil {
		return nil, err
	}

	if isLegacy(root) {
		return NewLegacyConfig(root), nil
	}

	return NewConfig(root), nil
}
