package datapoint

import (
	"backupmonitor/configuration"
	log "github.com/sirupsen/logrus"
	"io/fs"
	"regexp"
	"time"
)

type Datapoint struct {
	Name string
	Size int64
	Age  int64
}

func NewDatapoint(file fs.FileInfo, config configuration.BackupConfig) (*Datapoint, error) {
	name := file.Name()

	// extract date part with regex
	compile, err := regexp.Compile(config.DateTimeRegex)
	if err != nil {
		return nil, err
	}

	subMatch := compile.FindStringSubmatch(name)
	if len(subMatch) != 2 {
		log.Warningf("Failed to extract datetime from '%s' with regex '%s', subMatch: '%s'", name, config.DateTimeRegex, subMatch)
		return nil, nil
	}

	parse, err := time.Parse(config.DateTimeLayout, subMatch[1])
	if err != nil {
		log.Warningf("Failed to parse datetime from '%s' with layout '%s'", name, config.DateTimeLayout)
		return nil, err
	}

	return &Datapoint{
		Name: name,
		Size: file.Size(),
		Age:  time.Now().Unix() - parse.Unix(),
	}, nil
}
