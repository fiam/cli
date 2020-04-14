package context

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

const defaultHostname = "github.com"
const defaultGitProtocol = "https"

type NotFoundError struct {
	error
}

type Config interface {
	Hosts() ([]*HostConfig, error)
	ConfigForHost(string) (*HostConfig, error)
	DefaultHostConfig() (*HostConfig, error)
	Get(string, string) (string, error) // accept potentially blank hostname
	Set(string, string, string) error   // accept potentially blank hostname
	Write() error
}

type AuthConfig struct {
	ConfigMap
	User  string
	Token string `yaml:"oauth_token"`
}

type HostConfig struct {
	Host  string
	Auths []*AuthConfig
}

type ConfigMap struct {
	Root *yaml.Node
}

func (cm *ConfigMap) GetStringValue(key string) (string, error) {
	_, valueNode, err := cm.FindEntry(key)
	var notFound *NotFoundError

	if err != nil && errors.As(err, &notFound) {
		return defaultFor(key), nil
	} else if err != nil {
		return "", err
	}

	if valueNode.Value == "" {
		return defaultFor(key), nil
	}

	return valueNode.Value, nil
}

func (cm *ConfigMap) SetStringValue(key, value string) error {
	_, valueNode, err := cm.FindEntry(key)

	var notFound *NotFoundError

	if err != nil && errors.As(err, &notFound) {
		keyNode := &yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: key,
		}
		valueNode = &yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: "",
		}

		cm.Root.Content[0].Content = append(cm.Root.Content[0].Content, keyNode, valueNode)
	} else if err != nil {
		return err
	}

	valueNode.Value = value
	return nil
}

func (cm *ConfigMap) FindEntry(key string) (keyNode, valueNode *yaml.Node, err error) {
	err = nil

	topLevelKeys := cm.Root.Content
	for i, v := range topLevelKeys {
		if v.Value == key && i+1 < len(topLevelKeys) {
			keyNode = v
			valueNode = topLevelKeys[i+1]
			return
		}
	}

	return nil, nil, &NotFoundError{errors.New("not found")}
}

func NewConfig(root *yaml.Node) Config {
	return &fileConfig{
		ConfigMap:    ConfigMap{Root: root.Content[0]},
		documentRoot: root,
	}
}

type fileConfig struct {
	ConfigMap
	documentRoot *yaml.Node
	hosts        []*HostConfig
}

func (c *fileConfig) Get(hostname, key string) (string, error) {
	if hostname == "" {
		return c.GetStringValue(key)
	} else {
		hostCfg, err := c.ConfigForHost(hostname)
		if err != nil {
			return "", err
		}
		// TODO change once we actually support using different usernames for a given host
		return hostCfg.Auths[0].GetStringValue(key)
	}
}

func (c *fileConfig) Set(hostname, key, value string) error {
	if hostname == "" {
		return c.SetStringValue(key, value)
	} else {
		hostCfg, err := c.ConfigForHost(hostname)
		if err != nil {
			return err
		}
		// TODO change once we actually support using different usernames for a given host
		return hostCfg.Auths[0].SetStringValue(key, value)
	}
}

func (c *fileConfig) ConfigForHost(hostname string) (*HostConfig, error) {
	return hostConfigByHostname(c, hostname)
}

func (c *fileConfig) DefaultHostConfig() (*HostConfig, error) {
	return c.ConfigForHost(defaultHostname)
}

func (c *fileConfig) Write() error {
	marshalled, err := yaml.Marshal(c.documentRoot)
	if err != nil {
		return err
	}

	cfgFile, err := os.OpenFile(configFile(), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600) // cargo coded from setup
	if err != nil {
		return err
	}
	defer cfgFile.Close()

	n, err := cfgFile.Write(marshalled)
	if err == nil && n < len(marshalled) {
		err = io.ErrShortWrite
	}

	if err != nil {
		return err
	}

	return nil
}

func (c *fileConfig) Hosts() ([]*HostConfig, error) {
	if len(c.hosts) > 0 {
		return c.hosts, nil
	}

	_, hostsEntry, err := c.FindEntry("hosts")
	if err != nil {
		return nil, fmt.Errorf("could not find hosts config: %s", err)
	}

	hostConfigs, err := parseHosts(hostsEntry)
	if err != nil {
		return nil, fmt.Errorf("could not parse hosts config: %s", err)
	}

	c.hosts = hostConfigs

	return hostConfigs, nil
}

