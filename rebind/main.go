// Copyright 2020 Re-Bind Author (Fabrizio Torelli). All rights reserved.
// Use of this source code is governed by a LGPL-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"github.com/hellgate75/rebind/dns"
	"github.com/hellgate75/rebind/log"
	"net"
	"os"
	"strings"
	"time"
)
var defaultFolder string = "/Users/Fabrizio/var/dns"
//var defaultFolder string = "/var/dns"
var defaultIpAddress = "0.0.0.0"
//var defaultIpAddress = "8.8.8.8"

var rwDirPath  string
var listenIP   string
var listenPort int

const(
	internalListenPort int = 953
	internalDialPort int = 954
)
//TODO: Give Life to File Logger
//var logger log.Logger = log.NewLogger("re-bind", log.INFO)
var logger log.Logger = log.NewLogger("re-bind", log.DEBUG)

func init() {
	logger.Info("Initializing Re-Bind ....")
	flag.StringVar(&rwDirPath,"rwdir",defaultFolder,"dns storage dir")
	flag.StringVar(&listenIP,"listen-ip", defaultIpAddress, "dns forward ip")
	flag.IntVar(&listenPort, "listen-port", 53, "dns forward port")
}

func main() {
	flag.Parse()
	if strings.Contains(strings.Join(flag.Args(), "") , "-h") ||
		strings.Contains(strings.Join(flag.Args(), "") , "--help"){
		flag.Usage()
		os.Exit(0)
	}
	if err := os.MkdirAll(rwDirPath, 0666); err != nil {
		logger.Errorf("create rwdirpath: %v error: %v", rwDirPath, err)
		return
	}
	logger.Info("Starting re-bind ...")
	logger.Infof("Required ip address : %v", listenIP)
	logger.Infof("Required port : %v", listenPort)
	dnsServer := dns.Start(rwDirPath, listenIP, listenPort, logger, []net.UDPAddr{{IP: net.ParseIP(listenIP), Port: listenPort}})
	time.Sleep(5 * time.Second)
	dnsServer.Wait()
}