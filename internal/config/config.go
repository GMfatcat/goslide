package config

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	API APIConfig `yaml:"api"`
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
