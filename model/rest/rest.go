// Copyright 2020 Re-Bind Author (Fabrizio Torelli). All rights reserved.
// Use of this source code is governed by a LGPL-style
// license that can be found in the LICENSE file.

package rest

import (
	"github.com/hellgate75/rebind/data"
	"github.com/hellgate75/rebind/model"
	"github.com/hellgate75/rebind/store"
	"net"
	"strings"
)

const (
	DefaultStorageFolder     string = "/var/rebind"
	DefaultDnsServerPort            = 53
	DefaultRestServerPort           = 9000
	DefaultIpAddress                = "0.0.0.0"
	DefaultDnsPipeAddress           = "127.0.0.1"
	DefaultDnsPipePort              = 953
	DefaultDnsAnswerPipePort        = 954
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

type GroupRequest struct {
	Name       string        `yaml:"name" json:"name" xml:"name"`
	Forwarders []net.UDPAddr `yaml:"fowarders" json:"fowarders" xml:"fowarders"`
	Domains    []string      `yaml:"domains" json:"domains" xml:"domains"`
}

type GroupFilterRequest struct {
	Filter struct {
		Name   string `yaml:"name" json:"name" xml:"name"`
		Domain string `yaml:"domain" json:"domain" xml:"domain"`
	}
}

type Action string

const (
	AddResoource    Action = "ADD"
	UpdateResoource Action = "UPDATE"
	DeleteResoource Action = "DELTE"
)

func (a Action) Equals(act Action) bool {
	return string(act) != "" && strings.ToUpper(string(act)) == strings.ToUpper(string(a))
}

func (a Action) Same(act string) bool {
	return act != "" && strings.ToUpper(act) == strings.ToUpper(string(a))
}

func (a Action) String(act string) string {
	return strings.ToUpper(string(a))
}

type Field string

func (f Field) Equals(field Field) bool {
	return string(field) != "" && strings.ToUpper(string(field)) == strings.ToUpper(string(f))
}

type DnsGroupResponse struct {
	Group     data.Group        `yaml:"group" json:"group" xml:"group"`
	Resources []store.DNSRecord `yaml:"resources" json:"resources" xml:"resources"`
}

type DnsUpdateRequest struct {
	Action Action            `yaml:"action" json:"action" xml:"action"`
	Field  string            `yaml:"field" json:"field" xml:"field"`
	Data   UpdateRequestForm `yaml:"data" json:"data" xml:"data"`
}

type UpdateListForm struct {
	Value string `yaml:"value" json:"value" xml:"value"`
	Index int    `yaml:"index" json:"index" xml:"index"`
}

type UpdateRequestForm struct {
	ListData   UpdateListForm `yaml:"fromList" json:"fromList" xml:"from-list"`
	RecordData model.Request  `yaml:"fromRecord" json:"fromRecord" xml:"from-record"`
	NewValue   interface{}    `yaml:"value" json:"value" xml:"value"`
}

type GroupCreationRequest struct {
	Forwarders []net.UDPAddr `yaml:"fowarders" json:"fowarders" xml:"fowarders"`
	Domains    []string      `yaml:"domains" json:"domains" xml:"domains"`
}

type DnsGroupsResponse struct {
	Groups []data.Group `yaml:"groups" json:"groups" xml:"groups"`
}

type DnsTemplateDataType struct {
	Method  string      `yaml:"method" json:"method" xml:"method"`
	Header  []string    `yaml:"header" json:"header" xml:"header"`
	Query   []string    `yaml:"query" json:"query" xml:"query"`
	Request interface{} `yaml:"request" json:"request" xml:"request"`
}
type DnsTemplateResponse struct {
	Templates []DnsTemplateDataType `yaml:"template" json:"template" xml:"template"`
}
