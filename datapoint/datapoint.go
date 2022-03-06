package datapoint

import (
	"backupmonitor/configuration"
	log "github.com/sirupsen/logrus"
	"io/fs"
	"regexp"
	"strconv"
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
	var unixTime int64
	var regex string
	if config.DateTimeRegex != "" {
		regex = config.DateTimeRegex
	} else {
		regex = config.EpochRegex
	}
	compile, err := regexp.Compile(regex)
	if err != nil {
		return nil, err
	}

	subMatch := compile.FindStringSubmatch(name)
	if len(subMatch) != 2 {
		log.Warningf("Failed to extract from '%s' with regex '%s', subMatch: '%s'", name, regex, subMatch)
		return nil, nil
	}

	if config.DateTimeRegex != "" {
		date, err := time.Parse(config.DateTimeLayout, subMatch[1])
		if err != nil {
			log.Warningf("Failed to parse datetime from '%s' with layout '%s'", name, config.DateTimeLayout)
			return nil, err
		}
		unixTime = date.Unix()
	} else {
		unixTime, err = strconv.ParseInt(subMatch[1], 10, 64)
		if err != nil {
			log.Warningf("Failed to parse unix timesamp from '%s' with value '%s'", name, subMatch[1])
			return nil, err
		}
	}

	return &Datapoint{
		Name: name,
		Size: file.Size(),
		Age:  time.Now().Unix() - unixTime,
	}, nil
}
