// Copyright 2020 Re-Bind Author (Fabrizio Torelli). All rights reserved.
// Use of this source code is governed by a LGPL-style
// license that can be found in the LICENSE file.

package utils

import (
	"errors"
	"github.com/hellgate75/rebind/model"
	"golang.org/x/net/dns/dnsmessage"
	"net"
	"regexp"
	"strings"
)

var (
	ErrTypeNotSupport = errors.New("type not support")
	ErrIPInvalid      = errors.New("invalid IP address")
)

func StringsListContainItem(elem string, elems []string, ignoreCase bool) bool {
	if ignoreCase {
		elemLower := strings.ToLower(elem)
		for _, item := range elems {
			if ignoreCase && strings.ToLower(item) == elemLower {
				return true
			}
		}
	} else {
		for _, item := range elems {
			if item == elem {
				return true
			}
		}
	}
	return false
}

func GenericListContainItem(elem interface{}, elems []interface{}) bool {
	for _, item := range elems {
		if item == elem {
			return true
		}
	}
	return false
}

func ReplaceQuestionUnrelated(val string) (string, error) {
	expr, err := regexp.Compile("[^a-zA-Z0-9.-]+")
	if err != nil {
		return val, err
	}
	value := expr.ReplaceAllString(val, "")
	if value[len(value)-1:] == "." {
		value = value[:len(value)-1]
	}
	return value, nil
}

func ToRecordData(req model.Request) (string, string, net.IP, string, dnsmessage.Resource, error) {
	rName, err := dnsmessage.NewName(req.Host)
	rData := ""

	var udpAddr net.IP = nil
	none := dnsmessage.Resource{}
	if err != nil {
		return req.Host, req.Type, udpAddr, rData, none, err
	}

	var rType dnsmessage.Type
	var rBody dnsmessage.ResourceBody

	switch req.Type {
	case "A":
		rType = dnsmessage.TypeA
		ip := net.ParseIP(req.Data)
		if ip == nil {
			return req.Host, req.Type, udpAddr, rData, none, ErrIPInvalid
		}
		udpAddr = ip
		rBody = &dnsmessage.AResource{A: [4]byte{ip[12], ip[13], ip[14], ip[15]}}
	case "NS":
		rType = dnsmessage.TypeNS
		ns, err := dnsmessage.NewName(req.Data)
		if err != nil {
			return req.Host, req.Type, udpAddr, rData, none, err
		}
		rData = req.Data
		rBody = &dnsmessage.NSResource{NS: ns}
	case "CNAME":
		rType = dnsmessage.TypeCNAME
		cname, err := dnsmessage.NewName(req.Data)
		if err != nil {
			return req.Host, req.Type, udpAddr, rData, none, err
		}
		rData = req.Data
		rBody = &dnsmessage.CNAMEResource{CNAME: cname}
	case "SOA":
		rType = dnsmessage.TypeSOA
		soa := req.SOA
		soaNS, err := dnsmessage.NewName(soa.NS)
		if err != nil {
			return req.Host, req.Type, udpAddr, rData, none, err
		}
		soaMBox, err := dnsmessage.NewName(soa.MBox)
		if err != nil {
			return req.Host, req.Type, udpAddr, rData, none, err
		}
		rData = soa.MBox
		rBody = &dnsmessage.SOAResource{NS: soaNS, MBox: soaMBox, Serial: soa.Serial, Refresh: soa.Refresh, Retry: soa.Retry, Expire: soa.Expire}
	case "PTR":
		rType = dnsmessage.TypePTR
		ptr, err := dnsmessage.NewName(req.Data)
		if err != nil {
			return req.Host, req.Type, udpAddr, rData, none, err
		}
		rData = req.Data
		rBody = &dnsmessage.PTRResource{PTR: ptr}
	case "MX":
		rType = dnsmessage.TypeMX
		mxName, err := dnsmessage.NewName(req.MX.MX)
		if err != nil {
			return req.Host, req.Type, udpAddr, rData, none, err
		}
		rData = req.MX.MX
		rBody = &dnsmessage.MXResource{Pref: req.MX.Pref, MX: mxName}
	case "AAAA":
		rType = dnsmessage.TypeAAAA
		ip := net.ParseIP(req.Data)
		if ip == nil {
			return req.Host, req.Type, udpAddr, rData, none, ErrIPInvalid
		}
		var ipV6 [16]byte
		copy(ipV6[:], ip)
		udpAddr = ip
		rBody = &dnsmessage.AAAAResource{AAAA: ipV6}
	case "SRV":
		rType = dnsmessage.TypeSRV
		srv := req.SRV
		srvTarget, err := dnsmessage.NewName(srv.Target)
		if err != nil {
			return req.Host, req.Type, udpAddr, rData, none, err
		}
		rData = srv.Target
		rBody = &dnsmessage.SRVResource{Priority: srv.Priority, Weight: srv.Weight, Port: srv.Port, Target: srvTarget}
	case "TXT":
		fallthrough
	case "OPT":
		fallthrough
	default:
		return req.Host, req.Type, udpAddr, rData, none, ErrTypeNotSupport
	}

	return req.Host, req.Type, udpAddr, rData, dnsmessage.Resource{
		Header: dnsmessage.ResourceHeader{
			Name:  rName,
			Type:  rType,
			Class: dnsmessage.ClassINET,
			TTL:   req.TTL,
		},
		Body: rBody,
	}, nil
}

