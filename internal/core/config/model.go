package config

type Config struct {
	Version int `mapstructure:"version" yaml:"version"`
	Git     struct {
		DefaultBase  string `mapstructure:"default_base" yaml:"default_base"`
		BranchPrefix string `mapstructure:"branch_prefix" yaml:"branch_prefix"`
	} `mapstructure:"git" yaml:"git"`
	Policy struct {
		Enabled bool   `mapstructure:"enabled" yaml:"enabled"`
		Path    string `mapstructure:"path" yaml:"path"`
	} `mapstructure:"policy" yaml:"policy"`
	Hooks struct {
		PreRun  []string `mapstructure:"pre_run" yaml:"pre_run"`
		PostRun []string `mapstructure:"post_run" yaml:"post_run"`
	} `mapstructure:"hooks" yaml:"hooks"`
	Output struct {
		Color   bool `mapstructure:"color" yaml:"color"`
		Verbose bool `mapstructure:"verbose" yaml:"verbose"`
	} `mapstructure:"output" yaml:"output"`
}

func DefaultConfig() *Config {
	cfg := &Config{Version: 1}
	cfg.Git.DefaultBase = "main"
	cfg.Git.BranchPrefix = "bar/"
	cfg.Policy.Enabled = false
	cfg.Policy.Path = ".bar/policy.yaml"
	cfg.Hooks.PreRun = []string{}
	cfg.Hooks.PostRun = []string{}
	cfg.Output.Color = true
	cfg.Output.Verbose = false
	return cfg
}
