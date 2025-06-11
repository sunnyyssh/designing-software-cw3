package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/sunnyyssh/designing-software-cw3/gateway/router"
	"gopkg.in/yaml.v3"
)

const configPath = "/etc/gateway/config.yaml"

func main() {
	config, err := readConfig(configPath)
	if err != nil {
		slog.Error("failed to read config", "error", err)
	}

	router := router.New(config, slog.Default())

	if err := http.ListenAndServe(":80", router); err != nil {
		slog.Error("serving http failed", "error", err)
	}
}

func readConfig(path string) (*router.Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var rawConfig struct {
		Locations map[string]struct {
			URL string `yaml:"url"`
		} `yaml:"locations"`
	}

	if err := yaml.Unmarshal(data, &rawConfig); err != nil {
		return nil, err
	}

	config := new(router.Config)
	for prefix, location := range rawConfig.Locations {
		config.Locations = append(config.Locations, router.Location{
			Prefix: prefix,
			URL:    location.URL,
		})
	}

	return config, nil
}
