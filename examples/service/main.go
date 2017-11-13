/**
 * @license
 * Copyright 2017 Telefónica Investigación y Desarrollo, S.A.U
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"encoding/json"
	"flag"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/telefonica/govice"
)

type config struct {
	Address  string `json:"address" env:"ADDRESS"`
	BasePath string `json:"basePath" env:"BASE_PATH"`
	LogLevel string `json:"logLevel" env:"LOG_LEVEL"`
}

func withMws(op string) func(http.HandlerFunc) http.HandlerFunc {
	ctxt := &govice.LogContext{
		Service:   "demo",
		Operation: op,
	}
	return func(next http.HandlerFunc) http.HandlerFunc {
		return govice.WithLogContext(ctxt)(govice.WithLog(next))
	}
}

func main() {
	// Prepare logger
	time.Local = time.UTC
	logContext := govice.LogContext{
		Service:   "demo",
		Operation: "init",
	}
	logger := govice.NewLogger()
	logger.SetLogContext(&logContext)
	alarmContext := &govice.LogContext{Alarm: "ALARM_INIT"}

	// Prepare the configuration
	cfgFile := flag.String("config", "./config.json", "path to config file")
	flag.Parse()
	var cfg config
	if err := govice.GetConfig(*cfgFile, &cfg); err != nil {
		logger.FatalC(alarmContext, "Bad configuration with file '%s'. %s", *cfgFile, err)
		os.Exit(1)
	}
	logger.SetLevel(cfg.LogLevel)
	govice.SetDefaultLogLevel(cfg.LogLevel)

	// Log the configuration
	if configBytes, err := json.Marshal(cfg); err == nil {
		logger.Info("Configuration: %s", string(configBytes))
	}

	// Create the validator and validate the configuration
	validator := govice.NewValidator()
	if err := validator.LoadSchemas("schemas"); err != nil {
		logger.FatalC(alarmContext, "Error loading JSON schemas for validator. %s", err)
		os.Exit(1)
	}
	if err := validator.ValidateConfig("config", &cfg); err != nil {
		logger.FatalC(alarmContext, "Bad configuration according to JSON schema. %s", err)
		os.Exit(1)
	}

	// Create the logic of the service
	u := NewUsersService(validator)

	// Create the router (based on mux)
	r := mux.NewRouter()
	r.HandleFunc("/users", withMws("createUser")(u.CreateUser)).Methods("POST")
	r.HandleFunc("/users/{login}", withMws("getUser")(u.GetUser)).Methods("GET")
	r.HandleFunc("/users/{login}", withMws("deleteUser")(u.DeleteUser)).Methods("DELETE")

	// Launch the HTTP server
	s := &http.Server{Addr: cfg.Address, Handler: r}
	if err := s.ListenAndServe(); err != nil {
		logger.FatalC(alarmContext, "Error launching the server. %s", err)
		os.Exit(1)
	}
}
