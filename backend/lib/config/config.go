package config

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path"
	"time"
)

const (
	processedAudiobookFolder = "processed_audiobook"
)

type Config struct {
	Port                   int
	AudiobookDirectory     string
	ProcessedAudiobookPath string
	ScanInterval           time.Duration
	ApplicationDirectory   string
	Database               DatabaseConfig
}

type DatabaseConfig struct {
	Migrations string `json:"migrations"`
	Path       string `json:"dbPath"`
	Driver     string `json:"driver"`
}

type intermediateConfig struct {
	Port                 int            `json:"port"`
	AudiobookDirectory   string         `json:"audiobookDirectory"`
	ScanInterval         configDuration `json:"scanInterval"`
	ApplicationDirectory string         `json:"applicationDirectory"`
	Database             DatabaseConfig `json:"database"`
}

type configDuration time.Duration

func (c *configDuration) UnmarshalJSON(data []byte) error {
	var tmp string
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}
	duration, err := time.ParseDuration(tmp)
	if err != nil {
		return err
	}
	*c = configDuration(duration)
	return nil
}

func ParseConfig(configFilePath string) (*Config, error) {
	stat, err := os.Stat(configFilePath)
	if err != nil {
		return nil, err
	}
	if stat.IsDir() {
		return nil, fmt.Errorf("path %s does not point to .env file", configFilePath)
	}

	configData, err := os.ReadFile(configFilePath)
	if err != nil {
		return nil, err
	}
	intermediateConfig := intermediateConfig{}
	if err := json.Unmarshal(configData, &intermediateConfig); err != nil {
		return nil, err
	}

	config := Config{
		Port:                   intermediateConfig.Port,
		AudiobookDirectory:     intermediateConfig.AudiobookDirectory,
		ProcessedAudiobookPath: path.Join(intermediateConfig.ApplicationDirectory, processedAudiobookFolder),
		ScanInterval:           time.Duration(intermediateConfig.ScanInterval),
		ApplicationDirectory:   intermediateConfig.ApplicationDirectory,
		Database:               intermediateConfig.Database,
	}

	return &config, nil
}

func GetEnvPathFromFlags() (string, error) {
	var configPath string
	flag.StringVar(&configPath, "configPath", "", "Path to environment configuration file")
	flag.Parse()

	if len(configPath) == 0 {
		return "", errors.New("path to configuration file is not set")
	}

	stat, err := os.Stat(configPath)
	if err != nil {
		return "", err
	}
	if stat.IsDir() {
		return "", fmt.Errorf("path %s does not point to file", configPath)
	}
	return configPath, nil

}
