package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	DbURL           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

func (c Config) SetUser(name string) {
	c.CurrentUserName = name
	if err := write(c); err != nil {
		fmt.Errorf("Cannot save user")
	}
}

const configFile = ".gatorconfig.json"

func Read() (Config, error) {
	data, err := os.ReadFile(getConfigFilePath())
	var config Config
	if err != nil {
		return config, err
	}
	if err := json.Unmarshal(data, &config); err != nil {
		return config, err
	}
	return config, nil
}

func write(cfg Config) error {
	jsonString, _ := json.Marshal(cfg)
	err := os.WriteFile(getConfigFilePath(), jsonString, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

func getConfigFilePath() string {
	homeDir, _ := os.UserHomeDir()
	return fmt.Sprintf("%s/%s", homeDir, configFile)
}
