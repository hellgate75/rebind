// Copyright 2020 Re-Bind Author (Fabrizio Torelli). All rights reserved.
// Use of this source code is governed by a LGPL-style
// license that can be found in the LICENSE file.

package model

import (
	"golang.org/x/net/dns/dnsmessage"
	"net"
	"time"
)

// DNSServer will do Listen, Query and Send.
type DNSServer interface {
	Listen(ipAddress string, port int, pipeAddress string, pipePort int, pipeResponsePort int) error
	Query(Packet)
	GetService() DNSService
	Wait()
}

type DNSService interface {
	Save(key string, resource dnsmessage.Resource, addr net.IPAddr, recordData string, old *dnsmessage.Resource) bool
	SaveBulk(key string, resources []dnsmessage.Resource)
	All() []Get
	Remove(key string, r *dnsmessage.Resource) bool
}

// Packet carries DNS packet payload and sender address.
type Packet struct {
	Addr    net.UDPAddr
	Message dnsmessage.Message
}

// Collects indformation about any answer
type AnswerBlock struct {
	Created time.Time
	TTL     time.Duration
	Answer  []dnsmessage.Resource
}

func (answer *AnswerBlock) IsValid() bool {
	return int64((time.Now().Nanosecond() - answer.Created.Nanosecond())) >= int64(answer.TTL.Nanoseconds())

}

type Get struct {
	Host string
	TTL  uint32
	Type string
	Data string
}
type Request struct {
	Host    string     `yaml:"host" json:"host" xml:"host"`
	TTL     uint32     `yaml:"ttl" json:"ttl" xml:"ttl"`
	Type    string     `yaml:"type" json:"type" xml:"type"`
	Data    string     `yaml:"data" json:"data" xml:"data"`
	OldData string     `yaml:"oldData" json:"oldData" xml:"old-data"`
	SOA     RequestSOA `yaml:"soaData" json:"soaData" xml:"soa-data"`
	OldSOA  RequestSOA `yaml:"oldSoaData" json:"oldSoaData" xml:"old-soa-data"`
	MX      RequestMX  `yaml:"mxData" json:"mxData" xml:"mx-data"`
	OldMX   RequestMX  `yaml:"oldMxData" json:"oldMxData" xml:"old-mx-data"`
	SRV     RequestSRV `yaml:"srvData" json:"srvData" xml:"srv-data"`
	OldSRV  RequestSRV `yaml:"oldSrvData" json:"oldSrvData" xml:"old-srv-data"`
}

type RequestSOA struct {
	NS      string `yaml:"ns" json:"ns" xml:"ns"`
	MBox    string `yaml:"mBox" json:"mBox" xml:"mbox"`
	Serial  uint32 `yaml:"serial" json:"serial" xml:"serial"`
	Refresh uint32 `yaml:"refresh" json:"refresh" xml:"refresh"`
	Retry   uint32 `yaml:"retry" json:"retry" xml:"retry"`
	Expire  uint32 `yaml:"expire" json:"expire" xml:"expire"`
	MinTTL  uint32 `yaml:"minTTL" json:"minTTL" xml:"min-ttl"`
}

type RequestMX struct {
	Pref uint16 `yaml:"pref" json:"pref" xml:"pref"`
	MX   string `yaml:"mx" json:"mx" xml:"mx"`
}

type RequestSRV struct {
	Priority uint16 `yaml:"priority" json:"priority" xml:"priority"`
	Weight   uint16 `yaml:"weight" json:"weight" xml:"weight"`
	Port     uint16 `yaml:"port" json:"port" xml:"port"`
	Target   string `yaml:"target" json:"target" xml:"target"`
}

type Response struct {
	Status  int         `yaml:"status" json:"status" xml:"status"`
	Message string      `yaml:"message" json:"message" xml:"message"`
	Data    interface{} `yaml:"data" json:"data" xml:"data"`
}
