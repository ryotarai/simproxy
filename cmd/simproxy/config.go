package main

import (
	"io/ioutil"

	"net/url"

	"github.com/ryotarai/simproxy"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Listen          *string          `yaml:"listen"`
	Backends        []*BackendConfig `yaml:"backends"`
	BalancingMethod *string          `yaml:"balancing_method"`
}

type BackendConfig struct {
	URL    *string `yaml:"url"`
	Weight *int    `yaml:"weight"`
}

func (b *BackendConfig) Backend() (*simproxy.Backend, error) {
	url, err := url.Parse(*b.URL)
	if err != nil {
		return nil, err
	}
	return &simproxy.Backend{
		URL:    url,
		Weight: *b.Weight,
	}, nil
}

func LoadConfigFromYAML(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	c := &Config{}
	err = yaml.Unmarshal(data, c)
	if err != nil {
		return nil, err
	}

	return c, nil
}