func ToResource(req model.Request) (dnsmessage.Resource, error) {
	rName, err := dnsmessage.NewName(req.Host)
	none := dnsmessage.Resource{}
	if err != nil {
		return none, err
	}

	var rType dnsmessage.Type
	var rBody dnsmessage.ResourceBody

	switch req.Type {
	case "A":
		rType = dnsmessage.TypeA
		ip := net.ParseIP(req.Data)
		if ip == nil {
			return none, ErrIPInvalid
		}
		rBody = &dnsmessage.AResource{A: [4]byte{ip[12], ip[13], ip[14], ip[15]}}
	case "NS":
		rType = dnsmessage.TypeNS
		ns, err := dnsmessage.NewName(req.Data)
		if err != nil {
			return none, err
		}
		rBody = &dnsmessage.NSResource{NS: ns}
	case "CNAME":
		rType = dnsmessage.TypeCNAME
		cname, err := dnsmessage.NewName(req.Data)
		if err != nil {
			return none, err
		}
		rBody = &dnsmessage.CNAMEResource{CNAME: cname}
	case "SOA":
		rType = dnsmessage.TypeSOA
		soa := req.SOA
		soaNS, err := dnsmessage.NewName(soa.NS)
		if err != nil {
			return none, err
		}
		soaMBox, err := dnsmessage.NewName(soa.MBox)
		if err != nil {
			return none, err
		}
		rBody = &dnsmessage.SOAResource{NS: soaNS, MBox: soaMBox, Serial: soa.Serial, Refresh: soa.Refresh, Retry: soa.Retry, Expire: soa.Expire}
	case "PTR":
		rType = dnsmessage.TypePTR
		ptr, err := dnsmessage.NewName(req.Data)
		if err != nil {
			return none, err
		}
		rBody = &dnsmessage.PTRResource{PTR: ptr}
	case "MX":
		rType = dnsmessage.TypeMX
		mxName, err := dnsmessage.NewName(req.MX.MX)
		if err != nil {
			return none, err
		}
		rBody = &dnsmessage.MXResource{Pref: req.MX.Pref, MX: mxName}
	case "AAAA":
		rType = dnsmessage.TypeAAAA
		ip := net.ParseIP(req.Data)
		if ip == nil {
			return none, ErrIPInvalid
		}
		var ipV6 [16]byte
		copy(ipV6[:], ip)
		rBody = &dnsmessage.AAAAResource{AAAA: ipV6}
	case "SRV":
		rType = dnsmessage.TypeSRV
		srv := req.SRV
		srvTarget, err := dnsmessage.NewName(srv.Target)
		if err != nil {
			return none, err
		}
		rBody = &dnsmessage.SRVResource{Priority: srv.Priority, Weight: srv.Weight, Port: srv.Port, Target: srvTarget}
	case "TXT":
		fallthrough
	case "OPT":
		fallthrough
	default:
		return none, ErrTypeNotSupport
	}

	return dnsmessage.Resource{
		Header: dnsmessage.ResourceHeader{
			Name:  rName,
			Type:  rType,
			Class: dnsmessage.ClassINET,
			TTL:   req.TTL,
		},
		Body: rBody,
	}, nil
}

func ToRType(sType string) dnsmessage.Type {
	switch sType {
	case "A":
		return dnsmessage.TypeA
	case "NS":
		return dnsmessage.TypeNS
	case "CNAME":
		return dnsmessage.TypeCNAME
	case "SOA":
		return dnsmessage.TypeSOA
	case "PTR":
		return dnsmessage.TypePTR
	case "MX":
		return dnsmessage.TypeMX
	case "AAAA":
		return dnsmessage.TypeAAAA
	case "SRV":
		return dnsmessage.TypeSRV
	case "TXT":
		return dnsmessage.TypeTXT
	case "OPT":
		return dnsmessage.TypeOPT
	default:
		return 0
	}
}

func ToResourceHeader(name string, sType string) (h dnsmessage.ResourceHeader, err error) {
	h.Name, err = dnsmessage.NewName(name)
	if err != nil {
		return
	}
	h.Type = ToRType(sType)
	return
}
func ConvertKeyToId(key string) string {
	return strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(key, ".", "-"), " ", "-"))
}
