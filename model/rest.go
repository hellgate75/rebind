// Copyright 2020 Re-Bind Author (Fabrizio Torelli). All rights reserved.
// Use of this source code is governed by a LGPL-style
// license that can be found in the LICENSE file.

package model

type Request struct {
	Host    string     `yaml:"host" json:"host" xml:"host,chardata"`
	TTL     uint32     `yaml:"ttl,omitempty" json:"ttl,omitempty" xml:"ttl,chardata,omitempty"`
	Type    string     `yaml:"type" json:"type" xml:"type,chardata"`
	Data    string     `yaml:"data" json:"data" xml:"data,chardata"`
	OldData string     `yaml:"oldData,omitempty" json:"oldData,omitempty" xml:"old-data,chardata,omitempty"`
	SOA     RequestSOA `yaml:"soaData,omitempty" json:"soaData,omitempty" xml:"soa-data,chardata,omitempty"`
	OldSOA  RequestSOA `yaml:"oldSoaData,omitempty" json:"oldSoaData,omitempty" xml:"old-soa-data,chardata,omitempty"`
	MX      RequestMX  `yaml:"mxData,omitempty" json:"mxData,omitempty" xml:"mx-data,chardata,omitempty"`
	OldMX   RequestMX  `yaml:"oldMxData,omitempty" json:"oldMxData,omitempty" xml:"old-mx-data,chardata,omitempty"`
	SRV     RequestSRV `yaml:"srvData,omitempty" json:"srvData,omitempty" xml:"srv-data,chardata,omitempty"`
	OldSRV  RequestSRV `yaml:"oldSrvData,omitempty" json:"oldSrvData,omitempty" xml:"old-srv-data,chardata,omitempty"`
}

type RequestSOA struct {
	NS      string `yaml:"ns,omitempty" json:"ns,omitempty" xml:"ns,chardata,omitempty"`
	MBox    string `yaml:"mBox,omitempty" json:"mBox,omitempty" xml:"mbox,chardata,omitempty"`
	Serial  uint32 `yaml:"serial,omitempty" json:"serial,omitempty" xml:"serial,chardata,omitempty"`
	Refresh uint32 `yaml:"refresh,omitempty" json:"refresh,omitempty" xml:"refresh,chardata,omitempty"`
	Retry   uint32 `yaml:"retry,omitempty" json:"retry,omitempty" xml:"retry,chardata,omitempty"`
	Expire  uint32 `yaml:"expire,omitempty" json:"expire,omitempty" xml:"expire,chardata,omitempty"`
	MinTTL  uint32 `yaml:"minTTL,omitempty" json:"minTTL,omitempty" xml:"min-ttl,chardata,omitempty"`
}

type RequestMX struct {
	Pref uint16 `yaml:"pref,omitempty" json:"pref,omitempty" xml:"pref,chardata,omitempty"`
	MX   string `yaml:"mx,omitempty" json:"mx,omitempty" xml:"mx,chardata,omitempty"`
}

type RequestSRV struct {
	Priority uint16 `yaml:"priority,omitempty" json:"priority,omitempty" xml:"priority,chardata,omitempty"`
	Weight   uint16 `yaml:"weight,omitempty" json:"weight,omitempty" xml:"weight,chardata,omitempty"`
	Port     uint16 `yaml:"port,omitempty" json:"port,omitempty" xml:"port,chardata,omitempty"`
	Target   string `yaml:"target,omitempty" json:"target,omitempty" xml:"target,chardata,omitempty"`
}

type Get struct {
	Host string
	TTL  uint32
	Type string
	Data string
}
