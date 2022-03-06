package main

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

type startHandler struct {
	data *data
}

type StartBody struct {
	Name string `json:"name"`
}

func (sh startHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		// Declare a new StartBody struct.
		var body StartBody

		// Try to decode the request body into the struct. If there is an error,
		// respond to the client with the error message and a 400 status code.
		err := json.NewDecoder(r.Body).Decode(&body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if backupData, ok := sh.data.backups[body.Name]; ok {
			backupData.lastRunTime = time.Now().Unix()
			log.Infof("Received start call for %s", body.Name)
			w.WriteHeader(http.StatusOK)
		} else {
			log.Warningf("Backup with name %s not found", body.Name)
			http.Error(w, "Backup with name not found", http.StatusNotFound)
		}
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func newStartHandler(data *data) http.Handler {
	return startHandler{
		data: data,
	}
}

type finishHandler struct {
	data *data
}

func (sh finishHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		// Declare a new StopBody struct.
		var body StartBody

		// Try to decode the request body into the struct. If there is an error,
		// respond to the client with the error message and a 400 status code.
		err := json.NewDecoder(r.Body).Decode(&body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if backupData, ok := sh.data.backups[body.Name]; ok {
			startTime := backupData.lastRunTime
			backupData.lastRunTime = time.Now().Unix()
			backupData.metrics.lastRunDuration.Set(float64(backupData.lastRunTime - startTime))
			log.Infof("Received finish call for %s", body.Name)
			w.WriteHeader(http.StatusOK)
		} else {
			log.Warningf("Backup with name %s not found", body.Name)
			http.Error(w, "Backup with name not found", http.StatusNotFound)
		}
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func newFinishHandler(data *data) http.Handler {
	return finishHandler{
		data: data,
	}
}
