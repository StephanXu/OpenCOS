package config

import (
	"encoding/json"
	"errors"
	"os"
)

type AppContext struct {
	ContextFile   string      `json:"contextFile"`
	DefaultSource string      `json:"defaultSource"`
	PlexHost      string      `json:"plexHost"`
	LocalHash     string      `json:"localHash"`
	Port          int32       `json:"port"`
	Sources       interface{} `json:"sources"`
}

var App AppContext

func LoadContextFromConfigFile(filename string) error {
	var config AppContext

	content, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(content, &config); err != nil {
		return err
	}

	if len(config.ContextFile) == 0 {
		return errors.New("EmptyContextFilename")
	}
	if _, err := os.Stat(config.ContextFile); os.IsNotExist(err) {
		App = config
		return SaveContextToFile(config.ContextFile)
	}
	return LoadContextFromContextFile(config.ContextFile)
}

func LoadContextFromContextFile(filename string) error {
	content, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(content, &App); err != nil {
		return err
	}
	return nil
}

func SaveContextToFile(filename string) error {
	content, err := json.MarshalIndent(App, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, content, 0666)
}

func SaveContext() error {
	if len(App.ContextFile) == 0 {
		return errors.New("EmptyContextFilename")
	}
	return SaveContextToFile(App.ContextFile)
}
