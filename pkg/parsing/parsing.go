package parsing

import (
	"os"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

type Config interface {
	GetProxyPort() int
	GetProxyAddress() string
	GetServices() []Service
	GetLbPolicy() string
}
type config struct {
	Proxy proxy `yaml:"proxy"`
}

type proxy struct {
	LbPolicy string    `yaml:"lbPolicy"`
	Listen   host      `yaml:"listen"`
	Services []Service `yaml:"services"`
}

type Service struct {
	Name   string `yaml:"name"`
	Domain string `yaml:"domain"`
	Hosts  []host `yaml:"hosts"`
}

type host struct {
	Address string `yaml:"address"`
	Port    int    `yaml:"port"`
}

func NewConfig(filepath string) (*config, error) {
	cfg := &config{}
	fileBytes, err := os.ReadFile(filepath)
	if err != nil {
		return nil, errors.Wrap(err, "Reading file")
	}
	if err := yaml.Unmarshal(fileBytes, cfg); err != nil {
		return nil, errors.Wrap(err, "Unmarshaling config")
	}
	return cfg, nil
}

func (c *config) GetProxyPort() int {
	return c.Proxy.Listen.Port
}

func (c *config) GetProxyAddress() string {
	return c.Proxy.Listen.Address
}

func (c *config) GetServices() []Service {
	return c.Proxy.Services
}

func (c *config) GetLbPolicy() string {
	return c.Proxy.LbPolicy
}
