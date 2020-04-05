// Copyright 2020 Re-Bind Author (Fabrizio Torelli). All rights reserved.
// Use of this source code is governed by a LGPL-style
// license that can be found in the LICENSE file.

package model

import (
	"golang.org/x/net/dns/dnsmessage"
	"net"
)

// DNSServer will do Listen, Query and Send.
type DNSServer interface {
	Listen(ipAddress string, port int) error
	Query(Packet)
	GetService() DNSService
	Wait()
}

type DNSService interface {
	Save(key string, resource dnsmessage.Resource, old *dnsmessage.Resource) bool
	SaveBulk(key string, resources []dnsmessage.Resource)
	All() []Get
	Remove(key string, r *dnsmessage.Resource) bool
}

// Packet carries DNS packet payload and sender address.
type Packet struct {
	Addr    net.UDPAddr
	Message dnsmessage.Message
}

