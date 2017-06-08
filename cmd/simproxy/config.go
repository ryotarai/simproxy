package main

import (
	"io/ioutil"

	"time"

	"gopkg.in/go-playground/validator.v9"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Listen          *string            `yaml:"listen" validate:"required"`
	Backends        []*BackendConfig   `yaml:"backends" validate:"required,dive"`
	BalancingMethod *string            `yaml:"balancing_method" validate:"required"`
	Healthcheck     *HealthcheckConfig `yaml:"healthcheck" validate:"required,dive"`
	AccessLog       *string            `yaml:"access_log" validate:"required"`
}

type HealthcheckConfig struct {
	Path      *string        `yaml:"path" validate:"required"`
	Interval  *time.Duration `yaml:"interval" validate:"required"`
	FallCount *int           `yaml:"fall_count" validate:"required"`
	RiseCount *int           `yaml:"rise_count" validate:"required"`
}

type BackendConfig struct {
	URL    *string `yaml:"url" validate:"required"`
	Weight *int    `yaml:"weight" validate:"required"`
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

func (c *Config) Validate() error {
	v := validator.New()
	err := v.Struct(c)
	return err
}
