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
var listenIP string
var listenPort int
var dnsPipeIP string
var dnsPipePort int
var dnspipeResponsePort int
var fwdrsString model.ListArgument

var logger = log.NewLogger("re-bind", log.DEBUG)

var defaultForwarders = make([]net.UDPAddr, 0)

func init() {
	logger.Info("Initializing Re-Bind DNS Server ....")
	flag.StringVar(&rwDirPath, "rwdir", rest.DefaultStorageFolder, "dns storage dir")
	flag.StringVar(&listenIP, "listen-ip", rest.DefaultIpAddress, "dns forward ip")
	flag.IntVar(&listenPort, "listen-port", rest.DefaultDnsServerPort, "dns forward port")
	flag.StringVar(&dnsPipeIP, "dns-pipe-ip", rest.DefaultDnsPipeAddress, "tcp dns pipe ip")
	flag.IntVar(&dnsPipePort, "dns-pipe-port", rest.DefaultDnsPipePort, "tcp dns pipe port")
	flag.IntVar(&dnspipeResponsePort, "dns-pipe-response-port", rest.DefaultDnsAnswerPipePort, "tcp dns pipe responses port")
	flag.Var(&fwdrsString, "forwarder", "Forwarder address in format \"ipv4|ipv6;port;ipv6zone\" (mutliple values)")
}

func main() {
	logger.Info("Starting Re-Bind DNS Server ...")
	flag.Parse()
	if utils.StringsListContainItem("-h", flag.Args(), true) ||
		utils.StringsListContainItem("--help", flag.Args(), true) {
		flag.Usage()
		os.Exit(0)
	}
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
	dnsServer := dns.Start(rwDirPath, listenIP, listenPort, dnsPipeIP, dnsPipePort, dnspipeResponsePort, logger, []net.UDPAddr{{IP: net.ParseIP(listenIP), Port: listenPort}})
	time.Sleep(5 * time.Second)
	dnsServer.Wait()
}
