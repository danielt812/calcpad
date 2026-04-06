package main

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type ColorConfig struct {
	Normal      string `yaml:"normal"`
	Operators   string `yaml:"operators"`
	Parens      string `yaml:"parens"`
	Aggregates  string `yaml:"aggregates"`
	Constants   string `yaml:"constants"`
	Functions   string `yaml:"functions"`
	LineNumbers string `yaml:"line_numbers"`
	Results     string `yaml:"results"`
}

type KeyConfig struct {
	Quit   string `yaml:"quit"`
	Reset  string `yaml:"reset"`
	Format string `yaml:"format"`
	Help   string `yaml:"help"`
}

type Config struct {
	Colors     ColorConfig `yaml:"colors"`
	Keys       KeyConfig   `yaml:"keys"`
	Precision  int         `yaml:"precision"`   // decimal places, -1 = full precision
	AutoFormat bool        `yaml:"auto_format"` // format lines as you type
}

func defaultConfig() Config {
	return Config{
		Precision:  -1,
		AutoFormat: false,
		Colors: ColorConfig{
			Normal:      "15",
			Operators:   "14",
			Parens:      "214",
			Aggregates:  "141",
			Constants:   "220",
			Functions:   "213",
			LineNumbers: "240",
			Results:     "84",
		},
		Keys: KeyConfig{
			Quit:   "ctrl+c",
			Reset:  "ctrl+r",
			Format: "ctrl+f",
			Help:   "ctrl+h",
		},
	}
}

func loadConfig() Config {
	cfg := defaultConfig()
	home, err := os.UserHomeDir()
	if err != nil {
		return cfg
	}
	data, err := os.ReadFile(filepath.Join(home, ".config", "calcpad", "config.yaml"))
	if err != nil {
		return cfg
	}
	// Unmarshal into the populated default — only keys present in the file are overwritten.
	yaml.Unmarshal(data, &cfg) //nolint:errcheck
	return cfg
}
