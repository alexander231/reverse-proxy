package parsing

import (
	"os"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Proxy Proxy `yaml:"proxy"`
}

type Proxy struct {
	Listen   Host      `yaml:"listen"`
	Services []Service `yaml:"services"`
}

type Service struct {
	Name   string `yaml:"name"`
	Domain string `yaml:"domain"`
	Hosts  []Host `yaml:"hosts"`
}

type Host struct {
	Address string `yaml:"address"`
	Port    int    `yaml:"port"`
}

func GetConfig(filename string) (*Config, error) {
	cfg := &Config{}
	fileBytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, errors.Wrap(err, "Reading file")
	}
	if err := yaml.Unmarshal(fileBytes, cfg); err != nil {
		return nil, errors.Wrap(err, "Unmarshaling config")
	}
	return cfg, nil

}
