package configuration

import (
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type Config struct {
	ListenAddr           string         `yaml:"listenAddr"`
	CheckIntervalSeconds int            `yaml:"checkIntervalSeconds"`
	Backups              []BackupConfig `yaml:"backups"`
}

type BackupConfig struct {
	Name            string `yaml:"name"`
	BackupDirectory string `yaml:"backupDirectory"`
	DateTimeRegex   string `yaml:"dateTimeRegex"`
	DateTimeLayout  string `yaml:"dateTimeLayout"`
	EpochRegex      string `yaml:"epochRegex"`
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

	// validate
	for _, backup := range config.Backups {
		// TODO more
		if backup.DateTimeRegex == "" && backup.EpochRegex == "" {
			log.Fatalln("DateTimeRegex or EpochRegex must be set")
		}
	}

	return config
}
