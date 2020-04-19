// Copyright 2020 Re-Bind Author (Fabrizio Torelli). All rights reserved.
// Use of this source code is governed by a LGPL-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"github.com/hellgate75/rebind/dns"
	"github.com/hellgate75/rebind/log"
	"github.com/hellgate75/rebind/model"
	"github.com/hellgate75/rebind/model/rest"
	"github.com/hellgate75/rebind/utils"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

var rwDirPath string
var configDirPath string
var initializeAndExit bool
var useConfigFile bool
var enableFileLogging bool
var logFilePath string
var logVerbosity string
var enableLogRotate bool
var logMaxFileSize int64
var logMaxFileCount int
var listenIP string
var listenPort int
var dnsPipeIP string
var dnsPipePort int
var dnsPipeResponsePort int
var fwdrsString model.ArgumentsList

var logger = log.NewLogger("re-bind", log.DEBUG)

var defaultForwarders = make([]net.UDPAddr, 0)

func init() {
	logger.Info("Initializing Re-Bind DNS Server ....")
	flag.StringVar(&rwDirPath, "data-dir", rest.DefaultStorageFolder, "dns storage dir")
	flag.StringVar(&configDirPath, "config-dir", rest.DefaultConfigFolder, "dns config dir")
	flag.BoolVar(&initializeAndExit, "init-and-exit", false, "initialize config in the config dir and exit")
	flag.BoolVar(&useConfigFile, "use-config-file", false, "use config file instead parameters")
	flag.BoolVar(&enableFileLogging, "enable-file-log", false, "enable logginf over file")
	flag.StringVar(&logFilePath, "log-file-path", rest.DefaultLogFileFolder, "log file path")
	flag.StringVar(&logVerbosity, "log-verbosity", rest.DefaultLogFileLevel, "log file verbosity level (TRACE, DEBUG, INFO, WARN, ERROR, FATAL)")
	flag.BoolVar(&enableLogRotate, "log-rotate", true, "log file rotation enabled")
	flag.Int64Var(&logMaxFileSize, "log-max-size", 1024, "log file rotation max file size in bytes")
	flag.IntVar(&logMaxFileCount, "log-count", 1024, "log file rotation max number of file")
	flag.StringVar(&listenIP, "listen-ip", rest.DefaultIpAddress, "dns forward ip")
	flag.IntVar(&listenPort, "listen-port", rest.DefaultDnsServerPort, "dns forward port")
	flag.StringVar(&dnsPipeIP, "dns-pipe-ip", rest.DefaultDnsPipeAddress, "tcp dns pipe ip")
	flag.IntVar(&dnsPipePort, "dns-pipe-port", rest.DefaultDnsPipePort, "tcp dns pipe port")
	flag.IntVar(&dnsPipeResponsePort, "dns-pipe-response-port", rest.DefaultDnsAnswerPipePort, "tcp dns pipe responses port")
	flag.Var(&fwdrsString, "forwarder", "Forwarder address in format \"ipv4|ipv6;port;ipv6zone\" (mutliple values)")
}

func main() {
	flag.Parse()
	if utils.StringsListContainItem("-h", flag.Args(), true) ||
		utils.StringsListContainItem("--help", flag.Args(), true) {
		flag.Usage()
		os.Exit(0)
	}
	if initializeAndExit {
		logger.Info("Initialize Re-Bind Dns Server and Exit!!")
		config := model.ReBindConfig{
			DataDirPath:         rwDirPath,
			ConfigDirPath:       configDirPath,
			ListenIP:            listenIP,
			ListenPort:          listenPort,
			DnsPipeIP:           dnsPipeIP,
			DnsPipePort:         dnsPipePort,
			DnsPipeResponsePort: dnsPipeResponsePort,
			EnableFileLogging:   enableFileLogging,
			LogVerbosity:        logVerbosity,
			LogFilePath:         logFilePath,
			LogFileCount:        logMaxFileCount,
			LogMaxFileSize:      logMaxFileSize,
			EnableLogRotate:     enableLogRotate,
		}
		cSErr := model.SaveConfig(configDirPath, "rebind", &config)
		if cSErr != nil {
			logger.Errorf("Unable to save default config to file: ", cSErr)
		}
		os.Exit(0)
	}
	if useConfigFile {
		logger.Warn("Initialize Re-Bind from config file ...")
		logger.Warnf("Re-Bind config folder: %s", configDirPath)
		var config model.ReBindConfig
		cLErr := model.LoadConfig(configDirPath, "rebind", &config)
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
		}
	}
	verbosity := log.LogLevelFromString(logVerbosity)
	logger.Warnf("File logging enabled: %v", enableFileLogging)
	if enableFileLogging {
		logger.Warnf("Enabling file logging at path: %s", logFilePath)
		if _, err := os.Stat(logFilePath); err != nil {
			_ = os.MkdirAll(logFilePath, 0660)
		}
		logDir, _ := os.Open(logFilePath)
		logger.Warnf("File logging at path: %s enabled...", logFilePath)
		var logErr, logRErr error
		var rotator log.LogRotator
		if enableLogRotate {
			logger.Warn("Enable log rotating ...")
			rotator, logRErr = log.NewLogRotator(logDir, "rebind.log", logMaxFileSize, logMaxFileCount, nil)
		} else {
			logger.Warn("No log rotating is enabled...")
			rotator, logRErr = log.NewLogNoRotator(logDir, "rebind.log", nil)
		}
		if logRErr != nil {
			logger.Errorf("Unable to instantiate log rotator: ", logRErr)
		} else {
			logger.Warn("Starting file logging ...")
			logger, logErr = log.NewFileLogger("re-bind",
				rotator,
				verbosity)
			logger.Warn("Log initialization ...")
			if logErr != nil {
				logger.Warn("No File logging started for error...")
				logger = log.NewLogger("re-bind", verbosity)
				logger.Errorf("Unable to instantiate file logger: ", logErr)
			} else {
				logger.Warn("File logging started!!")
			}
		}
	} else {
		logger.Warn("No File logging selected ...")
		logger = log.NewLogger("re-bind", verbosity)
	}
	logger.Info("Starting Re-Bind DNS Server ...")
	if err := os.MkdirAll(rwDirPath, 0666); err != nil {
		logger.Errorf("Create rwdirpath: %v error: %v", rwDirPath, err)
		return
	}
	defaultForwarders = append(defaultForwarders, rest.DefaultGroupForwarders...)
	for _, fw := range fwdrsString {
		list := strings.Split(fw, ";")
		var ip net.IP
		if addr, err := net.ResolveIPAddr("udp", list[0]); err != nil {
			ip = addr.IP
			port := 53
			zone := ""
			if len(list) > 1 {
				p, err := strconv.Atoi(list[1])
				if err == nil {
					port = p
				}
			}
			if len(list) > 2 {
				zone = list[2]
			}
			defaultForwarders = append(defaultForwarders, net.UDPAddr{
				IP:   ip,
				Port: port,
				Zone: zone,
			})
		}
	}
	logger.Infof("Required ip address : %v", listenIP)
	logger.Infof("Required port : %v", listenPort)
	for _, fw := range defaultForwarders {
		logger.Infof("Default forwarder : %s:%v[:%s]", fw.IP, fw.Port, fw.Zone)
	}
	dnsServer := dns.Start(rwDirPath, listenIP, listenPort, dnsPipeIP, dnsPipePort, dnsPipeResponsePort, logger, []net.UDPAddr{{IP: net.ParseIP(listenIP), Port: listenPort}})
	time.Sleep(5 * time.Second)
	logger.Info("Re-Bind DNS Server started!!")
	dnsServer.Wait()
}
