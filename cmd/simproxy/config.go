package main

import (
	"io/ioutil"

	"time"

	"gopkg.in/go-playground/validator.v9"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Listen              *string            `yaml:"listen" validate:"required"`
	Backends            []*BackendConfig   `yaml:"backends" validate:"required,dive"`
	BalancingMethod     *string            `yaml:"balancing_method" validate:"required"`
	Healthcheck         *HealthcheckConfig `yaml:"healthcheck" validate:"required,dive"`
	AccessLog           *AccessLogConfig   `yaml:"access_log"`
	ErrorLog            *ErrorLogConfig    `yaml:"error_log" validate:"required,dive"`
	ReadTimeout         *time.Duration     `yaml:"read_timeout"`
	ReadHeaderTimeout   *time.Duration     `yaml:"read_header_timeout"`
	WriteTimeout        *time.Duration     `yaml:"write_timeout"`
	BackendURLHeader    *string            `yaml:"backend_url_header"`
	MaxIdleConnsPerHost *int               `yaml:"max_idle_conns_per_host"`
	MaxIdleConns        *int               `yaml:"max_idle_conns"`

	EnableBackendTrace bool    `yaml:"enable_backend_trace"`
	PprofAddr          *string `yaml:"pprof_addr"`
}

type HealthcheckConfig struct {
	Path      *string        `yaml:"path" validate:"required"`
	Interval  *time.Duration `yaml:"interval" validate:"required"`
	FallCount *int           `yaml:"fall_count" validate:"required"`
	RiseCount *int           `yaml:"rise_count" validate:"required"`
	StateFile *string        `yaml:"state_file" validate:"required"`
}

type BackendConfig struct {
	URL    *string `yaml:"url" validate:"required"`
	Weight *int    `yaml:"weight" validate:"required"`
}

type AccessLogConfig struct {
	Format *string  `yaml:"format" validate:"required"`
	Path   *string  `yaml:"path" validate:"required"`
	Fields []string `yaml:"fields" validate:"required"`
}

type ErrorLogConfig struct {
	Path *string `yaml:"path" validate:"required"`
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
