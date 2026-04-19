package config

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Theme    ThemeConfig    `yaml:"theme"`
	API      APIConfig      `yaml:"api"`
	Generate GenerateConfig `yaml:"generate"`
}

type GenerateConfig struct {
	BaseURL   string        `yaml:"base_url"`
	Model     string        `yaml:"model"`
	APIKeyEnv string        `yaml:"api_key_env"`
	Timeout   time.Duration `yaml:"timeout"`
}

type ThemeConfig struct {
	Overrides map[string]string `yaml:"overrides"`
}

type APIConfig struct {
	Proxy map[string]ProxyTarget `yaml:"proxy"`
}

type ProxyTarget struct {
	Target  string            `yaml:"target"`
	Headers map[string]string `yaml:"headers"`
}

func Load(dir string) (*Config, error) {
	path := filepath.Join(dir, "goslide.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse goslide.yaml: %w", err)
	}

	for prefix, proxy := range cfg.API.Proxy {
		if proxy.Target != "" {
			u, parseErr := url.Parse(proxy.Target)
			if parseErr != nil || u.Scheme == "" || u.Host == "" {
				return nil, fmt.Errorf("invalid target URL for %s: %s", prefix, proxy.Target)
			}
		}

		for k, v := range proxy.Headers {
			proxy.Headers[k] = os.ExpandEnv(v)
		}
		cfg.API.Proxy[prefix] = proxy
	}

	return &cfg, nil
}