func NewLegacyConfig(root *yaml.Node) Config {
	return &LegacyConfig{Root: root}
}

type LegacyConfig struct {
	Root  *yaml.Node
	hosts []*HostConfig
}

func (lc *LegacyConfig) Get(hostname, key string) (string, error) {
	return "", nil
}

func (lc *LegacyConfig) Set(hostname, key, value string) error {
	cfgFilename := configFile()
	data, err := readConfig(cfgFilename)
	if err != nil {
		return err
	}

	newConfig := "hosts:\n"
	for _, line := range strings.Split(string(data), "\n") {
		newConfig += fmt.Sprintf("  %s\n", line)
	}

	err = os.Rename(cfgFilename, cfgFilename+".bak")
	if err != nil {
		return fmt.Errorf("failed to back up existing config: %s", err)
	}

	cfgFile, err := os.OpenFile(cfgFilename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to open new config file for writing: %s", err)
	}
	defer cfgFile.Close()

	n, err := cfgFile.WriteString(newConfig)
	if err == nil && n < len(newConfig) {
		err = io.ErrShortWrite
	}

	if err != nil {
		return err
	}

	_, root, err := parseConfigFile(cfgFilename)
	if err != nil {
		return err
	}

	return NewConfig(root).Set(hostname, key, value)
}

func (lc *LegacyConfig) ConfigForHost(hostname string) (*HostConfig, error) {
	return hostConfigByHostname(lc, hostname)
}

func (lc *LegacyConfig) DefaultHostConfig() (*HostConfig, error) {
	return lc.ConfigForHost(defaultHostname)
}

func (lc *LegacyConfig) Hosts() ([]*HostConfig, error) {
	if len(lc.hosts) > 0 {
		return lc.hosts, nil
	}

	hostConfigs, err := parseHosts(lc.Root.Content[0])
	if err != nil {
		return nil, fmt.Errorf("could not parse hosts config: %s", err)
	}

	lc.hosts = hostConfigs

	return hostConfigs, nil
}

func (lc *LegacyConfig) Write() error {
	panic("should never be called; see .Set")
}

func parseHosts(hostsEntry *yaml.Node) ([]*HostConfig, error) {
	hostConfigs := []*HostConfig{}
	malformedError := errors.New("malformed hosts config")

	for i, v := range hostsEntry.Content {
		if v.Value == "" {
			continue
		}
		if i+1 == len(hostsEntry.Content) {
			return hostConfigs, malformedError
		}
		hostConfig := HostConfig{}
		hostConfig.Host = v.Value
		authsRoot := hostsEntry.Content[i+1]
		for _, v := range authsRoot.Content {
			authConfig := AuthConfig{ConfigMap: ConfigMap{Root: v}}
			authTopLevelKeys := v.Content
			for j, v := range authTopLevelKeys {
				switch v.Value {
				case "user":
					if j+1 == len(hostsEntry.Content) {
						return hostConfigs, malformedError
					}
					authConfig.User = authTopLevelKeys[j+1].Value
				case "oauth_token":
					if j+1 == len(hostsEntry.Content) {
						return hostConfigs, malformedError
					}
					authConfig.Token = authTopLevelKeys[j+1].Value
				}
			}
			hostConfig.Auths = append(hostConfig.Auths, &authConfig)
		}
		hostConfigs = append(hostConfigs, &hostConfig)
	}

	return hostConfigs, nil
}

func hostConfigByHostname(c Config, hostname string) (*HostConfig, error) {
	hosts, err := c.Hosts()
	if err != nil {
		return nil, fmt.Errorf("failed to parse hosts config: %s", err)
	}

	for _, hc := range hosts {
		if hc.Host == hostname {
			return hc, nil
		}
	}
	return nil, fmt.Errorf("could not find config entry for %q", hostname)
}

func defaultFor(key string) string {
	// we only have a set default for one setting right now
	switch key {
	case "git_protocol":
		return defaultGitProtocol
	default:
		return ""
	}
}
