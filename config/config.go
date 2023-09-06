package config

import (
	"fmt"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/jgivc/console/host"
	"github.com/jgivc/console/transport"
	"github.com/jgivc/console/util"
)

const (
	authPromptPattern         = `(?i)((user|pass)\w+:|[\w\-]+[>#])`
	promptPattern             = `[\w\-]+#`
	authTimeout               = 5 * time.Second
	execTimeout               = 5 * time.Second
	usernamePromptContains    = "username:"
	passwordPromptContains    = "password:"
	promptSuffix              = "#"
	enableSuffix              = ">"
	enableCommand             = "en"
	promptMatchLengt          = 20
	transportReadTimeout      = time.Second
	transportReaderBufferSize = 1024
)

var (
	defaultInitialCommands = []string{"term le 0"}
	defaultExitCommand     = "q"
)

type (
	Config struct {
		DefaultConfig   ConsoleConfig `yaml:"default_config"`
		Account         host.Account  `yaml:"default_account"`
		InitialCommands []string      `yaml:"initial_commands"`
		Commands        []string      `yaml:"commands"`
		ExitCommand     string        `yaml:"exit_command"`
		Hosts           []HostConfig  `yaml:"hosts"`
	}

	HostConfig struct {
		URI             string        `yaml:"uri"`
		InitialCommands []string      `yaml:"initial_commands"`
		Commands        []string      `yaml:"commands"`
		ExitCommand     string        `yaml:"exit_command"`
		Host            host.Host     `yaml:"-"`
		DummyConfig     string        `yaml:"-"`
		ConsoleConfig   ConsoleConfig `yaml:"console_config"`
	}

	ConsoleConfig struct {
		AuthPromptPattern         string        `yaml:"auth_prompt_pattern"`
		PromptPattern             string        `yaml:"prompt_pattern"`
		AuthTimeout               time.Duration `yaml:"auth_timeout"`
		ExecTimeout               time.Duration `yaml:"exec_timeout"`
		UsernamePromptContains    string        `yaml:"username_prompt_contains"`
		PasswordPromptContains    string        `yaml:"password_prompt_contains"`
		PromptSuffix              string        `yaml:"prompt_suffix"`
		EnableSuffix              string        `yaml:"enable_suffix"`
		EnableCommand             string        `yaml:"enable_command"`
		PromptMatchLengt          int           `yaml:"prompt_match_lengt"`
		TransportReadTimeout      time.Duration `yaml:"transport_read_timeout"`
		TransportReaderBufferSize int           `yaml:"transport_reader_buffer_size"`
		DummyTransportFileName    string        `yaml:"-"`
	}
)

func (c *HostConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*c = HostConfig{
		ConsoleConfig: *DefaultConsoleConfig(),
	}

	var obj interface{}

	if err := unmarshal(&obj); err != nil {
		return err
	}

	switch v := obj.(type) {
	case string:
		c.URI = v
	default:
		type hc HostConfig
		return unmarshal((*hc)(c))
	}

	return nil
}

type FromFlags struct {
	Commands    []string
	Account     *host.Account
	DummyConfig string
}

func Load(fileName string, flags *FromFlags) (*Config, error) {
	var (
		cfg Config
	)

	if err := cleanenv.ReadConfig(fileName, &cfg); err != nil {
		return nil, fmt.Errorf("cannot open config: %w", err)
	}

	if cfg.InitialCommands == nil {
		cfg.InitialCommands = defaultInitialCommands
	}

	if cfg.ExitCommand == "" {
		cfg.ExitCommand = defaultExitCommand
	}

	if flags.Commands != nil {
		cfg.Commands = flags.Commands
	}

	if flags.Account != nil {
		cfg.Account = *flags.Account
	}

	if cfg.Account.Password == "" {
		return nil, fmt.Errorf("no account defined")
	}

	factory := util.NewHostFactory(cfg.Account)

	for i := range cfg.Hosts {
		h, err := factory.GetHost(cfg.Hosts[i].URI)
		if err != nil {
			return nil, fmt.Errorf("cannot convert uri to host: %w", err)
		}
		cfg.Hosts[i].Host = *h

		if cfg.Hosts[i].InitialCommands == nil {
			cfg.Hosts[i].InitialCommands = cfg.InitialCommands
		}

		if flags.Commands != nil || cfg.Hosts[i].Commands == nil {
			cfg.Hosts[i].Commands = cfg.Commands
		}

		if cfg.Hosts[i].ExitCommand == "" {
			cfg.Hosts[i].ExitCommand = cfg.ExitCommand
		}

		if cfg.Hosts[i].Commands == nil {
			return nil, fmt.Errorf("commands cannot be empty for host: %s", cfg.Hosts[i].Host.Host)
		}

		if flags.DummyConfig != "" {
			cfg.Hosts[i].ConsoleConfig.DummyTransportFileName = flags.DummyConfig
			cfg.Hosts[i].Host.TransportType = transport.TransportDummy
		}
	}

	return &cfg, nil
}

func DefaultConsoleConfig() *ConsoleConfig {
	return &ConsoleConfig{
		AuthPromptPattern:         authPromptPattern,
		PromptPattern:             promptPattern,
		AuthTimeout:               authTimeout,
		ExecTimeout:               execTimeout,
		UsernamePromptContains:    usernamePromptContains,
		PasswordPromptContains:    passwordPromptContains,
		PromptSuffix:              promptSuffix,
		EnableSuffix:              enableSuffix,
		EnableCommand:             enableCommand,
		PromptMatchLengt:          promptMatchLengt,
		TransportReadTimeout:      transportReadTimeout,
		TransportReaderBufferSize: transportReaderBufferSize,
	}
}
