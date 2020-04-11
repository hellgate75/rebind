// Copyright 2020 Re-Bind Author (Fabrizio Torelli). All rights reserved.
// Use of this source code is governed by a LGPL-style
// license that can be found in the LICENSE file.

package model

import (
	"net"
)

const (
	DefaultStorageFolder string = "/Users/Fabrizio/var/dns"
	//DefaultStorageFolder string = "/var/dns"
	DefaultDnsServerPort  = 53
	DefaultRestServerPort = 9000
	DefaultIpAddress      = "0.0.0.0"
	DefaultDnsPipeAddress = "127.0.0.1"
	DefaultDnsPipePort    = 953
)

var (
	DefaultGroupForwarders = []net.UDPAddr{
		net.UDPAddr{
			IP:   net.IPv4(8, 8, 8, 8),
			Port: 53,
		},
		net.UDPAddr{
			IP:   net.IPv4(8, 8, 4, 4),
			Port: 53,
		},
	}
)

type Response struct {
	Status  int         `yaml:"status" json:"status" xml:"status"`
	Message string      `yaml:"message" json:"message" xml:"message"`
	Data    interface{} `yaml:"data" json:"data" xml:"data"`
}

type GroupRequest struct {
	Name       string        `yaml:"name" json:"name" xml:"name"`
	Forwarders []net.UDPAddr `yaml:"fowarders" json:"fowarders" xml:"fowarders"`
	Domains    []string      `yaml:"domains,omitempty" json:"domains,omitempty" xml:"domains,omitempty"`
}

type GroupFilterRequest struct {
	Filter struct {
		Name   string `yaml:"name,omitempty" json:"name,omitempty" xml:"name,omitempty"`
		Domain string `yaml:"domain,omitempty" json:"domain,omitempty" xml:"domain,omitempty"`
	}
}

type Request struct {
	Host    string     `yaml:"host" json:"host" xml:"host"`
	TTL     uint32     `yaml:"ttl,omitempty" json:"ttl,omitempty" xml:"ttl,omitempty"`
	Type    string     `yaml:"type" json:"type" xml:"type"`
	Data    string     `yaml:"data" json:"data" xml:"data"`
	OldData string     `yaml:"oldData,omitempty" json:"oldData,omitempty" xml:"old-data,omitempty"`
	SOA     RequestSOA `yaml:"soaData,omitempty" json:"soaData,omitempty" xml:"soa-data,omitempty"`
	OldSOA  RequestSOA `yaml:"oldSoaData,omitempty" json:"oldSoaData,omitempty" xml:"old-soa-data,omitempty"`
	MX      RequestMX  `yaml:"mxData,omitempty" json:"mxData,omitempty" xml:"mx-data,omitempty"`
	OldMX   RequestMX  `yaml:"oldMxData,omitempty" json:"oldMxData,omitempty" xml:"old-mx-data,omitempty"`
	SRV     RequestSRV `yaml:"srvData,omitempty" json:"srvData,omitempty" xml:"srv-data,omitempty"`
	OldSRV  RequestSRV `yaml:"oldSrvData,omitempty" json:"oldSrvData,omitempty" xml:"old-srv-data,omitempty"`
}

type RequestSOA struct {
	NS      string `yaml:"ns,omitempty" json:"ns,omitempty" xml:"ns,omitempty"`
	MBox    string `yaml:"mBox,omitempty" json:"mBox,omitempty" xml:"mbox,omitempty"`
	Serial  uint32 `yaml:"serial,omitempty" json:"serial,omitempty" xml:"serial,omitempty"`
	Refresh uint32 `yaml:"refresh,omitempty" json:"refresh,omitempty" xml:"refresh,omitempty"`
	Retry   uint32 `yaml:"retry,omitempty" json:"retry,omitempty" xml:"retry,omitempty"`
	Expire  uint32 `yaml:"expire,omitempty" json:"expire,omitempty" xml:"expire,omitempty"`
	MinTTL  uint32 `yaml:"minTTL,omitempty" json:"minTTL,omitempty" xml:"min-ttl,omitempty"`
}

type RequestMX struct {
	Pref uint16 `yaml:"pref,omitempty" json:"pref,omitempty" xml:"pref,omitempty"`
	MX   string `yaml:"mx,omitempty" json:"mx,omitempty" xml:"mx,omitempty"`
}

type RequestSRV struct {
	Priority uint16 `yaml:"priority,omitempty" json:"priority,omitempty" xml:"priority,omitempty"`
	Weight   uint16 `yaml:"weight,omitempty" json:"weight,omitempty" xml:"weight,omitempty"`
	Port     uint16 `yaml:"port,omitempty" json:"port,omitempty" xml:"port,omitempty"`
	Target   string `yaml:"target,omitempty" json:"target,omitempty" xml:"target,omitempty"`
}

type Get struct {
	Host string
	TTL  uint32
	Type string
	Data string
}
