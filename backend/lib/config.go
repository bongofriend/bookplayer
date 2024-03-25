package config

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/spf13/viper"
)

type Environment string

const (
	Dev  Environment = "development"
	Test Environment = "test"
)

type Config struct {
	Server     ServerConfig     `mapstructure:"server"`
	Audiobooks AudiobooksConfig `mapstructure:"audiobooks"`
}

type ServerConfig struct {
	Port int `yaml:"PORT"`
}

type AudiobooksConfig struct {
	AudibookDirectoryPath string        `mapstructure:"DATA_DIR"`
	Interval              time.Duration `yaml:"INTERVAL"`
}

func GetConfig(env Environment) (Config, error) {
	var envPath string

	switch env {
	case Test:
		envPath = "./dev.yml"
	case Dev:
		flag.StringVar(&envPath, "env", "", "path to .env file")
		flag.Parse()
	default:
		return Config{}, errors.New("invalid environment option")
	}

	if len(envPath) == 0 {
		return Config{}, fmt.Errorf("path .env file not provided")
	}
	serverConfig, err := parseConfig(envPath)
	if err != nil {
		return Config{}, err
	}
	return serverConfig, nil

}

func parseConfig(envPath string) (Config, error) {
	stat, err := os.Stat(envPath)
	if err != nil {
		return Config{}, err
	}
	if stat.IsDir() {
		return Config{}, fmt.Errorf("path %s does not point to .env file", envPath)
	}
	viper.SetConfigFile(envPath)
	if err := viper.ReadInConfig(); err != nil {
		return Config{}, err
	}
	config := Config{}
	viper.Unmarshal(&config)
	return config, nil
}
