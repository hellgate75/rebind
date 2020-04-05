// Copyright 2020 Re-Bind Author (Fabrizio Torelli). All rights reserved.
// Use of this source code is governed by a LGPL-style
// license that can be found in the LICENSE file.

package model

type Request struct {
	Host    string
	TTL     uint32
	Type    string
	Data    string
	OldData string
	SOA     RequestSOA
	OldSOA  RequestSOA
	MX      RequestMX
	OldMX   RequestMX
	SRV     RequestSRV
	OldSRV  RequestSRV
}

type RequestSOA struct {
	NS      string
	MBox    string
	Serial  uint32
	Refresh uint32
	Retry   uint32
	Expire  uint32
	MinTTL  uint32
}

type RequestMX struct {
	Pref uint16
	MX   string
}

type RequestSRV struct {
	Priority uint16
	Weight   uint16
	Port     uint16
	Target   string
}

type Get struct {
	Host string
	TTL  uint32
	Type string
	Data string
}

