package main

import (
	"io/ioutil"

	"time"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Listen          *string            `yaml:"listen"`
	Backends        []*BackendConfig   `yaml:"backends"`
	BalancingMethod *string            `yaml:"balancing_method"`
	Healthcheck     *HealthcheckConfig `yaml:"healthcheck"`
	AccessLog       *string            `yaml:"access_log"`
}

type HealthcheckConfig struct {
	Path      *string        `yaml:"path"`
	Interval  *time.Duration `yaml:"interval"`
	FallCount *int           `yaml:"fall_count"`
	RiseCount *int           `yaml:"rise_count"`
}

type BackendConfig struct {
	URL    *string `yaml:"url"`
	Weight *int    `yaml:"weight"`
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
