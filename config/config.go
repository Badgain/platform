package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"platform/database/models"
)

var (
	GlobalConfig *ServiceConfigFile
	CONFIG_PATH  = ""
	NEED_AUTH    = true
	DEBUG_MODE   = false
)

type ServiceConfigFile struct {
	Service            string                  `json:"service"`
	DatabaseConfigList []models.DatabaseConfig `json:"database_config_list"`
	Version            string                  `json:"version"`
	Host               string                  `json:"host"`
	Port               int                     `json:"port"`
}

func ReadConfig() error {
	fmt.Println("Read config file...")
	config, err := os.OpenFile(CONFIG_PATH, 0666, os.FileMode(os.O_RDONLY))
	if err != nil {
		return err
	}
	buf := bytes.Buffer{}
	_, err = buf.ReadFrom(config)
	if err != nil {
		return err
	}
	serviceConfig := ServiceConfigFile{}
	err = json.Unmarshal(buf.Bytes(), &serviceConfig)
	if err != nil {
		return err
	}
	GlobalConfig = &serviceConfig
	return nil
}

func (c *ServiceConfigFile) GetDatabaseConfig(name string) *models.DatabaseConfig {
	for _, dbc := range c.DatabaseConfigList {
		if dbc.Database == name {
			return &dbc
		}
	}
	return nil
}
