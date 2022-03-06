package main

import (
	"backupmonitor/configuration"
	"backupmonitor/datapoint"
	"flag"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"sort"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type metrics struct {
	lastBackupSize  prometheus.Gauge
	lastBackupAge   prometheus.Gauge
	lastRunAge      prometheus.Gauge
	lastRunDuration prometheus.Gauge
}

type backupData struct {
	metrics *metrics
	config  configuration.BackupConfig
	// unix seconds
	lastRunTime int64
}

type data struct {
	backups map[string]*backupData
}

func main() {
	configFile := flag.String("configuration-file", "config.yml", "The configuration file")

	// parse flags
	flag.Parse()

	// read configuration
	config := configuration.ParseConfig(*configFile)

	data := data{
		backups: map[string]*backupData{},
	}

	// setup metrics
	setupData(&data, config)

	// setup handlers for backup scripts
	http.Handle("/start", newStartHandler(&data))
	http.Handle("/finish", newFinishHandler(&data))

	http.Handle("/metrics", promhttp.Handler())

	go func() {
		for {
			loop(&data)
			time.Sleep(time.Duration(config.CheckIntervalSeconds) * time.Second)
		}
	}()

	log.Infof("Starting webserver on addr %s", config.ListenAddr)
	err := http.ListenAndServe(config.ListenAddr, nil)
	if err != nil {
		log.Fatalln("Error while listening", err)
	}

}

func loop(data *data) {
	log.Debugln("Running loop to update metrics")
	// main runtime for updating metrics
	for i, backup := range data.backups {
		now := time.Now().Unix()
		if backup.lastRunTime > 0 {
			// update last run age
			age := now - backup.lastRunTime
			backup.metrics.lastRunAge.Set(float64(age))
		}

		fileInfos, err := ioutil.ReadDir(backup.config.BackupDirectory)
		if err != nil {
			log.Errorln("Failed to read backup directory, ignoring...", err.Error())
			break
		}

		var datapoints []*datapoint.Datapoint
		for _, fileInfo := range fileInfos {
			dp, err := datapoint.NewDatapoint(fileInfo, backup.config)
			if err != nil {
				log.Errorln("Invalid file in backup folder", err)
				break
			}
			datapoints = append(datapoints, dp)
		}

		sort.SliceStable(datapoints, func(i, j int) bool {
			return datapoints[i].Age < datapoints[j].Age
		})

		if len(datapoints) == 0 {
			log.Warningf("No backups found for %s in '%s'", backup.config.Name, backup.config.BackupDirectory)
			break
		}
		var newest = datapoints[0]

		backup.metrics.lastBackupSize.Set(float64(newest.Size))
		backup.metrics.lastBackupAge.Set(float64(newest.Age))

		log.Debugf("Updated metrics of %s", i)
	}
}

func setupData(data *data, config configuration.Config) {
	for _, backup := range config.Backups {
		data.backups[backup.Name] = &backupData{
			config:      backup,
			lastRunTime: -1,
			metrics: &metrics{
				lastBackupAge: promauto.NewGauge(prometheus.GaugeOpts{
					Namespace:   "backup_monitor",
					Name:        "last_backup_age_seconds",
					Help:        "Age in seconds of the last backup file",
					ConstLabels: prometheus.Labels{"backup_name": backup.Name},
				}),
				lastBackupSize: promauto.NewGauge(prometheus.GaugeOpts{
					Namespace:   "backup_monitor",
					Name:        "last_backup_size_bytes",
					Help:        "Size in bytes of the last backup file",
					ConstLabels: prometheus.Labels{"backup_name": backup.Name},
				}),
				lastRunAge: promauto.NewGauge(prometheus.GaugeOpts{
					Namespace:   "backup_monitor",
					Name:        "last_backup_run_age_seconds",
					Help:        "How long since the backup script was executed the last time",
					ConstLabels: prometheus.Labels{"backup_name": backup.Name},
				}),
				lastRunDuration: promauto.NewGauge(prometheus.GaugeOpts{
					Namespace:   "backup_monitor",
					Name:        "last_backup_run_duration_seconds",
					Help:        "The last execution time of the backup script in seconds",
					ConstLabels: prometheus.Labels{"backup_name": backup.Name},
				}),
			}}
	}

}
