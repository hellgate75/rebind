// Copyright 2020 Re-Bind Author (Fabrizio Torelli). All rights reserved.
// Use of this source code is governed by a LGPL-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/hellgate75/rebind/log"
	"github.com/hellgate75/rebind/model"
	"github.com/hellgate75/rebind/model/rest"
	pnet "github.com/hellgate75/rebind/net"
	"github.com/hellgate75/rebind/registry"
	"github.com/hellgate75/rebind/rest/services"
	"github.com/hellgate75/rebind/utils"
	"net"
	"net/http"
	"os"
)

var rwDirPath string
var configDirPath string
var initializeAndExit bool
var useConfigFile bool
var enableFileLogging bool
var logVerbosity string
var logFilePath string
var enableLogRotate bool
var logMaxFileSize int64
var logMaxFileCount int
var listenIP string
var listenPort int
var dnsPipeIP string
var dnsPipePort int
var dnsPipeResponsePort int
var tlsCert string
var tlsKey string

//TODO: Give Life to Logger
var logger log.Logger = log.NewLogger("re-web", log.DEBUG)

var defaultForwarders = make([]net.UDPAddr, 0)

func init() {
	logger.Info("Initializing Re-Web Rest Server ....")
	flag.StringVar(&rwDirPath, "data-dir", rest.DefaultStorageFolder, "dns storage dir")
	flag.StringVar(&configDirPath, "config-dir", rest.DefaultConfigFolder, "dns config dir")
	flag.BoolVar(&initializeAndExit, "init-and-exit", false, "initialize config in the config dir and exit")
	flag.BoolVar(&useConfigFile, "use-config-file", false, "use config file instead parameters")
	flag.BoolVar(&enableFileLogging, "enable-file-log", false, "enable logginf over file")
	flag.StringVar(&logVerbosity, "log-verbosity", rest.DefaultLogFileLevel, "log file verbosity level (TRACE, DEBUG, INFO, WARN, ERROR, FATAL)")
	flag.StringVar(&logFilePath, "log-file-path", rest.DefaultLogFileFolder, "log file path")
	flag.BoolVar(&enableLogRotate, "log-rotate", true, "log file rotation enabled")
	flag.Int64Var(&logMaxFileSize, "log-max-size", 1024, "log file rotation max file size in bytes")
	flag.IntVar(&logMaxFileCount, "log-count", 1024, "log file rotation max number of file")
	flag.StringVar(&listenIP, "listen-ip", rest.DefaultIpAddress, "http server ip")
	flag.IntVar(&listenPort, "listen-port", rest.DefaultRestServerPort, "http server port")
	flag.StringVar(&dnsPipeIP, "dns-pipe-ip", rest.DefaultDnsPipeAddress, "tcp dns pipe ip")
	flag.IntVar(&dnsPipePort, "dns-pipe-port", rest.DefaultDnsAnswerPipePort, "tcp dns pipe port")
	flag.IntVar(&dnsPipeResponsePort, "dns-pipe-response-port", rest.DefaultDnsPipePort, "tcp dns pipe responses port")
	flag.StringVar(&tlsCert, "tsl-cert", "", "tls certificate file path")
	flag.StringVar(&tlsKey, "tsl-key", "", "tls certificate key file path")
}

