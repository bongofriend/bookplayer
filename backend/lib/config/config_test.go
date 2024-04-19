package config_test

import (
	"testing"

	"github.com/bongofriend/bookplayer/backend/lib/config"
)

const testConfigFilePath = "../../config.json"

func TestConfig(t *testing.T) {
	config, err := config.ParseConfig(testConfigFilePath)
	if err != nil {
		t.Fatal(err)
	}
	if config == nil {
		t.Fatalf("Config at %s could not be parsed", testConfigFilePath)
	}
}
