package config

import (
	"errors"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/BurntSushi/toml"
)

// Config for supervisor, created during -install.
type Config struct {
	MystHome    string
	MystPath    string
	OpenVPNPath string
}

func (c Config) valid() bool {
	return c.MystHome != "" &&
		c.MystPath != "" &&
		c.OpenVPNPath != ""
}

// Write config file.
func (c Config) Write() error {
	if !c.valid() {
		return errors.New("configuration is not valid")
	}
	var out strings.Builder
	err := toml.NewEncoder(&out).Encode(c)
	if err != nil {
		return fmt.Errorf("could not encode cofiguration: %w", err)
	}
	if err := ioutil.WriteFile(Path, []byte(out.String()), 0700); err != nil {
		return fmt.Errorf("could not write %q: %w", Path, err)
	}
	return nil
}

// Read config file.
func Read() (*Config, error) {
	c := Config{}
	_, err := toml.DecodeFile(Path, &c)
	if err != nil {
		return nil, fmt.Errorf("could not read %q: %w", Path, err)
	}
	if !c.valid() {
		return nil, fmt.Errorf("invalid configuration file %q, please re-install the supervisor (-install)", Path)
	}
	return &c, nil
}
