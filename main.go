package main

import (
	"backupMonitor/configuration"
	"flag"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	log "github.com/sirupsen/logrus"
	"net/http"
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
	// unix seconds
	lastRunTime int64
}

type data struct {
	backups map[string]*backupData
}

func main() {
	listenAddr := flag.String("listen-address", ":8080", "The address to listen on for HTTP requests.")
	configFile := flag.String("configuration-file", "config.yml", "The configuration file")

	// parse flags
	flag.Parse()

	// read configuration
	config := configuration.ParseConfig(*configFile)

	data := data{
		backups: map[string]*backupData{},
	}

	// setup metrics
	setupMetrics(&data, config)

	// setup handlers for backup scripts
	http.Handle("/start", newStartHandler(&data))
	http.Handle("/finish", newFinishHandler(&data))

	http.Handle("/metrics", promhttp.Handler())

	go func() {
		for {
			loop(&data)
		}
	}()

	err := http.ListenAndServe(*listenAddr, nil)
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

		log.Debugf("Updated metrics of %s", i)
	}
	time.Sleep(10 * time.Second)
}

func setupMetrics(data *data, config configuration.Config) {
	for _, backup := range config.Backups {
		data.backups[backup.Name] = &backupData{
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