func main() {
	flag.Parse()
	if utils.StringsListContainItem("-h", flag.Args(), true) ||
		utils.StringsListContainItem("--help", flag.Args(), true) {
		flag.Usage()
		os.Exit(0)
	}
	if initializeAndExit {
		logger.Info("Initialize Re-Web Rest Server and Exit!!")
		config := model.ReWebConfig{
			DataDirPath:         rwDirPath,
			ConfigDirPath:       configDirPath,
			ListenIP:            listenIP,
			ListenPort:          listenPort,
			DnsPipeIP:           dnsPipeIP,
			DnsPipePort:         dnsPipePort,
			DnsPipeResponsePort: dnsPipeResponsePort,
			TlsCert:             tlsCert,
			TlsKey:              tlsKey,
			EnableFileLogging:   enableFileLogging,
			LogVerbosity:        logVerbosity,
			LogFilePath:         logFilePath,
			LogFileCount:        logMaxFileCount,
			LogMaxFileSize:      logMaxFileSize,
			EnableLogRotate:     enableLogRotate,
		}
		cSErr := model.SaveConfig(configDirPath, "reweb", &config)
		if cSErr != nil {
			logger.Errorf("Unable to save default config to file: ", cSErr)
		}
		os.Exit(0)
	}
	if useConfigFile {
		logger.Warn("Initialize Re-Web from config file ...")
		logger.Warnf("Re-Web config folder: %s", configDirPath)
		var config model.ReWebConfig
		cLErr := model.LoadConfig(configDirPath, "reweb", &config)
		if cLErr != nil {
			logger.Errorf("Unable to load default config from file: ", cLErr)
		} else {
			logger.Warnf("Loading configuration from file complete!!", config)
			logger.Debugf("Configuration: %v", config)
			rwDirPath = config.DataDirPath
			configDirPath = config.ConfigDirPath
			listenIP = config.ListenIP
			listenPort = config.ListenPort
			dnsPipeIP = config.DnsPipeIP
			dnsPipePort = config.DnsPipePort
			dnsPipeResponsePort = config.DnsPipeResponsePort
			enableFileLogging = config.EnableFileLogging
			logVerbosity = config.LogVerbosity
			logFilePath = config.LogFilePath
			logMaxFileCount = config.LogFileCount
			logMaxFileSize = config.LogMaxFileSize
			enableLogRotate = config.EnableLogRotate
			tlsCert = config.TlsCert
			tlsKey = config.TlsKey
		}
	}
	verbosity := log.LogLevelFromString(logVerbosity)
	logger.Warnf("File logging enabled: %v", enableFileLogging)
	if enableFileLogging {
		logger.Warnf("Enabling file logging at path: %s", logFilePath)
		if _, err := os.Stat(logFilePath); err != nil {
			_ = os.MkdirAll(logFilePath, 0660)
		}
		logger.Warnf("File logging at path: %s enabled...", logFilePath)
		logDir, _ := os.Open(logFilePath)
		var logErr, logRErr error
		var rotator log.LogRotator
		if enableLogRotate {
			logger.Warn("Enable log rotating ...")
			rotator, logRErr = log.NewLogRotator(logDir, "reweb.log", logMaxFileSize, logMaxFileCount, nil)
		} else {
			logger.Warn("No log rotating is enabled...")
			rotator, logRErr = log.NewLogNoRotator(logDir, "reweb.log", nil)
		}
		if logRErr != nil {
			logger.Errorf("Unable to instantiate log rotator: ", logRErr)
		} else {
			logger.Warn("Starting file logging ...")
			logger, logErr = log.NewFileLogger("re-web",
				rotator,
				verbosity)
			if logErr != nil {
				logger.Warn("No File logging started for error...")
				logger = log.NewLogger("re-web", verbosity)
				logger.Errorf("Unable to instantiate file logger: ", logErr)
			} else {
				logger.Warn("File logging started!!")
			}
		}
	} else {
		logger.Warn("No File logging selected ...")
		logger = log.NewLogger("re-web", verbosity)
	}
	logger.Info("Starting Re-Web Rest Server ...")
	if err := os.MkdirAll(rwDirPath, 0666); err != nil {
		logger.Errorf("Create rwdirpath: %v error: %v", rwDirPath, err)
		return
	}

	// Improve list of default group forwarders if provided
	defaultForwarders = append(defaultForwarders, rest.DefaultGroupForwarders...)

	// Create network Pipe Stream with the dns server
	pipe, err := pnet.NewInputOutputPipeWith(dnsPipeIP, dnsPipePort, dnsPipeIP, dnsPipeResponsePort, nil, logger)
	if err != nil {
		logger.Fatalf("Unable to create NetPipe in listen: %v and bind: %v/%v\n", dnsPipePort, dnsPipeResponsePort)
	}
	// Create Data Store
	store := registry.NewStore(logger, rwDirPath, defaultForwarders)
	store.Load()

	// Handler stuf for the API service groups
	dnsHandler := func(serv services.RestService) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodPost:
				serv.Create(w, r)
			case http.MethodGet:
				serv.Read(w, r)
			case http.MethodPut:
				serv.Update(w, r)
			case http.MethodDelete:
				serv.Delete(w, r)
			}
		}
	}

	withAuth := func(h http.HandlerFunc) http.HandlerFunc {
		// authentication intercepting
		var _ = "intercept"
		return func(w http.ResponseWriter, r *http.Request) {
			h(w, r)
		}
	}

	rtr := mux.NewRouter()
	var proto string = "http"
	if tlsCert != "" && tlsKey != "" {
		proto = "https"
	}
	// Creates/Sets API endpoints handlers
	services.CreateApiEndpoints(rtr, withAuth, dnsHandler,
		pipe, store, logger, fmt.Sprintf("%s://%s:%v", proto, listenIP, listenPort))

	//Adding entry point for generic queries (GET)
	http.Handle("/", rtr)

	// Adding TLS certificates if required
	if tlsCert == "" || tlsKey == "" {
		logger.Infof("RestService start-up:: Starting server in simple mode on ip: %s and port: %v\n", listenIP, listenPort)
		err = http.ListenAndServe(fmt.Sprintf("%s:%v", listenIP, listenPort), nil)
	} else {
		logger.Infof("RestService start-up:: Starting server in simple mode on ip: %s and port: %v\n", listenIP, listenPort)
		logger.Infof("RestService start-up:: Using certificate file: %s and certticate key file: %v..\n", tlsCert, tlsKey)
		err = http.ListenAndServeTLS(fmt.Sprintf("%s:%v", listenIP, listenPort), tlsCert, tlsKey, nil)
	}
	if err != nil {
		logger.Fatalf("RestService start-up:: Error listening on s:%v - Error: %v\n", listenIP, listenPort, err)
		os.Exit(1)
	}
	logger.Info("Re-Web DNS Server started!!")
}
