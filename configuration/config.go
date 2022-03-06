package configuration

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type Config struct {
	Backups []BackupConfig `yaml:"backups"`
}

type BackupConfig struct {
	Name            string `yaml:"name"`
	BackupDirectory string `yaml:"backupDirectory"`
}

func ParseConfig(filename string) Config {
	yamlFile, err := ioutil.ReadFile(filename)

	if err != nil {
		panic(err)
	}

	var config Config

	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		panic(err)
	}

	return config
}
