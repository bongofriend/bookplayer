package config

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/spf13/viper"
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

func ParseConfig(envPath string) (Config, error) {
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

func GetEnvPathFromFlags() (string, error) {
	var envPath string
	flag.StringVar(&envPath, "envPath", "", "Path to environment configuration file")
	flag.Parse()

	if len(envPath) == 0 {
		return "", errors.New("path to configuration file is not set")
	}

	stat, err := os.Stat(envPath)
	if err != nil {
		return "", err
	}
	if stat.IsDir() {
		return "", fmt.Errorf("path %s does not point to file", envPath)
	}
	return envPath, nil

}
