package config

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"os"
)

type Config struct {
	BackconnectAddr string `yaml:"backconnect_address"`
	AdminAddr       string `yaml:"admin_interface_address"`
	Logfile         string `yaml:"log_file"`
}

func New(configPath string) (*Config, error) {
	config := &Config{}
	if err := isFile(configPath); err != nil {
		return nil, err
	}
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	d := yaml.NewDecoder(file)
	if err := d.Decode(&config); err != nil {
		return nil, err
	}
	return config, nil
}

func isFile(path string) error {
	s, err := os.Stat(path)
	if err != nil {
		return err
	}
	if s.IsDir() {
		return fmt.Errorf("'%s' is a directory, not a normal file", path)
	}
	return nil
}
