package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server ServerConfig `yaml:"server"`
	Checks ChecksConfig `yaml:"checks"`
}

type ServerConfig struct {
	Port int `yaml:"port"`
}

type ChecksConfig struct {
	LLM            LLMCheck            `yaml:"llm"`
	TokenBudget    TokenBudgetCheck    `yaml:"token_budget"`
	Tools          ToolsCheck          `yaml:"tools"`
	ResponseQuality ResponseQualityCheck `yaml:"response_quality"`
}

type LLMCheck struct {
	Enabled   bool          `yaml:"enabled"`
	Endpoint  string        `yaml:"endpoint"`
	APIKeyEnv string        `yaml:"api_key_env"`
	Timeout   time.Duration `yaml:"timeout"`
}

type TokenBudgetCheck struct {
	Enabled           bool          `yaml:"enabled"`
	Endpoint          string        `yaml:"endpoint"`
	APIKeyEnv         string        `yaml:"api_key_env"`
	WarnThreshold     float64       `yaml:"warn_threshold"`
	CriticalThreshold float64       `yaml:"critical_threshold"`
	Timeout           time.Duration `yaml:"timeout"`
}

type ToolsCheck struct {
	Enabled   bool            `yaml:"enabled"`
	Endpoints []ToolEndpoint  `yaml:"endpoints"`
}

type ToolEndpoint struct {
	Name    string        `yaml:"name"`
	URL     string        `yaml:"url"`
	Timeout time.Duration `yaml:"timeout"`
}

type ResponseQualityCheck struct {
	Enabled           bool          `yaml:"enabled"`
	Endpoint          string        `yaml:"endpoint"`
	APIKeyEnv         string        `yaml:"api_key_env"`
	TestPrompt        string        `yaml:"test_prompt"`
	ExpectedSubstring string        `yaml:"expected_substring"`
	Timeout           time.Duration `yaml:"timeout"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		Server: ServerConfig{Port: 8089},
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
