// Copyright 2020 Re-Bind Author (Fabrizio Torelli). All rights reserved.
// Use of this source code is governed by a LGPL-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"github.com/hellgate75/rebind/dns"
	"github.com/hellgate75/rebind/log"
	"github.com/hellgate75/rebind/model"
	"github.com/hellgate75/rebind/utils"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

var defaultFolder string = "/Users/Fabrizio/var/dns"

//var defaultFolder string = "/var/dns"
var defaultIpAddress = "0.0.0.0"

//var defaultIpAddress = "8.8.8.8"

var rwDirPath string
var listenIP string
var listenPort int
var dnsPipeIP string
var dnsPipePort int
var fwdrsString model.ListArgument

const (
	internalListenPort int = 953
	internalDialPort   int = 954
)

//TODO: Give Life to File Logger
//var logger log.Logger = log.NewLogger("re-bind", log.INFO)
var logger log.Logger = log.NewLogger("re-bind", log.DEBUG)

var defaultForwarders []net.UDPAddr = []net.UDPAddr{
	net.UDPAddr{
		IP:   net.IPv4(8, 8, 8, 8),
		Port: 53,
	},
	net.UDPAddr{
		IP:   net.IPv4(8, 8, 4, 4),
		Port: 53,
	},
}

func init() {
	logger.Info("Initializing Re-Bind ....")
	flag.StringVar(&rwDirPath, "rwdir", defaultFolder, "dns storage dir")
	flag.StringVar(&listenIP, "listen-ip", defaultIpAddress, "dns forward ip")
	flag.IntVar(&listenPort, "listen-port", 53, "dns forward port")
	flag.StringVar(&dnsPipeIP, "dns-pipe-ip", "127.0.0.1", "tcp dns pipe ip")
	flag.IntVar(&dnsPipePort, "dns-pipe-port", 953, "tcp dns pipe port")
	flag.Var(&fwdrsString, "forwarder", "Forwarder address in format \"ipv4|ipv6;port;ipv6zone\" (mutliple values)")
}

func main() {
	flag.Parse()
	if utils.StringsListContainItem("-h", flag.Args(), true) ||
		utils.StringsListContainItem("--help", flag.Args(), true) {
		flag.Usage()
		os.Exit(0)
	}
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
	if err := os.MkdirAll(rwDirPath, 0666); err != nil {
		logger.Errorf("create rwdirpath: %v error: %v", rwDirPath, err)
		return
	}
	logger.Info("Starting re-bind dns server ...")
	logger.Infof("Required ip address : %v", listenIP)
	logger.Infof("Required port : %v", listenPort)
	for _, fw := range defaultForwarders {
		logger.Infof("Default forwarder : %s:%v[:%s]", fw.IP, fw.Port, fw.Zone)
	}
	dnsServer := dns.Start(rwDirPath, listenIP, listenPort, dnsPipeIP, dnsPipePort, logger, []net.UDPAddr{{IP: net.ParseIP(listenIP), Port: listenPort}})
	time.Sleep(5 * time.Second)
	dnsServer.Wait()
}
